package keygen

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestGenerate(t *testing.T) {
	seed := [32]byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}
	comment := "test@derivation"

	pair, err := Generate(seed, comment)
	require.NoError(t, err)

	assert.True(t, bytes.HasPrefix(pair.PrivateKeyPEM, []byte("-----BEGIN OPENSSH PRIVATE KEY-----")), "private key should have OpenSSH header")
	assert.True(t, bytes.HasPrefix(pair.PublicKey, []byte("ssh-ed25519 ")), "public key should have ssh-ed25519 prefix")
	assert.Contains(t, string(pair.PublicKey), comment, "public key should contain comment")
	assert.NotEmpty(t, pair.Fingerprint, "fingerprint should not be empty")

	// Round-trip verification
	parsedKey, err := ssh.ParseRawPrivateKey(pair.PrivateKeyPEM)
	require.NoError(t, err)

	priv, ok := parsedKey.(*ed25519.PrivateKey)
	require.True(t, ok, "parsed key should be ed25519.PrivateKey")

	expectedPriv := ed25519.NewKeyFromSeed(seed[:])
	assert.True(t, bytes.Equal(*priv, expectedPriv), "parsed private key should match expected")
}

func TestGenerate_NoComment(t *testing.T) {
	seed := [32]byte{0x42}

	pair, err := Generate(seed, "")
	require.NoError(t, err)

	assert.True(t, bytes.HasPrefix(pair.PrivateKeyPEM, []byte("-----BEGIN OPENSSH PRIVATE KEY-----")), "private key should have OpenSSH header")
	assert.True(t, bytes.HasPrefix(pair.PublicKey, []byte("ssh-ed25519 ")), "public key should have ssh-ed25519 prefix")
	assert.NotEmpty(t, pair.Fingerprint, "fingerprint should not be empty")
}

func TestGenerate_Deterministic(t *testing.T) {
	seed := [32]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	comment := "test"

	pair1, err := Generate(seed, comment)
	require.NoError(t, err)

	pair2, err := Generate(seed, comment)
	require.NoError(t, err)

	// Both should generate the same key material
	assert.Equal(t, pair1.Fingerprint, pair2.Fingerprint)
}

func TestGenerate_DifferentSeeds(t *testing.T) {
	seed1 := [32]byte{0x01}
	seed2 := [32]byte{0x02}

	pair1, err := Generate(seed1, "")
	require.NoError(t, err)

	pair2, err := Generate(seed2, "")
	require.NoError(t, err)

	assert.NotEqual(t, pair1.Fingerprint, pair2.Fingerprint)
}

func TestWriteKeyPair(t *testing.T) {
	tmpDir := t.TempDir()

	seed := [32]byte{0x42}
	pair, err := Generate(seed, "test")
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "id_ed25519")

	// Test successful write
	err = WriteKeyPair(pair, outputPath, false)
	require.NoError(t, err)

	// Verify private key permissions
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "private key should have 0600 permissions")

	// Verify public key permissions
	pubPath := outputPath + ".pub"
	info, err = os.Stat(pubPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm(), "public key should have 0644 permissions")

	// Verify file contents
	privContent, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, pair.PrivateKeyPEM, privContent)

	pubContent, err := os.ReadFile(pubPath)
	require.NoError(t, err)
	assert.Equal(t, pair.PublicKey, pubContent)
}

func TestWriteKeyPair_RefuseOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	seed := [32]byte{0x42}
	pair, _ := Generate(seed, "test")
	outputPath := filepath.Join(tmpDir, "id_ed25519")

	// Write first time
	err := WriteKeyPair(pair, outputPath, false)
	require.NoError(t, err)

	// Try to write again without force
	err = WriteKeyPair(pair, outputPath, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestWriteKeyPair_ForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	seed := [32]byte{0x42}
	pair, _ := Generate(seed, "test")
	outputPath := filepath.Join(tmpDir, "id_ed25519")

	// Write first time
	err := WriteKeyPair(pair, outputPath, false)
	require.NoError(t, err)

	// Generate a different key to overwrite with
	seed2 := [32]byte{0x43}
	pair2, _ := Generate(seed2, "test")

	// Overwrite with force
	err = WriteKeyPair(pair2, outputPath, true)
	require.NoError(t, err)

	// Verify the new key was written
	privContent, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, pair2.PrivateKeyPEM, privContent)
}

func TestWriteKeyPair_ExistingPubOnly(t *testing.T) {
	tmpDir := t.TempDir()

	seed := [32]byte{0x42}
	pair, _ := Generate(seed, "test")
	outputPath := filepath.Join(tmpDir, "id_ed25519")

	// Create only the public key file
	os.WriteFile(outputPath+".pub", []byte("existing pub"), 0644)

	// Try to write without force
	err := WriteKeyPair(pair, outputPath, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestWriteKeyPair_ExistingPrivOnly(t *testing.T) {
	tmpDir := t.TempDir()

	seed := [32]byte{0x42}
	pair, _ := Generate(seed, "test")
	outputPath := filepath.Join(tmpDir, "id_ed25519")

	// Create only the private key file
	os.WriteFile(outputPath, []byte("existing priv"), 0600)

	// Try to write without force
	err := WriteKeyPair(pair, outputPath, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestSSHKeyPair_Fields(t *testing.T) {
	seed := [32]byte{0x42}
	pair, err := Generate(seed, "test")
	require.NoError(t, err)

	assert.NotNil(t, pair.PrivateKeyPEM)
	assert.NotNil(t, pair.PublicKey)
	assert.NotEmpty(t, pair.Fingerprint)

	// Verify fingerprint format (should be SHA256:...)
	assert.Contains(t, pair.Fingerprint, "SHA256:")
}
