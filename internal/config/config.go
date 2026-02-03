package config

import (
	"errors"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	"gossh/internal/crypto"
	"gossh/internal/model"
)

// Manager handles configuration persistence
type Manager struct {
	mu            sync.RWMutex
	config        model.Config
	path          string
	cryptoService *crypto.CryptoService
	unlocked      bool
}

// NewManager creates a new config manager
func NewManager() (*Manager, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	m := &Manager{
		config: model.NewConfig(),
		path:   path,
	}

	if err := m.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return m, nil
}

// Load reads the config from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}

	var cfg model.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	m.config = cfg
	return nil
}

// Save writes the config to disk
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.saveUnlocked()
}

// IsFirstRun returns true if the app has not been initialized yet
func (m *Manager) IsFirstRun() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return !m.config.Settings.Initialized
}

// IsUnlocked returns true if the config has been unlocked or doesn't need unlock
func (m *Manager) IsUnlocked() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// If password protection is disabled, always unlocked
	if !m.config.Settings.PasswordProtectionEnabled {
		return true
	}
	// If password protection is enabled, check unlocked state
	return m.unlocked
}

// SetupMasterPassword sets the master password for the first time
func (m *Manager) SetupMasterPassword(password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Settings.IsPasswordSet() {
		return errors.New("master password already set")
	}

	// Generate salt
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}

	// Hash password
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	// Create crypto service
	cryptoService, err := crypto.NewCryptoService(password, salt)
	if err != nil {
		return err
	}

	m.config.Settings.MasterPasswordHash = hash
	m.config.Settings.EncryptionSalt = salt
	m.config.Settings.PasswordProtectionEnabled = true
	m.config.Settings.Initialized = true
	m.cryptoService = cryptoService
	m.unlocked = true

	return m.saveUnlocked()
}

// SetupWithoutPassword initializes the app without password protection
func (m *Manager) SetupWithoutPassword() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Settings.Initialized {
		return errors.New("already initialized")
	}

	// Generate salt for encryption (we still encrypt data, just auto-unlock)
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}

	// Use machine-derived key for encryption when no password is set
	// This provides better security than a fixed key
	machineKey, err := crypto.DeriveKeyFromMachine()
	if err != nil {
		// Fallback to a less secure but functional approach
		machineKey = []byte(crypto.GetMachineID())
	}
	
	cryptoService, err := crypto.NewCryptoServiceWithKey(machineKey, salt)
	if err != nil {
		return err
	}

	m.config.Settings.EncryptionSalt = salt
	m.config.Settings.PasswordProtectionEnabled = false
	m.config.Settings.Initialized = true
	m.cryptoService = cryptoService
	m.unlocked = true

	return m.saveUnlocked()
}

// Unlock verifies the master password and unlocks the config
func (m *Manager) Unlock(password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If password protection is disabled, auto-unlock
	if !m.config.Settings.PasswordProtectionEnabled {
		return m.autoUnlock()
	}

	if !m.config.Settings.IsPasswordSet() {
		m.unlocked = true
		return nil
	}

	// Verify password
	valid, err := crypto.VerifyPassword(password, m.config.Settings.MasterPasswordHash)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid password")
	}

	// Create crypto service
	cryptoService, err := crypto.NewCryptoService(password, m.config.Settings.EncryptionSalt)
	if err != nil {
		return err
	}

	m.cryptoService = cryptoService
	m.unlocked = true

	// Decrypt all connection passwords
	for i := range m.config.Connections {
		conn := &m.config.Connections[i]
		if conn.EncryptedPassword != "" {
			decrypted, err := m.cryptoService.Decrypt(conn.EncryptedPassword)
			if err == nil {
				conn.Password = decrypted
			}
		}
		if conn.EncryptedKeyPassphrase != "" {
			decrypted, err := m.cryptoService.Decrypt(conn.EncryptedKeyPassphrase)
			if err == nil {
				conn.KeyPassword = decrypted
			}
		}
	}

	return nil
}

// autoUnlock unlocks without password (for password protection disabled mode)
func (m *Manager) autoUnlock() error {
	if m.config.Settings.EncryptionSalt == "" {
		m.unlocked = true
		return nil
	}

	// Use machine-derived key for decryption
	machineKey, err := crypto.DeriveKeyFromMachine()
	if err != nil {
		// Fallback
		machineKey = []byte(crypto.GetMachineID())
	}
	
	cryptoService, err := crypto.NewCryptoServiceWithKey(machineKey, m.config.Settings.EncryptionSalt)
	if err != nil {
		return err
	}

	m.cryptoService = cryptoService
	m.unlocked = true

	// Decrypt all connection passwords
	for i := range m.config.Connections {
		conn := &m.config.Connections[i]
		if conn.EncryptedPassword != "" {
			decrypted, err := m.cryptoService.Decrypt(conn.EncryptedPassword)
			if err == nil {
				conn.Password = decrypted
			}
		}
		if conn.EncryptedKeyPassphrase != "" {
			decrypted, err := m.cryptoService.Decrypt(conn.EncryptedKeyPassphrase)
			if err == nil {
				conn.KeyPassword = decrypted
			}
		}
	}

	return nil
}

