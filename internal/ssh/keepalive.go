package ssh

import (
	"fmt"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

const (
	keepaliveInterval  = 10 * time.Second
	keepaliveMaxFailed = 3
	keepaliveRequest   = "keepalive@openssh.com"
)

// Keepalive sends periodic keepalive requests over an SSH connection
// and closes the connection when it detects it is dead.
type Keepalive struct {
	conn     gossh.Conn
	stop     chan struct{}
	once     sync.Once
	deadErr  error
	deadLock sync.Mutex
}

// NewKeepalive creates a new Keepalive for the given SSH connection.
func NewKeepalive(conn gossh.Conn) *Keepalive {
	return &Keepalive{
		conn: conn,
		stop: make(chan struct{}),
	}
}

// Start begins sending keepalive requests in a background goroutine.
func (k *Keepalive) Start() {
	go k.loop()
}

// Stop stops the keepalive loop. Safe to call multiple times.
func (k *Keepalive) Stop() {
	k.once.Do(func() {
		close(k.stop)
	})
}

// DeadError returns a non-nil error if the connection was detected as dead.
func (k *Keepalive) DeadError() error {
	k.deadLock.Lock()
	defer k.deadLock.Unlock()
	return k.deadErr
}

func (k *Keepalive) loop() {
	ticker := time.NewTicker(keepaliveInterval)
	defer ticker.Stop()

	failures := 0
	for {
		select {
		case <-k.stop:
			return
		case <-ticker.C:
			_, _, err := k.conn.SendRequest(keepaliveRequest, true, nil)
			if err != nil {
				failures++
				if failures >= keepaliveMaxFailed {
					k.deadLock.Lock()
					k.deadErr = fmt.Errorf("ssh keepalive failed %d times: %w", failures, err)
					k.deadLock.Unlock()
					// Close the connection to unblock session.Wait()
					_ = k.conn.Close()
					return
				}
			} else {
				failures = 0
			}
		}
	}
}
