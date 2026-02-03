package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	enc, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple text", "hello world"},
		{"unicode", "你好世界 🌍"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"long text", strings.Repeat("a", 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := enc.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			decrypted, err := enc.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewEncryptor(key)

	plaintext := "test message"
	cipher1, _ := enc.Encrypt(plaintext)
	cipher2, _ := enc.Encrypt(plaintext)

	if cipher1 == cipher2 {
		t.Error("Same plaintext should produce different ciphertexts (due to random nonce)")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewEncryptor(key)

	tests := []struct {
		name       string
		ciphertext string
	}{
		{"invalid base64", "not-valid-base64!"},
		{"too short", "YWJjZA=="}, // "abcd" in base64, too short for nonce
		{"wrong key decryption", func() string {
			otherKey, _ := GenerateKey()
			otherEnc, _ := NewEncryptor(otherKey)
			ct, _ := otherEnc.Encrypt("secret")
			return ct
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.Decrypt(tt.ciphertext)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestNewEncryptorKeyLength(t *testing.T) {
	tests := []struct {
		name    string
		keyLen  int
		wantErr bool
	}{
		{"correct length 32", 32, false},
		{"too short", 16, true},
		{"too long", 64, true},
		{"empty", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLen)
			_, err := NewEncryptor(key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEncryptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if salt1 == salt2 {
		t.Error("GenerateSalt should produce different salts")
	}

	if len(salt1) == 0 {
		t.Error("GenerateSalt returned empty salt")
	}
}
