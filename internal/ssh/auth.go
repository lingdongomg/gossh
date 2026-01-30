package ssh

import (
	"os"

	"golang.org/x/crypto/ssh"
	"gossh/internal/model"
)

// BuildAuthMethods creates SSH auth methods for a connection
func BuildAuthMethods(conn model.Connection) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	switch conn.AuthMethod {
	case model.AuthPassword:
		methods = append(methods, ssh.Password(conn.Password))
	case model.AuthKey:
		keyAuth, err := loadKeyAuth(conn.KeyPath, conn.KeyPassword)
		if err != nil {
			return nil, err
		}
		methods = append(methods, keyAuth)
	}

	return methods, nil
}

// loadKeyAuth loads a private key for authentication
func loadKeyAuth(keyPath, passphrase string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(key)
	}
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}
