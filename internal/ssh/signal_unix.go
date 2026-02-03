//go:build !windows

package ssh

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// setupWindowResize sets up window resize signal handling on Unix systems
func setupWindowResize(session *Session, fd int) func() {
	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)

	go func() {
		for range sigwinch {
			if w, h, err := term.GetSize(fd); err == nil {
				_ = session.WindowChange(h, w)
			}
		}
	}()

	return func() {
		signal.Stop(sigwinch)
		close(sigwinch)
	}
}
