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
	conn            model.Connection
	client          *ssh.Client
	hostKeyCallback ssh.HostKeyCallback
}

// NewClient creates a new SSH client for a connection
func NewClient(conn model.Connection) *Client {
	return &Client{conn: conn}
}

// SetHostKeyCallback sets the host key callback for verification
func (c *Client) SetHostKeyCallback(callback ssh.HostKeyCallback) {
	c.hostKeyCallback = callback
}

// Connect establishes the SSH connection using the factory function
func (c *Client) Connect() error {
	client, err := ConnectWithConnection(c.conn, c.hostKeyCallback)
	if err != nil {
		return err
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
