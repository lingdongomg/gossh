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
	conn            model.Connection
	sshClient       *ssh.Client
	sftpClient      *sftp.Client
	currentDir      string // Track current working directory
	hostKeyCallback ssh.HostKeyCallback
}

// NewClient creates a new SFTP client for a connection
func NewClient(conn model.Connection) *Client {
	return &Client{conn: conn}
}

// SetHostKeyCallback sets the host key callback for verification
func (c *Client) SetHostKeyCallback(callback ssh.HostKeyCallback) {
	c.hostKeyCallback = callback
}

// Connect establishes the SFTP connection using the factory function
func (c *Client) Connect() error {
	sshClient, err := gossh.ConnectWithConnection(c.conn, c.hostKeyCallback)
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}

	c.sshClient = sshClient
	c.sftpClient = sftpClient

	// Initialize current directory
	c.currentDir, _ = c.sftpClient.Getwd()

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
	if c.currentDir != "" {
		return c.currentDir, nil
	}
	return c.sftpClient.Getwd()
}

// Cd changes the current working directory
func (c *Client) Cd(path string) error {
	// Resolve path
	newPath := c.resolvePath(path)

	// Verify it exists and is a directory
	info, err := c.sftpClient.Stat(newPath)
	if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", newPath)
	}

	c.currentDir = newPath
	return nil
}

// CurrentDir returns the current working directory
func (c *Client) CurrentDir() string {
	return c.currentDir
}

// resolvePath resolves a path relative to the current directory
func (c *Client) resolvePath(path string) string {
	if path == "" {
		return c.currentDir
	}
	// Absolute path
	if strings.HasPrefix(path, "/") {
		return filepath.Clean(path)
	}
	// Handle ~ for home directory
	if path == "~" || strings.HasPrefix(path, "~/") {
		// On remote, ~ typically resolves to user's home
		home, err := c.sftpClient.Getwd()
		if err == nil && path == "~" {
			return home
		}
		if strings.HasPrefix(path, "~/") {
			return filepath.Join(home, path[2:])
		}
	}
	// Relative path
	return filepath.Clean(filepath.Join(c.currentDir, path))
}

// ProgressCallback is called during file transfer to report progress
type ProgressCallback func(transferred, total int64)

// UploadWithProgress uploads a local file to the remote server with progress reporting
func (c *Client) UploadWithProgress(localPath, remotePath string, progress ProgressCallback) error {
	// Expand local path
	localPath = expandPath(localPath)
	// Resolve remote path
	remotePath = c.resolvePath(remotePath)

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get local file info for permissions and size
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

	// Copy content with progress
	var transferred int64
	total := localInfo.Size()
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := localFile.Read(buf)
		if n > 0 {
			written, writeErr := remoteFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write: %w", writeErr)
			}
			transferred += int64(written)
			if progress != nil {
				progress(transferred, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}
	}

	// Set permissions
	err = c.sftpClient.Chmod(remotePath, localInfo.Mode())
	if err != nil {
		fmt.Printf("Warning: failed to set permissions: %v\n", err)
	}

	return nil
}

// DownloadWithProgress downloads a remote file to the local machine with progress reporting
func (c *Client) DownloadWithProgress(remotePath, localPath string, progress ProgressCallback) error {
	// Expand local path
	localPath = expandPath(localPath)
	// Resolve remote path
	remotePath = c.resolvePath(remotePath)

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

	// Copy content with progress
	var transferred int64
	total := remoteInfo.Size()
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := remoteFile.Read(buf)
		if n > 0 {
			written, writeErr := localFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write: %w", writeErr)
			}
			transferred += int64(written)
			if progress != nil {
				progress(transferred, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}
	}

	// Set permissions
	err = os.Chmod(localPath, remoteInfo.Mode())
	if err != nil {
		fmt.Printf("Warning: failed to set permissions: %v\n", err)
	}

	return nil
}

// ListCurrentDir lists files in the current working directory
func (c *Client) ListCurrentDir() ([]FileInfo, error) {
	return c.List(c.currentDir)
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
