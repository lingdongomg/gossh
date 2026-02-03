package model

import (
	"time"

	"github.com/google/uuid"
)

// AuthType represents the authentication method
type AuthType string

const (
	AuthPassword AuthType = "password"
	AuthKey      AuthType = "key"
)

// ConnStatus represents the connection status
type ConnStatus string

const (
	ConnStatusSuccess ConnStatus = "success"
	ConnStatusFailed  ConnStatus = "failed"
	ConnStatusUnknown ConnStatus = "unknown"
)

// Connection represents an SSH connection configuration
type Connection struct {
	ID                     string     `yaml:"id"`
	Name                   string     `yaml:"name"`
	Host                   string     `yaml:"host"`
	Port                   int        `yaml:"port"`
	User                   string     `yaml:"user"`
	AuthType               AuthType   `yaml:"auth_type"`
	AuthMethod             AuthType   `yaml:"auth_method"` // Deprecated: use AuthType
	Password               string     `yaml:"password,omitempty"`               // Plain text (for runtime use)
	EncryptedPassword      string     `yaml:"encrypted_password,omitempty"`      // AES-256-GCM encrypted
	KeyPath                string     `yaml:"key_path,omitempty"`
	KeyPassword            string     `yaml:"key_password,omitempty"`            // Plain text (for runtime use)
	EncryptedKeyPassphrase string     `yaml:"encrypted_key_passphrase,omitempty"` // AES-256-GCM encrypted
	Group                  string     `yaml:"group,omitempty"`
	Tags                   []string   `yaml:"tags,omitempty"`
	StartupCommand         string     `yaml:"startup_command,omitempty"`
	LastConnected          *time.Time `yaml:"last_connected,omitempty"`
	LastStatus             ConnStatus `yaml:"last_status"`
	HealthStatus           ConnStatus `yaml:"health_status,omitempty"` // For health check results
	CreatedAt              time.Time  `yaml:"created_at"`
	UpdatedAt              time.Time  `yaml:"updated_at"`
}

// NewConnection creates a new connection with defaults
func NewConnection() Connection {
	return Connection{
		ID:         uuid.New().String(),
		Port:       22,
		AuthMethod: AuthPassword,
		LastStatus: ConnStatusUnknown,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Validate checks if the connection has required fields
func (c *Connection) Validate() error {
	if c.Name == "" {
		return ErrNameRequired
	}
	if c.Host == "" {
		return ErrHostRequired
	}
	if c.User == "" {
		return ErrUserRequired
	}
	if c.Port <= 0 || c.Port > 65535 {
		return ErrInvalidPort
	}
	if c.AuthMethod == AuthKey && c.KeyPath == "" {
		return ErrKeyPathRequired
	}
	return nil
}

// MatchesFilter checks if connection matches search filter
func (c *Connection) MatchesFilter(filter string) bool {
	if filter == "" {
		return true
	}
	filter = toLower(filter)
	if contains(toLower(c.Name), filter) {
		return true
	}
	if contains(toLower(c.Host), filter) {
		return true
	}
	if contains(toLower(c.User), filter) {
		return true
	}
	if contains(toLower(c.Group), filter) {
		return true
	}
	for _, tag := range c.Tags {
		if contains(toLower(tag), filter) {
			return true
		}
	}
	return false
}

// Group represents a connection group
type Group struct {
	Name  string `yaml:"name"`
	Color string `yaml:"color"`
}

// Settings represents application settings
type Settings struct {
	MasterPasswordHash        string `yaml:"master_password_hash,omitempty"`
	EncryptionSalt            string `yaml:"encryption_salt,omitempty"`
	PasswordProtectionEnabled bool   `yaml:"password_protection_enabled"`
	Initialized               bool   `yaml:"initialized"` // True after first-time setup
	ConnectionTimeout         int    `yaml:"connection_timeout"`
	DefaultPort               int    `yaml:"default_port"`
	Theme                     string `yaml:"theme"`
	Language                  string `yaml:"language,omitempty"` // "en" or "zh"
}

// NewSettings creates default settings
func NewSettings() Settings {
	return Settings{
		PasswordProtectionEnabled: false,
		Initialized:               false,
		ConnectionTimeout:         10,
		DefaultPort:               22,
		Theme:                     "dark",
		Language:                  "en",
	}
}

// IsPasswordSet returns true if master password has been set
func (s *Settings) IsPasswordSet() bool {
	return s.MasterPasswordHash != ""
}

// NeedsUnlock returns true if password protection is enabled and password is set
func (s *Settings) NeedsUnlock() bool {
	return s.PasswordProtectionEnabled && s.MasterPasswordHash != ""
}

// Config represents the application configuration
type Config struct {
	Version     string       `yaml:"version"`
	Settings    Settings     `yaml:"settings"`
	Groups      []Group      `yaml:"groups"`
	Connections []Connection `yaml:"connections"`
}

// NewConfig creates a new config with defaults
func NewConfig() Config {
	return Config{
		Version:  "1.0",
		Settings: NewSettings(),
		Groups: []Group{
			{Name: "Production", Color: "#ff6b6b"},
			{Name: "Development", Color: "#4ecdc4"},
			{Name: "Testing", Color: "#ffe66d"},
		},
		Connections: []Connection{},
	}
}

// GetGroups returns all group names including "Ungrouped"
func (c *Config) GetGroups() []string {
	groups := make([]string, 0, len(c.Groups)+1)
	groups = append(groups, "") // Empty string for ungrouped
	for _, g := range c.Groups {
		groups = append(groups, g.Name)
	}
	return groups
}

// GetConnectionsByGroup returns connections grouped by group name
func (c *Config) GetConnectionsByGroup() map[string][]Connection {
	result := make(map[string][]Connection)
	for _, conn := range c.Connections {
		group := conn.Group
		if group == "" {
			group = "Ungrouped"
		}
		result[group] = append(result[group], conn)
	}
	return result
}

// Errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

var (
	ErrNameRequired    = ValidationError{Field: "name", Message: "name is required"}
	ErrHostRequired    = ValidationError{Field: "host", Message: "host is required"}
	ErrUserRequired    = ValidationError{Field: "user", Message: "user is required"}
	ErrInvalidPort     = ValidationError{Field: "port", Message: "port must be between 1 and 65535"}
	ErrKeyPathRequired = ValidationError{Field: "key_path", Message: "key path is required for key authentication"}
)

// Helper functions for case-insensitive matching
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
