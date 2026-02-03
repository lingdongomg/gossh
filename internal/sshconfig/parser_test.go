package sshconfig

import (
	"os"
	"path/filepath"
	"testing"

	"gossh/internal/model"
)

func TestParseFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	content := `# SSH Config for testing
Host myserver
    HostName 192.168.1.100
    User admin
    Port 2222
    IdentityFile ~/.ssh/id_rsa

Host webserver
    HostName web.example.com
    User deploy
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	parser := NewParser()
	connections, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(connections))
	}

	// Check first connection
	var myserver, webserver model.Connection
	for _, c := range connections {
		if c.Name == "myserver" {
			myserver = c
		}
		if c.Name == "webserver" {
			webserver = c
		}
	}

	if myserver.Name == "" {
		t.Error("myserver not found")
	} else {
		if myserver.Host != "192.168.1.100" {
			t.Errorf("myserver host = %q, want %q", myserver.Host, "192.168.1.100")
		}
		if myserver.User != "admin" {
			t.Errorf("myserver user = %q, want %q", myserver.User, "admin")
		}
		if myserver.Port != 2222 {
			t.Errorf("myserver port = %d, want %d", myserver.Port, 2222)
		}
		if myserver.AuthType != model.AuthKey {
			t.Errorf("myserver authType = %v, want %v", myserver.AuthType, model.AuthKey)
		}
	}

	if webserver.Name == "" {
		t.Error("webserver not found")
	} else {
		if webserver.Host != "web.example.com" {
			t.Errorf("webserver host = %q, want %q", webserver.Host, "web.example.com")
		}
		if webserver.Port != 22 {
			t.Errorf("webserver port = %d, want %d", webserver.Port, 22)
		}
	}
}

func TestParseFileSkipsWildcards(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	content := `Host *
    ServerAliveInterval 60

Host *.example.com
    User deploy

Host myserver
    HostName 192.168.1.1
    User admin
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	parser := NewParser()
	connections, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Should only have myserver, wildcard entries should be skipped
	if len(connections) != 1 {
		t.Errorf("Expected 1 connection (wildcards skipped), got %d", len(connections))
	}

	if len(connections) > 0 && connections[0].Name != "myserver" {
		t.Errorf("Expected myserver, got %q", connections[0].Name)
	}
}

func TestParseFileHostAsHostname(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	// When HostName is missing, Host pattern should be used as hostname
	content := `Host server.example.com
    User admin
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	parser := NewParser()
	connections, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(connections) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(connections))
	}

	if connections[0].Host != "server.example.com" {
		t.Errorf("Host = %q, want %q", connections[0].Host, "server.example.com")
	}
}

func TestParseFileEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")

	content := `# Just comments
# Nothing here
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	parser := NewParser()
	connections, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(connections) != 0 {
		t.Errorf("Expected 0 connections, got %d", len(connections))
	}
}

func TestParseFileNotFound(t *testing.T) {
	parser := NewParser()
	_, err := parser.ParseFile("/nonexistent/path/config")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestMerge(t *testing.T) {
	existing := []model.Connection{
		{Name: "server1", Host: "1.1.1.1"},
		{Name: "server2", Host: "2.2.2.2"},
	}

	imported := []model.Connection{
		{Name: "server1", Host: "new1.1.1.1"}, // Duplicate
		{Name: "server3", Host: "3.3.3.3"},     // New
		{Name: "Server2", Host: "new2.2.2.2"}, // Duplicate (case insensitive)
	}

	newConns, skipped := Merge(existing, imported)

	if len(newConns) != 1 {
		t.Errorf("Expected 1 new connection, got %d", len(newConns))
	}

	if skipped != 2 {
		t.Errorf("Expected 2 skipped, got %d", skipped)
	}

	if len(newConns) > 0 && newConns[0].Name != "server3" {
		t.Errorf("Expected server3, got %q", newConns[0].Name)
	}
}

func TestMergeEmpty(t *testing.T) {
	existing := []model.Connection{}
	imported := []model.Connection{
		{Name: "server1", Host: "1.1.1.1"},
	}

	newConns, skipped := Merge(existing, imported)

	if len(newConns) != 1 {
		t.Errorf("Expected 1 new connection, got %d", len(newConns))
	}

	if skipped != 0 {
		t.Errorf("Expected 0 skipped, got %d", skipped)
	}
}
