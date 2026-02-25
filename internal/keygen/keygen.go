// Package keygen provides Ed25519 SSH key generation and file operations.
//
// This package implements the creation of OpenSSH-compatible Ed25519 key pairs
// from 32-byte seeds derived via SLIP-0010 hierarchical deterministic key derivation.
// It handles proper OpenSSH key formatting, including PEM encoding for private keys
// and authorized_keys format for public keys.
package keygen

import (
	"crypto/ed25519"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// SSHKeyPair holds the generated SSH key pair data.
//
// This structure contains both the private key in OpenSSH PEM format
// and the public key in OpenSSH authorized_keys format, along with
// the SHA-256 fingerprint of the public key.
type SSHKeyPair struct {
	PrivateKeyPEM []byte
	PublicKey     []byte
	Fingerprint   string
}

// Generate creates an OpenSSH Ed25519 key pair from a 32-byte seed.
//
// The seed must be exactly 32 bytes and is typically derived from a BIP39
// mnemonic using SLIP-0010 hierarchical deterministic key derivation.
//
// The comment parameter is added to the public key in authorized_keys format.
// If provided, it appears after the key type and base64-encoded key data.
// The derivation path is commonly used as the comment for identification.
//
// The generated key pair uses:
//   - OpenSSH private key format (PEM with "OPENSSH PRIVATE KEY" header)
//   - OpenSSH public key format (ssh-ed25519 <base64-key> <comment>)
//   - SHA-256 fingerprint for key identification
//
// Returns an error if key generation or marshaling fails.
func Generate(seed [32]byte, comment string) (*SSHKeyPair, error) {
	privKey := ed25519.NewKeyFromSeed(seed[:])

	pubKey, err := ssh.NewPublicKey(privKey.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %w", err)
	}

	pemBlock, err := ssh.MarshalPrivateKey(privKey, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	privPEM := pem.EncodeToMemory(pemBlock)

	pubBytes := ssh.MarshalAuthorizedKey(pubKey)

	if comment != "" {
		if len(pubBytes) > 0 && pubBytes[len(pubBytes)-1] == '\n' {
			pubBytes = append(pubBytes[:len(pubBytes)-1], []byte(" "+comment+"\n")...)
		}
	}

	return &SSHKeyPair{
		PrivateKeyPEM: privPEM,
		PublicKey:     pubBytes,
		Fingerprint:   ssh.FingerprintSHA256(pubKey),
	}, nil
}

// WriteKeyPair writes the private and public keys to the filesystem.
//
// The private key is written to the specified outputPath with permissions 0600
// (readable and writable only by the owner). The public key is written to
// outputPath + ".pub" with permissions 0644 (readable by all, writable by owner).
//
// If the force parameter is false and either file already exists, an error
// is returned to prevent accidental overwrites. If force is true, existing
// files are overwritten.
//
// Returns an error if file creation fails or if files exist and force is false.
func WriteKeyPair(pair *SSHKeyPair, outputPath string, force bool) error {
	pubPath := outputPath + ".pub"

	if !force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("private key file already exists: %s", outputPath)
		}
		if _, err := os.Stat(pubPath); err == nil {
			return fmt.Errorf("public key file already exists: %s", pubPath)
		}
	}

	if err := os.WriteFile(outputPath, pair.PrivateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	if err := os.WriteFile(pubPath, pair.PublicKey, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}
