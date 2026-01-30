package ssh

import (
	"io"

	"golang.org/x/crypto/ssh"
)

// Session wraps an SSH session
type Session struct {
	session *ssh.Session
}

// Close closes the session
func (s *Session) Close() error {
	return s.session.Close()
}

// Shell starts an interactive shell
func (s *Session) Shell() error {
	return s.session.Shell()
}

// Wait waits for the session to finish
func (s *Session) Wait() error {
	return s.session.Wait()
}

// RequestPty requests a pseudo-terminal
func (s *Session) RequestPty(term string, height, width int) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	return s.session.RequestPty(term, height, width, modes)
}

// WindowChange sends a window change request
func (s *Session) WindowChange(height, width int) error {
	return s.session.WindowChange(height, width)
}

// StdinPipe returns a pipe for stdin
func (s *Session) StdinPipe() (io.WriteCloser, error) {
	return s.session.StdinPipe()
}

// StdoutPipe returns a pipe for stdout
func (s *Session) StdoutPipe() (io.Reader, error) {
	return s.session.StdoutPipe()
}

// StderrPipe returns a pipe for stderr
func (s *Session) StderrPipe() (io.Reader, error) {
	return s.session.StderrPipe()
}

// SetStdin sets the stdin
func (s *Session) SetStdin(r io.Reader) {
	s.session.Stdin = r
}

// SetStdout sets the stdout
func (s *Session) SetStdout(w io.Writer) {
	s.session.Stdout = w
}

// SetStderr sets the stderr
func (s *Session) SetStderr(w io.Writer) {
	s.session.Stderr = w
}
