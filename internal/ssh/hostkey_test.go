package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestFormatFingerprint(t *testing.T) {
	// Generate a test key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	pubKey := signer.PublicKey()

	// Test fingerprint formatting
	fingerprint := FormatFingerprint(pubKey)
	if fingerprint == "" {
		t.Error("Expected non-empty fingerprint")
	}

	// Should contain key type and SHA256
	if len(fingerprint) < 10 {
		t.Errorf("Fingerprint too short: %s", fingerprint)
	}

	// Should contain SHA256 prefix
	if !contains(fingerprint, "SHA256:") {
		t.Errorf("Fingerprint should contain SHA256: got %s", fingerprint)
	}
}

func TestHostKeyManager(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "gossh-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a manager with custom path
	hkm := &HostKeyManager{
		knownHosts: make(map[string]string),
		filePath:   filepath.Join(tmpDir, "known_hosts"),
	}

	// Generate test key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	pubKey := signer.PublicKey()

	// Test adding a host
	t.Run("AddHost", func(t *testing.T) {
		err := hkm.AddHost("test.example.com", 22, pubKey)
		if err != nil {
			t.Errorf("Failed to add host: %v", err)
		}

		// Check file was created
		if _, err := os.Stat(hkm.filePath); os.IsNotExist(err) {
			t.Error("Known hosts file was not created")
		}
	})

	// Test checking a known host
	t.Run("CheckKnownHost", func(t *testing.T) {
		result := hkm.CheckHostKey("test.example.com", 22, pubKey)
		if result.Status != HostKeyOK {
			t.Errorf("Expected HostKeyOK, got %v", result.Status)
		}
	})

	// Test checking an unknown host
	t.Run("CheckUnknownHost", func(t *testing.T) {
		result := hkm.CheckHostKey("unknown.example.com", 22, pubKey)
		if result.Status != HostKeyNew {
			t.Errorf("Expected HostKeyNew, got %v", result.Status)
		}
	})

	// Test with non-standard port
	t.Run("AddHostNonStandardPort", func(t *testing.T) {
		err := hkm.AddHost("custom.example.com", 2222, pubKey)
		if err != nil {
			t.Errorf("Failed to add host with non-standard port: %v", err)
		}

		result := hkm.CheckHostKey("custom.example.com", 2222, pubKey)
		if result.Status != HostKeyOK {
			t.Errorf("Expected HostKeyOK, got %v", result.Status)
		}
	})
}

func TestFormatHostPort(t *testing.T) {
	tests := []struct {
		host     string
		port     int
		expected string
	}{
		{"example.com", 22, "example.com"},
		{"example.com", 2222, "[example.com]:2222"},
		{"192.168.1.1", 22, "192.168.1.1"},
		{"192.168.1.1", 443, "[192.168.1.1]:443"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatHostPort(tt.host, tt.port)
			if result != tt.expected {
				t.Errorf("formatHostPort(%s, %d) = %s; want %s", tt.host, tt.port, result, tt.expected)
			}
		})
	}
}

func TestHostKeyChanged(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "gossh-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hkm := &HostKeyManager{
		knownHosts: make(map[string]string),
		filePath:   filepath.Join(tmpDir, "known_hosts"),
	}

	// Generate first key
	_, privateKey1, _ := ed25519.GenerateKey(rand.Reader)
	signer1, _ := ssh.NewSignerFromKey(privateKey1)
	pubKey1 := signer1.PublicKey()

	// Generate second key
	_, privateKey2, _ := ed25519.GenerateKey(rand.Reader)
	signer2, _ := ssh.NewSignerFromKey(privateKey2)
	pubKey2 := signer2.PublicKey()

	// Add first key
	hkm.AddHost("changed.example.com", 22, pubKey1)

	// Check with different key
	result := hkm.CheckHostKey("changed.example.com", 22, pubKey2)
	if result.Status != HostKeyChanged {
		t.Errorf("Expected HostKeyChanged, got %v", result.Status)
	}
	if result.OldKey == "" {
		t.Error("Expected OldKey to be set for changed key")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