// AutoUnlockIfNeeded automatically unlocks if password protection is disabled
func (m *Manager) AutoUnlockIfNeeded() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.Settings.PasswordProtectionEnabled && m.config.Settings.Initialized {
		return m.autoUnlock()
	}
	return nil
}

// Connections returns all connections
func (m *Manager) Connections() []model.Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]model.Connection, len(m.config.Connections))
	copy(result, m.config.Connections)
	return result
}

// GetConnection returns a connection by ID
func (m *Manager) GetConnection(id string) (model.Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.config.Connections {
		if c.ID == id {
			return c, true
		}
	}
	return model.Connection{}, false
}

// AddConnection adds a new connection
func (m *Manager) AddConnection(conn model.Connection) error {
	if err := conn.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	conn.CreatedAt = time.Now()
	conn.UpdatedAt = time.Now()

	// Encrypt sensitive data if crypto service is available
	if m.cryptoService != nil {
		if conn.Password != "" {
			encrypted, err := m.cryptoService.Encrypt(conn.Password)
			if err != nil {
				return err
			}
			conn.EncryptedPassword = encrypted
		}
		if conn.KeyPassword != "" {
			encrypted, err := m.cryptoService.Encrypt(conn.KeyPassword)
			if err != nil {
				return err
			}
			conn.EncryptedKeyPassphrase = encrypted
		}
	}

	m.config.Connections = append(m.config.Connections, conn)

	return m.saveUnlocked()
}

// UpdateConnection updates an existing connection
func (m *Manager) UpdateConnection(conn model.Connection) error {
	if err := conn.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.config.Connections {
		if c.ID == conn.ID {
			conn.CreatedAt = c.CreatedAt
			conn.UpdatedAt = time.Now()

			// Encrypt sensitive data if crypto service is available
			if m.cryptoService != nil {
				if conn.Password != "" {
					encrypted, err := m.cryptoService.Encrypt(conn.Password)
					if err != nil {
						return err
					}
					conn.EncryptedPassword = encrypted
				}
				if conn.KeyPassword != "" {
					encrypted, err := m.cryptoService.Encrypt(conn.KeyPassword)
					if err != nil {
						return err
					}
					conn.EncryptedKeyPassphrase = encrypted
				}
			}

			m.config.Connections[i] = conn
			return m.saveUnlocked()
		}
	}

	return errors.New("connection not found")
}

// UpdateConnectionStatus updates the last connection status
func (m *Manager) UpdateConnectionStatus(id string, status model.ConnStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.config.Connections {
		if c.ID == id {
			now := time.Now()
			m.config.Connections[i].LastConnected = &now
			m.config.Connections[i].LastStatus = status
			m.config.Connections[i].UpdatedAt = now
			return m.saveUnlocked()
		}
	}

	return errors.New("connection not found")
}

// DeleteConnection removes a connection by ID
func (m *Manager) DeleteConnection(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.config.Connections {
		if c.ID == id {
			m.config.Connections = append(m.config.Connections[:i], m.config.Connections[i+1:]...)
			return m.saveUnlocked()
		}
	}

	return errors.New("connection not found")
}

// Groups returns all groups
func (m *Manager) Groups() []model.Group {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]model.Group, len(m.config.Groups))
	copy(result, m.config.Groups)
	return result
}

// GroupNames returns all group names
func (m *Manager) GroupNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, len(m.config.Groups))
	for i, g := range m.config.Groups {
		names[i] = g.Name
	}
	return names
}

// AddGroup adds a new group
func (m *Manager) AddGroup(group model.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, g := range m.config.Groups {
		if g.Name == group.Name {
			return errors.New("group already exists")
		}
	}

	m.config.Groups = append(m.config.Groups, group)
	return m.saveUnlocked()
}

// Settings returns the current settings
func (m *Manager) Settings() model.Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Settings
}

