package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
	"gossh/internal/model"
)

// ForwardType represents the type of port forwarding
type ForwardType string

const (
	ForwardLocal  ForwardType = "local"  // -L local:remote
	ForwardRemote ForwardType = "remote" // -R remote:local
)

// PortForward represents a port forwarding configuration
type PortForward struct {
	Type       ForwardType
	LocalHost  string
	LocalPort  int
	RemoteHost string
	RemotePort int
}

// ParsePortForward parses a port forward string like "8080:localhost:80"
func ParsePortForward(fwdType ForwardType, spec string) (*PortForward, error) {
	var localHost, remoteHost string
	var localPort, remotePort int

	// Parse spec: [bind_address:]port:host:hostport
	n, err := fmt.Sscanf(spec, "%d:%s:%d", &localPort, &remoteHost, &remotePort)
	if n == 3 && err == nil {
		localHost = "localhost"
	} else {
		n, err = fmt.Sscanf(spec, "%s:%d:%s:%d", &localHost, &localPort, &remoteHost, &remotePort)
		if n != 4 || err != nil {
			return nil, fmt.Errorf("invalid forward spec: %s (expected [bind:]port:host:port)", spec)
		}
	}

	return &PortForward{
		Type:       fwdType,
		LocalHost:  localHost,
		LocalPort:  localPort,
		RemoteHost: remoteHost,
		RemotePort: remotePort,
	}, nil
}

// String returns a string representation
func (pf *PortForward) String() string {
	if pf.Type == ForwardLocal {
		return fmt.Sprintf("-L %s:%d:%s:%d", pf.LocalHost, pf.LocalPort, pf.RemoteHost, pf.RemotePort)
	}
	return fmt.Sprintf("-R %s:%d:%s:%d", pf.RemoteHost, pf.RemotePort, pf.LocalHost, pf.LocalPort)
}

// Forwarder manages port forwarding
type Forwarder struct {
	conn            model.Connection
	client          *ssh.Client
	forwards        []*PortForward
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.Mutex
	running         bool
	hostKeyCallback ssh.HostKeyCallback
}

// NewForwarder creates a new port forwarder
func NewForwarder(conn model.Connection) *Forwarder {
	ctx, cancel := context.WithCancel(context.Background())
	return &Forwarder{
		conn:     conn,
		forwards: make([]*PortForward, 0),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// SetHostKeyCallback sets the host key callback for verification
func (f *Forwarder) SetHostKeyCallback(callback ssh.HostKeyCallback) {
	f.hostKeyCallback = callback
}

// AddForward adds a port forward rule
func (f *Forwarder) AddForward(pf *PortForward) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.forwards = append(f.forwards, pf)
}

// Connect establishes the SSH connection using the factory function
func (f *Forwarder) Connect() error {
	client, err := ConnectWithConnection(f.conn, f.hostKeyCallback)
	if err != nil {
		return err
	}
	f.client = client
	return nil
}

// Start starts all port forwards
func (f *Forwarder) Start() error {
	f.mu.Lock()
	if f.running {
		f.mu.Unlock()
		return fmt.Errorf("forwarder already running")
	}
	f.running = true
	f.mu.Unlock()

	for _, pf := range f.forwards {
		switch pf.Type {
		case ForwardLocal:
			if err := f.startLocalForward(pf); err != nil {
				return err
			}
		case ForwardRemote:
			if err := f.startRemoteForward(pf); err != nil {
				return err
			}
		}
	}

	return nil
}

// startLocalForward starts a local port forward (-L)
func (f *Forwarder) startLocalForward(pf *PortForward) error {
	localAddr := fmt.Sprintf("%s:%d", pf.LocalHost, pf.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		defer listener.Close()

		for {
			select {
			case <-f.ctx.Done():
				return
			default:
			}

			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-f.ctx.Done():
					return
				default:
					continue
				}
			}

			remoteAddr := fmt.Sprintf("%s:%d", pf.RemoteHost, pf.RemotePort)
			f.wg.Add(1)
			go func(localConn net.Conn) {
				defer f.wg.Done()
				defer localConn.Close()

				remoteConn, err := f.client.Dial("tcp", remoteAddr)
				if err != nil {
					return
				}
				defer remoteConn.Close()

				f.copyBidirectional(localConn, remoteConn)
			}(conn)
		}
	}()

	fmt.Printf("Local forward: %s -> %s:%d\n", localAddr, pf.RemoteHost, pf.RemotePort)
	return nil
}

// startRemoteForward starts a remote port forward (-R)
func (f *Forwarder) startRemoteForward(pf *PortForward) error {
	remoteAddr := fmt.Sprintf("%s:%d", pf.RemoteHost, pf.RemotePort)
	listener, err := f.client.Listen("tcp", remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on remote %s: %w", remoteAddr, err)
	}

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		defer listener.Close()

		for {
			select {
			case <-f.ctx.Done():
				return
			default:
			}

			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-f.ctx.Done():
					return
				default:
					continue
				}
			}

			localAddr := net.JoinHostPort(pf.LocalHost, fmt.Sprintf("%d", pf.LocalPort))
			f.wg.Add(1)
			go func(remoteConn net.Conn) {
				defer f.wg.Done()
				defer remoteConn.Close()

				localConn, err := net.Dial("tcp", localAddr)
				if err != nil {
					return
				}
				defer localConn.Close()

				f.copyBidirectional(remoteConn, localConn)
			}(conn)
		}
	}()

	fmt.Printf("Remote forward: %s -> %s:%d\n", remoteAddr, pf.LocalHost, pf.LocalPort)
	return nil
}

// copyBidirectional copies data between two connections
func (f *Forwarder) copyBidirectional(conn1, conn2 net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(conn1, conn2)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(conn2, conn1)
	}()

	wg.Wait()
}

// Stop stops all port forwards
func (f *Forwarder) Stop() {
	f.cancel()
	if f.client != nil {
		f.client.Close()
	}
	f.wg.Wait()

	f.mu.Lock()
	f.running = false
	f.mu.Unlock()
}

// Wait waits for all forwards to complete
func (f *Forwarder) Wait() {
	f.wg.Wait()
}
