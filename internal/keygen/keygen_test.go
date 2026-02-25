package keygen

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGenerate(t *testing.T) {
	// Known 32-byte seed
	seed := [32]byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}
	comment := "test@derivation"

	pair, err := Generate(seed, comment)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify private key PEM format
	if !bytes.HasPrefix(pair.PrivateKeyPEM, []byte("-----BEGIN OPENSSH PRIVATE KEY-----")) {
		t.Errorf("Private key PEM does not start with expected header")
	}

	// Verify public key format
	if !bytes.HasPrefix(pair.PublicKey, []byte("ssh-ed25519 ")) {
		t.Errorf("Public key does not start with ssh-ed25519")
	}

	// Verify public key contains comment
	if !bytes.Contains(pair.PublicKey, []byte(comment)) {
		t.Errorf("Public key does not contain comment: %s", comment)
	}

	// Round-trip verification: Parse the generated private key
	parsedKey, err := ssh.ParseRawPrivateKey(pair.PrivateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse generated private key: %v", err)
	}

	priv, ok := parsedKey.(*ed25519.PrivateKey)
	if !ok {
		t.Fatalf("Parsed key is not ed25519.PrivateKey")
	}

	// Verify the seed matches
	expectedPriv := ed25519.NewKeyFromSeed(seed[:])
	if !bytes.Equal(*priv, expectedPriv) {
		t.Errorf("Parsed private key does not match expected key from seed")
	}
}

func TestWriteKeyPair(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keygen-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	seed := [32]byte{0x42}
	pair, _ := Generate(seed, "test")
	outputPath := filepath.Join(tmpDir, "id_ed25519")

	// Test successful write
	err = WriteKeyPair(pair, outputPath, false)
	if err != nil {
		t.Fatalf("WriteKeyPair failed: %v", err)
	}

	// Verify private key permissions
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Private key file not found: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Private key has wrong permissions: %o, expected 0600", info.Mode().Perm())
	}

	// Verify public key permissions
	pubPath := outputPath + ".pub"
	info, err = os.Stat(pubPath)
	if err != nil {
		t.Fatalf("Public key file not found: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("Public key has wrong permissions: %o, expected 0644", info.Mode().Perm())
	}

	// Test refuse overwrite
	err = WriteKeyPair(pair, outputPath, false)
	if err == nil {
		t.Error("WriteKeyPair should have failed when overwriting without force")
	}

	// Test force overwrite
	err = WriteKeyPair(pair, outputPath, true)
	if err != nil {
		t.Errorf("WriteKeyPair failed with force=true: %v", err)
	}
}
