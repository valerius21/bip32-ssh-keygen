package keygen

import (
	"crypto/ed25519"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// SSHKeyPair holds the generated SSH key pair data.
type SSHKeyPair struct {
	PrivateKeyPEM []byte
	PublicKey     []byte
	Fingerprint   string
}

// Generate creates an OpenSSH Ed25519 key pair from a 32-byte seed.
func Generate(seed [32]byte, comment string) (*SSHKeyPair, error) {
	// 1. ed25519.NewKeyFromSeed(seed[:]) -> private key
	privKey := ed25519.NewKeyFromSeed(seed[:])

	// 2. ssh.NewPublicKey(privateKey.Public()) -> SSH public key
	pubKey, err := ssh.NewPublicKey(privKey.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %w", err)
	}

	// 3. ssh.MarshalPrivateKey(privateKey, comment) -> PEM block
	pemBlock, err := ssh.MarshalPrivateKey(privKey, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	// 4. pem.EncodeToMemory(pemBlock) -> private key bytes
	privPEM := pem.EncodeToMemory(pemBlock)

	// 5. ssh.MarshalAuthorizedKey(pubKey) -> public key bytes
	pubBytes := ssh.MarshalAuthorizedKey(pubKey)

	// Add comment to public key if provided
	if comment != "" {
		// MarshalAuthorizedKey returns "type key\n", we want "type key comment\n"
		// Actually MarshalAuthorizedKey returns "type key\n"
		// We can just append the comment before the newline
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
func WriteKeyPair(pair *SSHKeyPair, outputPath string, force bool) error {
	pubPath := outputPath + ".pub"

	// Check if files exist
	if !force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("private key file already exists: %s", outputPath)
		}
		if _, err := os.Stat(pubPath); err == nil {
			return fmt.Errorf("public key file already exists: %s", pubPath)
		}
	}

	// Write private key with mode 0600
	if err := os.WriteFile(outputPath, pair.PrivateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Write public key with mode 0644
	if err := os.WriteFile(pubPath, pair.PublicKey, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}
