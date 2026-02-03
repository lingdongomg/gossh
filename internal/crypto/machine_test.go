package crypto

import (
	"testing"
)

func TestGetMachineID(t *testing.T) {
	id := GetMachineID()

	if id == "" {
		t.Error("GetMachineID should return non-empty string")
	}

	// Should be consistent
	id2 := GetMachineID()
	if id != id2 {
		t.Error("GetMachineID should return consistent value")
	}

	// Should be a hex string (64 chars for SHA256)
	if len(id) != 64 {
		t.Errorf("GetMachineID should return 64-char hex string, got %d chars", len(id))
	}
}

func TestDeriveKeyFromMachine(t *testing.T) {
	key, err := DeriveKeyFromMachine()
	if err != nil {
		t.Fatalf("DeriveKeyFromMachine failed: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("DeriveKeyFromMachine should return 32-byte key, got %d bytes", len(key))
	}

	// Should be consistent
	key2, _ := DeriveKeyFromMachine()
	if string(key) != string(key2) {
		t.Error("DeriveKeyFromMachine should return consistent value")
	}
}

func TestDeriveKey(t *testing.T) {
	password := "testPassword"
	salt := "dGVzdFNhbHQ=" // "testSalt" in base64

	key, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("DeriveKey should return 32-byte key, got %d bytes", len(key))
	}

	// Same input should give same output
	key2, _ := DeriveKey(password, salt)
	if string(key) != string(key2) {
		t.Error("DeriveKey should be deterministic")
	}

	// Different password should give different key
	key3, _ := DeriveKey("differentPassword", salt)
	if string(key) == string(key3) {
		t.Error("Different password should give different key")
	}

	// Different salt should give different key
	key4, _ := DeriveKey(password, "differentSalt")
	if string(key) == string(key4) {
		t.Error("Different salt should give different key")
	}
}

func TestCryptoService(t *testing.T) {
	password := "masterPassword"
	salt, _ := GenerateSalt()

	svc, err := NewCryptoService(password, salt)
	if err != nil {
		t.Fatalf("NewCryptoService failed: %v", err)
	}

	plaintext := "sensitive data"
	
	// Encrypt
	ciphertext, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Decrypt
	decrypted, err := svc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("got %q, want %q", decrypted, plaintext)
	}
}

func TestCryptoServiceWithKey(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	salt := "testSalt"

	svc, err := NewCryptoServiceWithKey(key, salt)
	if err != nil {
		t.Fatalf("NewCryptoServiceWithKey failed: %v", err)
	}

	plaintext := "sensitive data"
	
	ciphertext, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := svc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("got %q, want %q", decrypted, plaintext)
	}
}

func TestCryptoServiceWithKeyShort(t *testing.T) {
	// Test with key shorter than 32 bytes (should be hashed)
	key := []byte("shortKey")
	salt := "testSalt"

	svc, err := NewCryptoServiceWithKey(key, salt)
	if err != nil {
		t.Fatalf("NewCryptoServiceWithKey failed: %v", err)
	}

	plaintext := "test"
	ciphertext, _ := svc.Encrypt(plaintext)
	decrypted, _ := svc.Decrypt(ciphertext)

	if decrypted != plaintext {
		t.Errorf("got %q, want %q", decrypted, plaintext)
	}
}
