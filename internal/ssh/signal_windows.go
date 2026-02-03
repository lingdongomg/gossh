//go:build windows

package ssh

// setupWindowResize is a no-op on Windows as SIGWINCH is not available
func setupWindowResize(session *Session, fd int) func() {
	// Windows doesn't support SIGWINCH, return empty cleanup function
	return func() {}
}
