package ssh

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"gossh/internal/config"
)

// HostKeyStatus represents the status of a host key verification
type HostKeyStatus int

const (
	HostKeyOK       HostKeyStatus = iota // Key matches known_hosts
	HostKeyNew                           // New host, not in known_hosts
	HostKeyChanged                       // Key has changed from known_hosts
)

// HostKeyResult contains the result of host key verification
type HostKeyResult struct {
	Status      HostKeyStatus
	Host        string
	Fingerprint string
	KeyType     string
	OldKey      string // Only set if HostKeyChanged
}

// HostKeyManager manages known hosts
type HostKeyManager struct {
	knownHosts map[string]string // host:port -> key fingerprint
	filePath   string
	mu         sync.RWMutex
}

// NewHostKeyManager creates a new host key manager
func NewHostKeyManager() (*HostKeyManager, error) {
	hkm := &HostKeyManager{
		knownHosts: make(map[string]string),
		filePath:   config.GetKnownHostsPath(),
	}

	if err := hkm.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return hkm, nil
}

// load reads the known_hosts file
func (h *HostKeyManager) load() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	file, err := os.Open(h.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: host keytype key [comment]
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		host := parts[0]
		keyType := parts[1]
		keyData := parts[2]

		// Store fingerprint
		fingerprint := h.computeFingerprint(keyType, keyData)
		h.knownHosts[host] = fingerprint
	}

	return scanner.Err()
}

// computeFingerprint computes SHA256 fingerprint from key data
func (h *HostKeyManager) computeFingerprint(keyType, keyData string) string {
	decoded, err := base64.StdEncoding.DecodeString(keyData)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(decoded)
	return fmt.Sprintf("%s SHA256:%s", keyType, base64.RawStdEncoding.EncodeToString(hash[:]))
}

// Save saves the known_hosts file
func (h *HostKeyManager) Save() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	file, err := os.OpenFile(h.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// We store fingerprints, but for compatibility we would need original keys
	// For now, this is a simplified implementation
	for host, fingerprint := range h.knownHosts {
		fmt.Fprintf(file, "# %s %s\n", host, fingerprint)
	}

	return nil
}

// AddHost adds a host key to known_hosts
func (h *HostKeyManager) AddHost(host string, port int, key ssh.PublicKey) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	hostKey := formatHostPort(host, port)
	fingerprint := FormatFingerprint(key)

	h.knownHosts[hostKey] = fingerprint

	// Append to file
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	file, err := os.OpenFile(h.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write in OpenSSH format
	keyBytes := base64.StdEncoding.EncodeToString(key.Marshal())
	_, err = fmt.Fprintf(file, "%s %s %s\n", hostKey, key.Type(), keyBytes)
	return err
}

// UpdateHost updates a host key in known_hosts
func (h *HostKeyManager) UpdateHost(host string, port int, key ssh.PublicKey) error {
	h.mu.Lock()
	hostKey := formatHostPort(host, port)
	h.knownHosts[hostKey] = FormatFingerprint(key)
	h.mu.Unlock()

	return h.rewriteFile()
}

// rewriteFile rewrites the known_hosts file with current entries
func (h *HostKeyManager) rewriteFile() error {
	// This is a simplified implementation
	// In production, you'd want to preserve comments and original key data
	return h.Save()
}

// CheckHostKey checks a host key against known_hosts
func (h *HostKeyManager) CheckHostKey(host string, port int, key ssh.PublicKey) *HostKeyResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	hostKey := formatHostPort(host, port)
	fingerprint := FormatFingerprint(key)

	storedFingerprint, exists := h.knownHosts[hostKey]

	result := &HostKeyResult{
		Host:        host,
		Fingerprint: fingerprint,
		KeyType:     key.Type(),
	}

	if !exists {
		result.Status = HostKeyNew
	} else if storedFingerprint == fingerprint {
		result.Status = HostKeyOK
	} else {
		result.Status = HostKeyChanged
		result.OldKey = storedFingerprint
	}

	return result
}

// formatHostPort formats host and port for known_hosts
func formatHostPort(host string, port int) string {
	if port == 22 {
		return host
	}
	return fmt.Sprintf("[%s]:%d", host, port)
}

// FormatFingerprint returns a human-readable fingerprint of the key
func FormatFingerprint(key ssh.PublicKey) string {
	hash := sha256.Sum256(key.Marshal())
	return fmt.Sprintf("%s SHA256:%s", key.Type(), base64.RawStdEncoding.EncodeToString(hash[:]))
}

// HostKeyCallback is the callback function type for host key verification
type HostKeyCallback func(result *HostKeyResult) (accept bool, update bool)

// CreateHostKeyCallback creates an SSH host key callback with the given handler
func CreateHostKeyCallback(hkm *HostKeyManager, handler HostKeyCallback) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Extract host and port from remote address
		host, portStr, err := net.SplitHostPort(remote.String())
		if err != nil {
			host = hostname
		}
		port := 22
		if portStr != "" {
			fmt.Sscanf(portStr, "%d", &port)
		}

		result := hkm.CheckHostKey(host, port, key)

		switch result.Status {
		case HostKeyOK:
			return nil
		case HostKeyNew, HostKeyChanged:
			if handler == nil {
				// No handler, reject by default for safety
				if result.Status == HostKeyNew {
					return fmt.Errorf("unknown host: %s", host)
				}
				return fmt.Errorf("host key changed for: %s", host)
			}

			accept, update := handler(result)
			if !accept {
				if result.Status == HostKeyNew {
					return fmt.Errorf("host key rejected for: %s", host)
				}
				return fmt.Errorf("host key change rejected for: %s", host)
			}

			// Save the new key
			if result.Status == HostKeyNew {
				hkm.AddHost(host, port, key)
			} else if update {
				hkm.UpdateHost(host, port, key)
			}
			return nil
		}

		return fmt.Errorf("unknown host key status")
	}
}

// InsecureIgnoreHostKey returns a callback that accepts any host key (for testing only)
func InsecureIgnoreHostKey() ssh.HostKeyCallback {
	return ssh.InsecureIgnoreHostKey()
}
