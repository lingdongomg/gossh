package ssh

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"

	"golang.org/x/term"
	"gossh/internal/model"
)

// Terminal handles interactive SSH terminal sessions
type Terminal struct {
	conn   model.Connection
	client *Client
}

// NewTerminal creates a new terminal for a connection
func NewTerminal(conn model.Connection) *Terminal {
	return &Terminal{
		conn:   conn,
		client: NewClient(conn),
	}
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
	defer term.Restore(fd, oldState)

	// Connect stdin/stdout/stderr
	session.SetStdin(os.Stdin)
	session.SetStdout(os.Stdout)
	session.SetStderr(os.Stderr)

	// Handle window resize (only on Unix-like systems)
	if runtime.GOOS != "windows" {
		sigwinch := make(chan os.Signal, 1)
		signal.Notify(sigwinch, os.Interrupt) // Fallback signal
		defer signal.Stop(sigwinch)

		go func() {
			for range sigwinch {
				if w, h, err := term.GetSize(fd); err == nil {
					session.WindowChange(h, w)
				}
			}
		}()
	}

	// Start shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// Wait for session to end
	return session.Wait()
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

	return session.Wait()
}

// Close closes the terminal connection
func (t *Terminal) Close() error {
	if t.client != nil {
		return t.client.Close()
	}
	return nil
}
