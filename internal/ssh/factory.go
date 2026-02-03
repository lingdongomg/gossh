package ssh

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
	"gossh/internal/model"
)

// ConnectOptions holds options for creating an SSH connection
type ConnectOptions struct {
	Host            string
	Port            int
	User            string
	AuthMethods     []ssh.AuthMethod
	Timeout         time.Duration
	HostKeyCallback ssh.HostKeyCallback
}

// DefaultConnectOptions returns default connection options
func DefaultConnectOptions() ConnectOptions {
	return ConnectOptions{
		Timeout:         defaultTimeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Will be replaced by HostKey verification
	}
}

// Connect creates an SSH client connection with the given options
func Connect(opts ConnectOptions) (*ssh.Client, error) {
	if opts.Timeout == 0 {
		opts.Timeout = defaultTimeout
	}
	if opts.HostKeyCallback == nil {
		opts.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	config := &ssh.ClientConfig{
		User:            opts.User,
		Auth:            opts.AuthMethods,
		HostKeyCallback: opts.HostKeyCallback,
		Timeout:         opts.Timeout,
	}

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	return client, nil
}

// ConnectWithConnection creates an SSH connection using a model.Connection
func ConnectWithConnection(conn model.Connection, hostKeyCallback ssh.HostKeyCallback) (*ssh.Client, error) {
	authMethods, err := BuildAuthMethods(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to build auth methods: %w", err)
	}

	opts := ConnectOptions{
		Host:            conn.Host,
		Port:            conn.Port,
		User:            conn.User,
		AuthMethods:     authMethods,
		Timeout:         defaultTimeout,
		HostKeyCallback: hostKeyCallback,
	}

	if opts.HostKeyCallback == nil {
		opts.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	return Connect(opts)
}

// QuickCheck performs a quick TCP connection check
func QuickCheck(host string, port int, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// FullCheck performs a complete SSH handshake check
func FullCheck(conn model.Connection, hostKeyCallback ssh.HostKeyCallback) error {
	client, err := ConnectWithConnection(conn, hostKeyCallback)
	if err != nil {
		return err
	}
	client.Close()
	return nil
}
