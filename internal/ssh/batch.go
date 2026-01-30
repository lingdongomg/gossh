package ssh

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"gossh/internal/model"
)

// BatchResult represents the result of executing a command on one host
type BatchResult struct {
	Connection model.Connection
	Output     string
	Error      error
	Duration   time.Duration
	ExitCode   int
}

// BatchExecutor executes commands on multiple hosts
type BatchExecutor struct {
	connections []model.Connection
	timeout     time.Duration
	parallel    int
}

// NewBatchExecutor creates a new batch executor
func NewBatchExecutor(connections []model.Connection) *BatchExecutor {
	return &BatchExecutor{
		connections: connections,
		timeout:     30 * time.Second,
		parallel:    10, // Default parallel connections
	}
}

// SetTimeout sets the command timeout
func (b *BatchExecutor) SetTimeout(timeout time.Duration) {
	b.timeout = timeout
}

// SetParallel sets the max parallel connections
func (b *BatchExecutor) SetParallel(n int) {
	if n > 0 {
		b.parallel = n
	}
}

// Execute executes a command on all connections
func (b *BatchExecutor) Execute(ctx context.Context, command string) []BatchResult {
	results := make([]BatchResult, len(b.connections))
	var wg sync.WaitGroup
	sem := make(chan struct{}, b.parallel)

	for i, conn := range b.connections {
		wg.Add(1)
		go func(idx int, c model.Connection) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[idx] = BatchResult{
					Connection: c,
					Error:      ctx.Err(),
				}
				return
			}

			// Execute command
			results[idx] = b.executeOne(ctx, c, command)
		}(i, conn)
	}

	wg.Wait()
	return results
}

// executeOne executes a command on a single connection
func (b *BatchExecutor) executeOne(ctx context.Context, conn model.Connection, command string) BatchResult {
	start := time.Now()
	result := BatchResult{
		Connection: conn,
	}

	// Build auth methods
	authMethods, err := BuildAuthMethods(conn)
	if err != nil {
		result.Error = fmt.Errorf("auth error: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Create SSH config
	config := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         b.timeout,
	}

	// Connect
	addr := fmt.Sprintf("%s:%d", conn.Host, conn.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		result.Error = fmt.Errorf("connection error: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer client.Close()

	// Create session
	session, err := client.NewSession()
	if err != nil {
		result.Error = fmt.Errorf("session error: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer session.Close()

	// Set up output capture
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Create a channel to signal command completion
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	// Wait for command or context cancellation
	select {
	case err := <-done:
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				result.ExitCode = exitErr.ExitStatus()
			}
			result.Error = err
		}
	case <-ctx.Done():
		session.Signal(ssh.SIGTERM)
		result.Error = ctx.Err()
	}

	// Combine output
	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// FilterByGroup filters connections by group
func FilterByGroup(connections []model.Connection, group string) []model.Connection {
	if group == "" {
		return connections
	}

	var result []model.Connection
	for _, c := range connections {
		if c.Group == group {
			result = append(result, c)
		}
	}
	return result
}

// FilterByTags filters connections by tags (any match)
func FilterByTags(connections []model.Connection, tags []string) []model.Connection {
	if len(tags) == 0 {
		return connections
	}

	tagSet := make(map[string]bool)
	for _, t := range tags {
		tagSet[t] = true
	}

	var result []model.Connection
	for _, c := range connections {
		for _, t := range c.Tags {
			if tagSet[t] {
				result = append(result, c)
				break
			}
		}
	}
	return result
}

// FilterByNames filters connections by names
func FilterByNames(connections []model.Connection, names []string) []model.Connection {
	if len(names) == 0 {
		return connections
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	var result []model.Connection
	for _, c := range connections {
		if nameSet[c.Name] {
			result = append(result, c)
		}
	}
	return result
}

// PrintResults prints batch execution results
func PrintResults(results []BatchResult) {
	fmt.Println("\n" + string(make([]byte, 80)))
	fmt.Println("BATCH EXECUTION RESULTS")
	fmt.Println(string(make([]byte, 80)))

	successCount := 0
	failCount := 0

	for _, r := range results {
		status := "✓"
		if r.Error != nil {
			status = "✗"
			failCount++
		} else {
			successCount++
		}

		fmt.Printf("\n%s [%s] %s@%s:%d (%.2fs)\n",
			status, r.Connection.Name, r.Connection.User,
			r.Connection.Host, r.Connection.Port,
			r.Duration.Seconds())
		fmt.Println(string(make([]byte, 40)))

		if r.Error != nil {
			fmt.Printf("Error: %v\n", r.Error)
		}
		if r.Output != "" {
			fmt.Println(r.Output)
		}
	}

	fmt.Println("\n" + string(make([]byte, 80)))
	fmt.Printf("Summary: %d succeeded, %d failed, %d total\n",
		successCount, failCount, len(results))
}