// Config returns the full config (for export)
func (m *Manager) Config() model.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// ImportConnections imports connections from another config
func (m *Manager) ImportConnections(connections []model.Connection, overwrite bool) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	imported := 0
	for _, conn := range connections {
		// Check if connection with same name exists
		found := false
		for i, c := range m.config.Connections {
			if c.Name == conn.Name {
				found = true
				if overwrite {
					conn.ID = c.ID
					conn.CreatedAt = c.CreatedAt
					conn.UpdatedAt = time.Now()

					// Re-encrypt passwords if crypto service available
					if m.cryptoService != nil && conn.Password != "" {
						encrypted, err := m.cryptoService.Encrypt(conn.Password)
						if err == nil {
							conn.EncryptedPassword = encrypted
						}
					}
					if m.cryptoService != nil && conn.KeyPassword != "" {
						encrypted, err := m.cryptoService.Encrypt(conn.KeyPassword)
						if err == nil {
							conn.EncryptedKeyPassphrase = encrypted
						}
					}

					m.config.Connections[i] = conn
					imported++
				}
				break
			}
		}

		if !found {
			conn.ID = model.NewConnection().ID
			conn.CreatedAt = time.Now()
			conn.UpdatedAt = time.Now()

			// Encrypt passwords if crypto service available
			if m.cryptoService != nil && conn.Password != "" {
				encrypted, err := m.cryptoService.Encrypt(conn.Password)
				if err == nil {
					conn.EncryptedPassword = encrypted
				}
			}
			if m.cryptoService != nil && conn.KeyPassword != "" {
				encrypted, err := m.cryptoService.Encrypt(conn.KeyPassword)
				if err == nil {
					conn.EncryptedKeyPassphrase = encrypted
				}
			}

			m.config.Connections = append(m.config.Connections, conn)
			imported++
		}
	}

	if imported > 0 {
		if err := m.saveUnlocked(); err != nil {
			return 0, err
		}
	}

	return imported, nil
}

// saveUnlocked saves without acquiring lock (caller must hold lock)
func (m *Manager) saveUnlocked() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	// Create a copy for saving (without plain text passwords)
	saveCfg := m.config
	saveCfg.Connections = make([]model.Connection, len(m.config.Connections))
	for i, conn := range m.config.Connections {
		saveCfg.Connections[i] = conn
		// Clear plain text passwords from saved config
		saveCfg.Connections[i].Password = ""
		saveCfg.Connections[i].KeyPassword = ""
	}

	data, err := yaml.Marshal(&saveCfg)
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0600)
}

// IsPasswordProtected returns true if password protection is enabled
func (m *Manager) IsPasswordProtected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Settings.PasswordProtectionEnabled
}

// EnablePassword enables password protection with the given password
func (m *Manager) EnablePassword(password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate new salt
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}

	// Hash password
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	// Create crypto service with password
	cryptoService, err := crypto.NewCryptoService(password, salt)
	if err != nil {
		return err
	}

	// Re-encrypt all connection passwords with new key
	for i := range m.config.Connections {
		conn := &m.config.Connections[i]
		if conn.Password != "" {
			encrypted, err := cryptoService.Encrypt(conn.Password)
			if err != nil {
				return err
			}
			conn.EncryptedPassword = encrypted
		}
		if conn.KeyPassword != "" {
			encrypted, err := cryptoService.Encrypt(conn.KeyPassword)
			if err != nil {
				return err
			}
			conn.EncryptedKeyPassphrase = encrypted
		}
	}

	m.config.Settings.MasterPasswordHash = hash
	m.config.Settings.EncryptionSalt = salt
	m.config.Settings.PasswordProtectionEnabled = true
	m.cryptoService = cryptoService

	return m.saveUnlocked()
}

// DisablePassword disables password protection (requires current password verification)
func (m *Manager) DisablePassword(currentPassword string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify current password
	if m.config.Settings.MasterPasswordHash != "" {
		valid, err := crypto.VerifyPassword(currentPassword, m.config.Settings.MasterPasswordHash)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("invalid password")
		}
	}

	// Generate new salt for machine-based encryption
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}

	// Use machine-derived key for encryption
	machineKey, err := crypto.DeriveKeyFromMachine()
	if err != nil {
		machineKey = []byte(crypto.GetMachineID())
	}

	cryptoService, err := crypto.NewCryptoServiceWithKey(machineKey, salt)
	if err != nil {
		return err
	}

	// Re-encrypt all connection passwords with machine key
	for i := range m.config.Connections {
		conn := &m.config.Connections[i]
		if conn.Password != "" {
			encrypted, err := cryptoService.Encrypt(conn.Password)
			if err != nil {
				return err
			}
			conn.EncryptedPassword = encrypted
		}
		if conn.KeyPassword != "" {
			encrypted, err := cryptoService.Encrypt(conn.KeyPassword)
			if err != nil {
				return err
			}
			conn.EncryptedKeyPassphrase = encrypted
		}
	}

	m.config.Settings.MasterPasswordHash = ""
	m.config.Settings.EncryptionSalt = salt
	m.config.Settings.PasswordProtectionEnabled = false
	m.cryptoService = cryptoService

	return m.saveUnlocked()
}

// GetLanguage returns the configured language
func (m *Manager) GetLanguage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config.Settings.Language == "" {
		return "en"
	}
	return m.config.Settings.Language
}

// SetLanguage sets the language setting
func (m *Manager) SetLanguage(lang string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.Settings.Language = lang
	return m.saveUnlocked()
}

// GetSettings returns a copy of current settings
func (m *Manager) GetSettings() model.Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Settings
}
