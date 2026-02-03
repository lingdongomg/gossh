package config

import (
	"os"
	"path/filepath"
	"testing"

	"gossh/internal/model"
)

func TestGetConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir returned error: %v", err)
	}
	if dir == "" {
		t.Error("ConfigDir returned empty string")
	}

	// Should contain "gossh" in the path
	if !contains(dir, "gossh") {
		t.Errorf("Config dir should contain 'gossh': %s", dir)
	}
}

func TestGetConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath returned error: %v", err)
	}
	if path == "" {
		t.Error("ConfigPath returned empty string")
	}

	// Should end with config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("Expected config.yaml, got %s", filepath.Base(path))
	}
}

func TestGetKnownHostsPath(t *testing.T) {
	path := GetKnownHostsPath()
	if path == "" {
		t.Error("GetKnownHostsPath returned empty string")
	}

	// Should end with known_hosts
	if filepath.Base(path) != "known_hosts" {
		t.Errorf("Expected known_hosts, got %s", filepath.Base(path))
	}
}

func TestNewManagerCreatesDir(t *testing.T) {
	// Create a temp home directory
	tmpDir, err := os.MkdirTemp("", "gossh-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override HOME for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create a new manager
	cfg, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Check that IsFirstRun is true for new config
	if !cfg.IsFirstRun() {
		t.Error("Expected IsFirstRun to be true for new config")
	}
}

func TestManagerConnections(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "gossh-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Setup without password for testing
	if err := cfg.SetupWithoutPassword(); err != nil {
		t.Fatalf("Failed to setup without password: %v", err)
	}

	// Initially should have no connections
	conns := cfg.Connections()
	if len(conns) != 0 {
		t.Errorf("Expected 0 connections, got %d", len(conns))
	}

	// Add a connection
	conn := model.NewConnection()
	conn.Name = "test-server"
	conn.Host = "192.168.1.1"
	conn.User = "root"
	conn.Port = 22
	conn.AuthMethod = model.AuthPassword
	conn.Password = "test"

	if err := cfg.AddConnection(conn); err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Should have one connection
	conns = cfg.Connections()
	if len(conns) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(conns))
	}

	// Check the connection
	if conns[0].Name != "test-server" {
		t.Errorf("Expected name 'test-server', got '%s'", conns[0].Name)
	}
}

func TestManagerGroupNames(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gossh-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	cfg.SetupWithoutPassword()

	// Add connections with different groups
	conn1 := model.NewConnection()
	conn1.Name = "server1"
	conn1.Host = "192.168.1.1"
	conn1.User = "root"
	conn1.Port = 22
	conn1.AuthMethod = model.AuthPassword
	conn1.Password = "test"
	conn1.Group = "Production"
	cfg.AddConnection(conn1)

	conn2 := model.NewConnection()
	conn2.Name = "server2"
	conn2.Host = "192.168.1.2"
	conn2.User = "root"
	conn2.Port = 22
	conn2.AuthMethod = model.AuthPassword
	conn2.Password = "test"
	conn2.Group = "Development"
	cfg.AddConnection(conn2)

	conn3 := model.NewConnection()
	conn3.Name = "server3"
	conn3.Host = "192.168.1.3"
	conn3.User = "root"
	conn3.Port = 22
	conn3.AuthMethod = model.AuthPassword
	conn3.Password = "test"
	conn3.Group = "Production"
	cfg.AddConnection(conn3)

	groups := cfg.GroupNames()
	// Should contain at least Production and Development (may have default groups too)
	hasProduction := false
	hasDevelopment := false
	for _, g := range groups {
		if g == "Production" {
			hasProduction = true
		}
		if g == "Development" {
			hasDevelopment = true
		}
	}
	if !hasProduction {
		t.Error("Expected groups to contain 'Production'")
	}
	if !hasDevelopment {
		t.Error("Expected groups to contain 'Development'")
	}
}

func TestManagerDeleteConnection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gossh-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	cfg.SetupWithoutPassword()

	// Add a connection
	conn := model.NewConnection()
	conn.Name = "to-delete"
	conn.Host = "192.168.1.1"
	conn.User = "root"
	conn.Port = 22
	conn.AuthMethod = model.AuthPassword
	conn.Password = "test"
	cfg.AddConnection(conn)

	// Get the ID
	conns := cfg.Connections()
	if len(conns) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(conns))
	}
	id := conns[0].ID

	// Delete the connection
	if err := cfg.DeleteConnection(id); err != nil {
		t.Fatalf("Failed to delete connection: %v", err)
	}

	// Should have no connections
	conns = cfg.Connections()
	if len(conns) != 0 {
		t.Errorf("Expected 0 connections after delete, got %d", len(conns))
	}
}

func TestManagerGetConnection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gossh-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	cfg.SetupWithoutPassword()

	// Add connections
	conn1 := model.NewConnection()
	conn1.Name = "web-server"
	conn1.Host = "192.168.1.1"
	conn1.User = "root"
	conn1.Port = 22
	conn1.AuthMethod = model.AuthPassword
	conn1.Password = "test"
	cfg.AddConnection(conn1)

	// Get the connection ID
	conns := cfg.Connections()
	if len(conns) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(conns))
	}
	id := conns[0].ID

	// Get by ID
	found, exists := cfg.GetConnection(id)
	if !exists {
		t.Error("Expected to find connection by ID")
	}
	if found.Name != "web-server" {
		t.Errorf("Expected name 'web-server', got '%s'", found.Name)
	}

	// Get non-existent
	_, exists = cfg.GetConnection("non-existent-id")
	if exists {
		t.Error("Expected not to find 'non-existent-id'")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
