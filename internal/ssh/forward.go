package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
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
// For -L: spec is <local-port>:<remote-host>:<remote-port>
// For -R: spec is <remote-port>:<local-host>:<local-port>
func ParsePortForward(fwdType ForwardType, spec string) (*PortForward, error) {
	// Parse spec: [bind_address:]port:host:hostport
	parts := strings.Split(spec, ":")

	var bindHost string
	var port1, port2 int

	switch len(parts) {
	case 3:
		// port:host:hostport
		p1, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid forward spec: %s (invalid port: %s)", spec, parts[0])
		}
		p2, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid forward spec: %s (invalid port: %s)", spec, parts[2])
		}
		bindHost = parts[1]
		port1 = p1
		port2 = p2
	case 4:
		// bind_address:port:host:hostport
		p1, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid forward spec: %s (invalid port: %s)", spec, parts[1])
		}
		p2, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, fmt.Errorf("invalid forward spec: %s (invalid port: %s)", spec, parts[3])
		}
		bindHost = parts[2]
		port1 = p1
		port2 = p2
	default:
		return nil, fmt.Errorf("invalid forward spec: %s (expected [bind:]port:host:port)", spec)
	}

	pf := &PortForward{Type: fwdType}
	if fwdType == ForwardLocal {
		// -L local_port:remote_host:remote_port
		pf.LocalHost = "localhost"
		pf.LocalPort = port1
		pf.RemoteHost = bindHost
		pf.RemotePort = port2
		if len(parts) == 4 {
			pf.LocalHost = parts[0]
		}
	} else {
		// -R remote_port:local_host:local_port
		pf.RemoteHost = "localhost"
		pf.RemotePort = port1
		pf.LocalHost = bindHost
		pf.LocalPort = port2
		if len(parts) == 4 {
			pf.RemoteHost = parts[0]
		}
	}

	return pf, nil
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

	fmt.Printf("Local forward: %s -> [%s] -> %s:%d\n", localAddr, f.conn.Host, pf.RemoteHost, pf.RemotePort)
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

	fmt.Printf("Remote forward: [%s] %s -> %s:%d\n", f.conn.Host, remoteAddr, pf.LocalHost, pf.LocalPort)
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
