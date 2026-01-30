package sftp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gossh/internal/model"
	gossh "gossh/internal/ssh"
)

// Client wraps an SFTP client
type Client struct {
	conn       model.Connection
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// NewClient creates a new SFTP client for a connection
func NewClient(conn model.Connection) *Client {
	return &Client{conn: conn}
}

// Connect establishes the SFTP connection
func (c *Client) Connect() error {
	// Build SSH auth methods
	authMethods, err := gossh.BuildAuthMethods(c.conn)
	if err != nil {
		return fmt.Errorf("failed to build auth methods: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            c.conn.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", c.conn.Host, c.conn.Port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}

	c.sshClient = sshClient
	c.sftpClient = sftpClient
	return nil
}

// Close closes the SFTP connection
func (c *Client) Close() error {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		c.sshClient.Close()
	}
	return nil
}

// Upload uploads a local file to the remote server
func (c *Client) Upload(localPath, remotePath string) error {
	// Expand local path
	localPath = expandPath(localPath)

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get local file info for permissions
	localInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	// Create remote file
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// Copy content
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Set permissions
	err = c.sftpClient.Chmod(remotePath, localInfo.Mode())
	if err != nil {
		// Non-fatal error
		fmt.Printf("Warning: failed to set permissions: %v\n", err)
	}

	return nil
}

// Download downloads a remote file to the local machine
func (c *Client) Download(remotePath, localPath string) error {
	// Expand local path
	localPath = expandPath(localPath)

	// Open remote file
	remoteFile, err := c.sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer remoteFile.Close()

	// Get remote file info
	remoteInfo, err := remoteFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat remote file: %w", err)
	}

	// Create local directory if needed
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// Copy content
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Set permissions
	err = os.Chmod(localPath, remoteInfo.Mode())
	if err != nil {
		fmt.Printf("Warning: failed to set permissions: %v\n", err)
	}

	return nil
}

// List lists files in a remote directory
func (c *Client) List(remotePath string) ([]FileInfo, error) {
	files, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	result := make([]FileInfo, len(files))
	for i, f := range files {
		result[i] = FileInfo{
			Name:    f.Name(),
			Size:    f.Size(),
			Mode:    f.Mode(),
			ModTime: f.ModTime(),
			IsDir:   f.IsDir(),
		}
	}

	return result, nil
}

// Mkdir creates a remote directory
func (c *Client) Mkdir(remotePath string) error {
	return c.sftpClient.MkdirAll(remotePath)
}

// Remove removes a remote file
func (c *Client) Remove(remotePath string) error {
	return c.sftpClient.Remove(remotePath)
}

// RemoveAll removes a remote directory and all its contents
func (c *Client) RemoveAll(remotePath string) error {
	return c.removeRecursive(remotePath)
}

func (c *Client) removeRecursive(path string) error {
	info, err := c.sftpClient.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		files, err := c.sftpClient.ReadDir(path)
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := c.removeRecursive(filepath.Join(path, f.Name())); err != nil {
				return err
			}
		}
		return c.sftpClient.RemoveDirectory(path)
	}

	return c.sftpClient.Remove(path)
}

// Stat returns file info for a remote path
func (c *Client) Stat(remotePath string) (*FileInfo, error) {
	info, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// Pwd returns the current working directory
func (c *Client) Pwd() (string, error) {
	return c.sftpClient.Getwd()
}

// FileInfo represents file information
type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime interface{}
	IsDir   bool
}

// String returns a formatted string representation
func (f FileInfo) String() string {
	typeChar := "-"
	if f.IsDir {
		typeChar = "d"
	}

	return fmt.Sprintf("%s%s %10d %s",
		typeChar,
		f.Mode.Perm().String()[1:],
		f.Size,
		f.Name,
	)
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
