package sshconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gossh/internal/model"
)

// Parser parses OpenSSH config files
type Parser struct{}

// NewParser creates a new SSH config parser
func NewParser() *Parser {
	return &Parser{}
}

// hostEntry represents a parsed Host block
type hostEntry struct {
	patterns     []string
	hostName     string
	user         string
	port         int
	identityFile string
}

// ParseFile parses an SSH config file and returns connections
func (p *Parser) ParseFile(path string) ([]model.Connection, error) {
	// Expand path
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return p.parse(file)
}

// ParseDefault parses the default SSH config file (~/.ssh/config)
func (p *Parser) ParseDefault() ([]model.Connection, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return p.ParseFile(filepath.Join(home, ".ssh", "config"))
}

// parse parses an SSH config from a file
func (p *Parser) parse(file *os.File) ([]model.Connection, error) {
	var entries []*hostEntry
	var current *hostEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key-value pair
		key, value := p.parseLine(line)
		if key == "" {
			continue
		}

		key = strings.ToLower(key)

		switch key {
		case "host":
			// Start new host block
			if current != nil {
				entries = append(entries, current)
			}
			current = &hostEntry{
				patterns: strings.Fields(value),
				port:     22, // Default port
			}
		case "hostname":
			if current != nil {
				current.hostName = value
			}
		case "user":
			if current != nil {
				current.user = value
			}
		case "port":
			if current != nil {
				if port, err := strconv.Atoi(value); err == nil {
					current.port = port
				}
			}
		case "identityfile":
			if current != nil {
				// Expand ~ in path
				if strings.HasPrefix(value, "~/") {
					home, err := os.UserHomeDir()
					if err == nil {
						value = filepath.Join(home, value[2:])
					}
				}
				current.identityFile = value
			}
		}
	}

	// Don't forget the last entry
	if current != nil {
		entries = append(entries, current)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert entries to connections
	return p.entriesToConnections(entries), nil
}

// parseLine parses a line into key-value pair
func (p *Parser) parseLine(line string) (string, string) {
	// Handle both "key value" and "key=value" formats
	line = strings.TrimSpace(line)

	// Try space separator first
	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	// Try = separator
	parts = strings.SplitN(line, "=", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	return "", ""
}

// entriesToConnections converts host entries to model.Connection
func (p *Parser) entriesToConnections(entries []*hostEntry) []model.Connection {
	var connections []model.Connection
	wildcardPattern := regexp.MustCompile(`[*?]`)

	for _, entry := range entries {
		// Skip entries with wildcards
		hasWildcard := false
		for _, pattern := range entry.patterns {
			if wildcardPattern.MatchString(pattern) {
				hasWildcard = true
				break
			}
		}
		if hasWildcard {
			continue
		}

		// Create a connection for each pattern (usually just one)
		for _, pattern := range entry.patterns {
			conn := model.Connection{
				ID:      uuid.New().String(),
				Name:    pattern,
				Host:    entry.hostName,
				Port:    entry.port,
				User:    entry.user,
				Group:   "Imported",
			}

			// If no hostname specified, use the pattern as hostname
			if conn.Host == "" {
				conn.Host = pattern
			}

			// Set authentication type based on identity file
			if entry.identityFile != "" {
				conn.AuthType = model.AuthKey
				conn.KeyPath = entry.identityFile
			} else {
				conn.AuthType = model.AuthPassword
			}

			// Skip invalid connections
			if conn.Host == "" || conn.User == "" {
				continue
			}

			connections = append(connections, conn)
		}
	}

	return connections
}

// Merge merges imported connections with existing ones
// Returns the new connections and skipped count
func Merge(existing, imported []model.Connection) (newConns []model.Connection, skipped int) {
	existingNames := make(map[string]bool)
	for _, c := range existing {
		existingNames[strings.ToLower(c.Name)] = true
	}

	for _, c := range imported {
		if existingNames[strings.ToLower(c.Name)] {
			skipped++
			continue
		}
		newConns = append(newConns, c)
	}

	return newConns, skipped
}
