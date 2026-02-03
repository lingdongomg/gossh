package ssh

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
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

	// Set raw mode
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	// Connect stdin/stdout/stderr
	session.SetStdin(os.Stdin)
	session.SetStdout(os.Stdout)
	session.SetStderr(os.Stderr)

	// Handle window resize (only on Unix-like systems)
	if runtime.GOOS != "windows" {
		sigwinch := make(chan os.Signal, 1)
		// Use SIGWINCH for window resize signal
		signal.Notify(sigwinch, syscall.SIGWINCH)
		defer signal.Stop(sigwinch)

		go func() {
			for range sigwinch {
				if w, h, err := term.GetSize(fd); err == nil {
					_ = session.WindowChange(h, w)
				}
			}
		}()
	}

	// Start shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// Execute startup command if configured
	if t.conn.StartupCommand != "" {
		go t.executeStartupCommand(session)
	}

	// Wait for session to end
	return session.Wait()
}

// executeStartupCommand sends the startup command to the shell
func (t *Terminal) executeStartupCommand(session *Session) {
	// Wait a moment for the shell to initialize
	time.Sleep(500 * time.Millisecond)

	// Send the command followed by newline
	cmd := strings.TrimSpace(t.conn.StartupCommand)
	if cmd != "" {
		// Write the command to stdin via the session
		// Note: This writes through the PTY which simulates user input
		stdinPipe, err := session.StdinPipe()
		if err != nil {
			return
		}
		_, _ = stdinPipe.Write([]byte(cmd + "\n"))
	}
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

	session.SetStdin(stdin)
	session.SetStdout(stdout)
	session.SetStderr(stderr)

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// Execute startup command if configured
	if t.conn.StartupCommand != "" {
		go t.executeStartupCommand(session)
	}

	return session.Wait()
}

// Close closes the terminal connection
func (t *Terminal) Close() error {
	if t.client != nil {
		return t.client.Close()
	}
	return nil
}
