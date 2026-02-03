package crypto

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/argon2"
)

// DeriveKey derives an encryption key from a password and salt using Argon2id
func DeriveKey(password string, salt string) ([]byte, error) {
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		// If salt is not base64, use it as raw bytes
		saltBytes = []byte(salt)
	}

	// Use Argon2id for key derivation
	key := argon2.IDKey(
		[]byte(password),
		saltBytes,
		argonTime,
		argonMemory,
		argonThreads,
		32, // 256-bit key
	)

	return key, nil
}

// CryptoService manages encryption/decryption with a master password
type CryptoService struct {
	encryptor *Encryptor
	salt      string
}

// NewCryptoService creates a new crypto service
func NewCryptoService(masterPassword, salt string) (*CryptoService, error) {
	key, err := DeriveKey(masterPassword, salt)
	if err != nil {
		return nil, err
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		return nil, err
	}

	return &CryptoService{
		encryptor: encryptor,
		salt:      salt,
	}, nil
}

// NewCryptoServiceWithKey creates a new crypto service with a pre-derived key
func NewCryptoServiceWithKey(key []byte, salt string) (*CryptoService, error) {
	// Ensure key is 32 bytes
	var finalKey []byte
	if len(key) == 32 {
		finalKey = key
	} else {
		// Hash the key to get 32 bytes
		hash := sha256.Sum256(key)
		finalKey = hash[:]
	}

	encryptor, err := NewEncryptor(finalKey)
	if err != nil {
		return nil, err
	}

	return &CryptoService{
		encryptor: encryptor,
		salt:      salt,
	}, nil
}

// Encrypt encrypts plaintext
func (c *CryptoService) Encrypt(plaintext string) (string, error) {
	return c.encryptor.Encrypt(plaintext)
}

// Decrypt decrypts ciphertext
func (c *CryptoService) Decrypt(ciphertext string) (string, error) {
	return c.encryptor.Decrypt(ciphertext)
}

// GetSalt returns the salt used for key derivation
func (c *CryptoService) GetSalt() string {
	return c.salt
}
