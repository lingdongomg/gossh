package ssh

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
	"gossh/internal/model"
)

const (
	defaultTimeout = 30 * time.Second
)

// Client wraps an SSH client connection
type Client struct {
	conn   model.Connection
	client *ssh.Client
}

// NewClient creates a new SSH client for a connection
func NewClient(conn model.Connection) *Client {
	return &Client{conn: conn}
}

// Connect establishes the SSH connection
func (c *Client) Connect() error {
	authMethods, err := BuildAuthMethods(c.conn)
	if err != nil {
		return fmt.Errorf("failed to build auth methods: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            c.conn.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: implement known_hosts
		Timeout:         defaultTimeout,
	}

	addr := fmt.Sprintf("%s:%d", c.conn.Host, c.conn.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	c.client = client
	return nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// NewSession creates a new SSH session
func (c *Client) NewSession() (*Session, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}

	return &Session{session: session}, nil
}

// LocalAddr returns the local address of the connection
func (c *Client) LocalAddr() net.Addr {
	if c.client != nil {
		return c.client.LocalAddr()
	}
	return nil
}

// RemoteAddr returns the remote address of the connection
func (c *Client) RemoteAddr() net.Addr {
	if c.client != nil {
		return c.client.RemoteAddr()
	}
	return nil
}
