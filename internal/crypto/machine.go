package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"runtime"
	"strings"
)

// GetMachineID returns a unique identifier for the current machine
// This is used for no-password mode encryption
func GetMachineID() string {
	var parts []string

	// Get hostname
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		parts = append(parts, hostname)
	}

	// Get username
	if user := os.Getenv("USER"); user != "" {
		parts = append(parts, user)
	} else if user := os.Getenv("USERNAME"); user != "" {
		// Windows
		parts = append(parts, user)
	}

	// Get machine-specific ID
	machineID := getMachineUUID()
	if machineID != "" {
		parts = append(parts, machineID)
	}

	// Combine and hash
	combined := strings.Join(parts, ":")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// getMachineUUID attempts to get a machine-specific UUID
func getMachineUUID() string {
	switch runtime.GOOS {
	case "linux":
		return getLinuxMachineID()
	case "darwin":
		return getDarwinMachineID()
	case "windows":
		return getWindowsMachineID()
	default:
		return ""
	}
}

// getLinuxMachineID reads the machine-id from /etc/machine-id or /var/lib/dbus/machine-id
func getLinuxMachineID() string {
	paths := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	return ""
}

// getDarwinMachineID gets the IOPlatformUUID on macOS
func getDarwinMachineID() string {
	// Read from IOPlatformUUID using ioreg
	// For simplicity, we try to read from a cached location or use hostname hash
	// In a production environment, you'd want to exec ioreg -rd1 -c IOPlatformExpertDevice
	
	// Try reading from a common location
	data, err := os.ReadFile("/Library/Preferences/SystemConfiguration/com.apple.computer-name.plist")
	if err == nil {
		return hashBytes(data)
	}

	// Fallback: use home directory path as a semi-unique identifier
	if home, err := os.UserHomeDir(); err == nil {
		return hashString(home)
	}

	return ""
}

// getWindowsMachineID gets the MachineGuid from Windows registry
func getWindowsMachineID() string {
	// In a production environment, you'd read from:
	// HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography\MachineGuid
	// For simplicity, we use a combination of other factors

	// Try COMPUTERNAME environment variable
	if name := os.Getenv("COMPUTERNAME"); name != "" {
		return hashString(name + os.Getenv("USERDOMAIN"))
	}

	return ""
}

// hashString returns SHA256 hash of a string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes
}

// hashBytes returns SHA256 hash of bytes
func hashBytes(b []byte) string {
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes
}

// DeriveKeyFromMachine derives an encryption key from machine characteristics
func DeriveKeyFromMachine() ([]byte, error) {
	machineID := GetMachineID()
	
	// Use a fixed salt for machine-based key derivation
	// This is less secure than password-based encryption but provides convenience
	salt := "gossh-machine-key-v1"
	
	// Use Argon2id with reduced parameters for machine-based key
	// (Since the input has high entropy already)
	key, err := DeriveKey(machineID, salt)
	if err != nil {
		return nil, err
	}
	
	return key, nil
}
