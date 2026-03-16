package ssh

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"gossh/internal/model"
)

// Terminal handles interactive SSH terminal sessions
type Terminal struct {
	conn            model.Connection
	client          *Client
	startupTimeout  time.Duration
	hostKeyCallback ssh.HostKeyCallback
}

// NewTerminal creates a new terminal for a connection
func NewTerminal(conn model.Connection) *Terminal {
	return &Terminal{
		conn:           conn,
		client:         NewClient(conn),
		startupTimeout: 5 * time.Second,
	}
}

// SetHostKeyCallback sets the host key callback for verification
func (t *Terminal) SetHostKeyCallback(callback ssh.HostKeyCallback) {
	t.hostKeyCallback = callback
	t.client.SetHostKeyCallback(callback)
}

// SetStartupTimeout sets the timeout for startup command execution
func (t *Terminal) SetStartupTimeout(timeout time.Duration) {
	t.startupTimeout = timeout
}

// Run starts an interactive terminal session
func (t *Terminal) Run() error {
	// Connect to SSH server
	if err := t.client.Connect(); err != nil {
		return err
	}
	defer t.client.Close()

	// Create session
	session, err := t.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set up terminal
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal")
	}

	// Get terminal size (use defaults if unavailable)
	width, height := 80, 24
	if w, h, err := term.GetSize(fd); err == nil {
		width, height = w, h
	}

	// Request PTY
	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}
	if err := session.RequestPty(termType, height, width); err != nil {
		return fmt.Errorf("failed to request pty: %w", err)
	}

	// Forward locale environment variables for proper CJK wide-char display
	forwardLocaleEnv(session)

	// Set raw mode
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	// Connect stdin/stdout/stderr
	// When a startup command is configured, use io.Pipe as intermediary
	// so we can inject the command before forwarding os.Stdin.
	// Note: session.StdinPipe() cannot be used after session.Stdin is set
	// or after Shell() is called, so we must use this pipe approach.
	if t.conn.StartupCommand != "" {
		pr, pw := io.Pipe()
		session.SetStdin(pr)
		go func() {
			// Wait for shell to initialize
			time.Sleep(500 * time.Millisecond)
			cmd := strings.TrimSpace(t.conn.StartupCommand)
			if cmd != "" {
				_, _ = pw.Write([]byte(cmd + "\n"))
			}
			// Forward remaining stdin from user
			_, _ = io.Copy(pw, os.Stdin)
			_ = pw.Close()
		}()
	} else {
		session.SetStdin(os.Stdin)
	}
	session.SetStdout(os.Stdout)
	session.SetStderr(os.Stderr)

	// Handle window resize (platform-specific)
	cleanup := setupWindowResize(session, fd)
	defer cleanup()

	// Start shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// Start keepalive to detect dead connections
	ka := NewKeepalive(t.client.Conn())
	ka.Start()
	defer ka.Stop()

	// Wait for session to end
	waitErr := session.Wait()

	// Ensure cursor moves to a new line after session ends
	_, _ = os.Stdout.Write([]byte("\r\n"))

	// If keepalive detected a dead connection, report that instead
	if deadErr := ka.DeadError(); deadErr != nil {
		return fmt.Errorf("connection lost: %w", deadErr)
	}
	return waitErr
}

// RunWithIO runs an interactive session with custom IO
func (t *Terminal) RunWithIO(stdin io.Reader, stdout, stderr io.Writer, width, height int) error {
	if err := t.client.Connect(); err != nil {
		return err
	}
	defer t.client.Close()

	session, err := t.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}
	if err := session.RequestPty(termType, height, width); err != nil {
		return fmt.Errorf("failed to request pty: %w", err)
	}

	// Forward locale environment variables for proper CJK wide-char display
	forwardLocaleEnv(session)

	// When a startup command is configured, use io.Pipe to inject the command
	if t.conn.StartupCommand != "" {
		pr, pw := io.Pipe()
		session.SetStdin(pr)
		go func() {
			time.Sleep(500 * time.Millisecond)
			cmd := strings.TrimSpace(t.conn.StartupCommand)
			if cmd != "" {
				_, _ = pw.Write([]byte(cmd + "\n"))
			}
			_, _ = io.Copy(pw, stdin)
			_ = pw.Close()
		}()
	} else {
		session.SetStdin(stdin)
	}
	session.SetStdout(stdout)
	session.SetStderr(stderr)

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// Start keepalive to detect dead connections
	ka := NewKeepalive(t.client.Conn())
	ka.Start()
	defer ka.Stop()

	waitErr := session.Wait()

	// Write newline to ensure clean output after session ends
	_, _ = stdout.Write([]byte("\r\n"))

	if deadErr := ka.DeadError(); deadErr != nil {
		return fmt.Errorf("connection lost: %w", deadErr)
	}
	return waitErr
}

// forwardLocaleEnv forwards locale environment variables to the remote session.
// This ensures proper CJK wide-character display by letting the remote shell
// know the client's locale. Errors are silently ignored since not all SSH
// servers accept environment variables (matching OpenSSH client behavior).
func forwardLocaleEnv(session *Session) {
	for _, name := range []string{"LANG", "LC_CTYPE", "LC_ALL"} {
		if value := os.Getenv(name); value != "" {
			_ = session.Setenv(name, value)
		}
	}
}

// Close closes the terminal connection
func (t *Terminal) Close() error {
	if t.client != nil {
		return t.client.Close()
	}
	return nil
}
