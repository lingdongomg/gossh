package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	appName        = "gossh"
	configFile     = "config.yaml"
	knownHostsFile = "known_hosts"
)

// ConfigDir returns the configuration directory path
func ConfigDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	default: // linux, darwin
		baseDir = os.Getenv("XDG_CONFIG_HOME")
		if baseDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(home, ".config")
		}
	}

	return filepath.Join(baseDir, appName), nil
}

// ConfigPath returns the full path to the config file
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFile), nil
}

// GetKnownHostsPath returns the path to the known_hosts file
func GetKnownHostsPath() string {
	dir, err := ConfigDir()
	if err != nil {
		// Fallback to current directory
		return knownHostsFile
	}
	return filepath.Join(dir, knownHostsFile)
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0700)
}
