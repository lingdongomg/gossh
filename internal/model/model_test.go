package model

import (
	"testing"
)

func TestConnectionValidate(t *testing.T) {
	tests := []struct {
		name    string
		conn    Connection
		wantErr error
	}{
		{
			name: "valid password auth",
			conn: Connection{
				Name:       "test",
				Host:       "example.com",
				User:       "admin",
				Port:       22,
				AuthMethod: AuthPassword,
			},
			wantErr: nil,
		},
		{
			name: "valid key auth",
			conn: Connection{
				Name:       "test",
				Host:       "example.com",
				User:       "admin",
				Port:       22,
				AuthMethod: AuthKey,
				KeyPath:    "/home/user/.ssh/id_rsa",
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			conn: Connection{
				Host: "example.com",
				User: "admin",
				Port: 22,
			},
			wantErr: ErrNameRequired,
		},
		{
			name: "missing host",
			conn: Connection{
				Name: "test",
				User: "admin",
				Port: 22,
			},
			wantErr: ErrHostRequired,
		},
		{
			name: "missing user",
			conn: Connection{
				Name: "test",
				Host: "example.com",
				Port: 22,
			},
			wantErr: ErrUserRequired,
		},
		{
			name: "invalid port zero",
			conn: Connection{
				Name: "test",
				Host: "example.com",
				User: "admin",
				Port: 0,
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "invalid port negative",
			conn: Connection{
				Name: "test",
				Host: "example.com",
				User: "admin",
				Port: -1,
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "invalid port too high",
			conn: Connection{
				Name: "test",
				Host: "example.com",
				User: "admin",
				Port: 70000,
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "key auth missing key path",
			conn: Connection{
				Name:       "test",
				Host:       "example.com",
				User:       "admin",
				Port:       22,
				AuthMethod: AuthKey,
			},
			wantErr: ErrKeyPathRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conn.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnectionMatchesFilter(t *testing.T) {
	conn := Connection{
		Name:  "Production Web",
		Host:  "web.example.com",
		User:  "admin",
		Group: "Production",
		Tags:  []string{"web", "nginx"},
	}

	tests := []struct {
		name   string
		filter string
		want   bool
	}{
		{"empty filter", "", true},
		{"match name", "production", true},
		{"match name case insensitive", "PRODUCTION", true},
		{"match host", "example", true},
		{"match user", "admin", true},
		{"match group", "production", true},
		{"match tag", "nginx", true},
		{"no match", "database", false},
		{"partial match name", "web", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := conn.MatchesFilter(tt.filter)
			if got != tt.want {
				t.Errorf("MatchesFilter(%q) = %v, want %v", tt.filter, got, tt.want)
			}
		})
	}
}

func TestNewConnection(t *testing.T) {
	conn := NewConnection()

	if conn.ID == "" {
		t.Error("NewConnection should generate ID")
	}

	if conn.Port != 22 {
		t.Errorf("NewConnection port = %d, want 22", conn.Port)
	}

	if conn.AuthType != "" && conn.AuthMethod != AuthPassword {
		t.Errorf("NewConnection AuthMethod = %v, want %v", conn.AuthMethod, AuthPassword)
	}

	if conn.LastStatus != ConnStatusUnknown {
		t.Errorf("NewConnection LastStatus = %v, want %v", conn.LastStatus, ConnStatusUnknown)
	}
}

func TestNewSettings(t *testing.T) {
	settings := NewSettings()

	if settings.Initialized {
		t.Error("NewSettings should not be initialized")
	}

	if settings.PasswordProtectionEnabled {
		t.Error("NewSettings should not have password protection enabled")
	}

	if settings.DefaultPort != 22 {
		t.Errorf("NewSettings DefaultPort = %d, want 22", settings.DefaultPort)
	}

	if settings.Language != "en" {
		t.Errorf("NewSettings Language = %q, want %q", settings.Language, "en")
	}
}

func TestSettingsNeedsUnlock(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		want     bool
	}{
		{
			name: "no password set",
			settings: Settings{
				PasswordProtectionEnabled: true,
				MasterPasswordHash:        "",
			},
			want: false,
		},
		{
			name: "password protection disabled",
			settings: Settings{
				PasswordProtectionEnabled: false,
				MasterPasswordHash:        "somehash",
			},
			want: false,
		},
		{
			name: "needs unlock",
			settings: Settings{
				PasswordProtectionEnabled: true,
				MasterPasswordHash:        "somehash",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.settings.NeedsUnlock()
			if got != tt.want {
				t.Errorf("NeedsUnlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigGetGroups(t *testing.T) {
	cfg := Config{
		Groups: []Group{
			{Name: "Production"},
			{Name: "Development"},
		},
	}

	groups := cfg.GetGroups()

	if len(groups) != 3 { // "" + 2 groups
		t.Errorf("GetGroups() returned %d groups, want 3", len(groups))
	}

	if groups[0] != "" {
		t.Error("First group should be empty string (ungrouped)")
	}
}
