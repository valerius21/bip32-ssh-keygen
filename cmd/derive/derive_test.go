package derive

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestDeriveCmd_DeterministicOutput(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	outputBase := filepath.Join(tempDir, "id_ed25519")

	// First derivation
	cmd1 := Cmd()
	cmd1.SetArgs([]string{"--output", outputBase, "--force"})
	cmd1.SetIn(strings.NewReader(mnemonic))
	buf1 := new(bytes.Buffer)
	cmd1.SetOut(buf1)
	cmd1.SetErr(buf1)
	require.NoError(t, cmd1.Execute())

	priv1, err := os.ReadFile(outputBase)
	require.NoError(t, err)
	pub1, err := os.ReadFile(outputBase + ".pub")
	require.NoError(t, err)

	// Second derivation
	cmd2 := Cmd()
	cmd2.SetArgs([]string{"--output", outputBase + "_2", "--force"})
	cmd2.SetIn(strings.NewReader(mnemonic))
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetErr(buf2)
	require.NoError(t, cmd2.Execute())

	priv2, err := os.ReadFile(outputBase + "_2")
	require.NoError(t, err)
	pub2, err := os.ReadFile(outputBase + "_2.pub")
	require.NoError(t, err)

	// Parse and compare keys
	parsed1, err := ssh.ParseRawPrivateKey(priv1)
	require.NoError(t, err)
	parsed2, err := ssh.ParseRawPrivateKey(priv2)
	require.NoError(t, err)

	k1 := parsed1.(*ed25519.PrivateKey)
	k2 := parsed2.(*ed25519.PrivateKey)

	assert.True(t, bytes.Equal(*k1, *k2), "private keys should be identical")
	assert.True(t, bytes.Equal(pub1, pub2), "public keys should be identical")
}

func TestDeriveCmd_InvalidMnemonic(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "bad")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath})
	cmd.SetIn(strings.NewReader("invalid words here"))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mnemonic")
}

func TestDeriveCmd_EmptyMnemonic(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "bad")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath})
	cmd.SetIn(strings.NewReader(""))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mnemonic cannot be empty")
}

func TestDeriveCmd_ExistingFileWithoutForce(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing")
	require.NoError(t, os.WriteFile(existingFile, []byte("data"), 0600))

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", existingFile})
	cmd.SetIn(strings.NewReader(mnemonic))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeriveCmd_ForceOverwrite(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "id_ed25519")
	require.NoError(t, os.WriteFile(outputPath, []byte("old data"), 0600))
	require.NoError(t, os.WriteFile(outputPath+".pub", []byte("old pub"), 0644))

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath, "--force"})
	cmd.SetIn(strings.NewReader(mnemonic))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify new key was written
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "BEGIN OPENSSH PRIVATE KEY")
}

func TestDeriveCmd_InvalidPath(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "id_ed25519")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath, "--path", "invalid/path", "--force"})
	cmd.SetIn(strings.NewReader(mnemonic))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path")
}

func TestDeriveCmd_NonHardenedPath(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "id_ed25519")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath, "--path", "m/44'/22'/0/0", "--force"})
	cmd.SetIn(strings.NewReader(mnemonic))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-hardened")
}

func TestDeriveCmd_CustomPath(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "id_ed25519")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath, "--path", "m/44'/22'/0'/1'", "--force"})
	cmd.SetIn(strings.NewReader(mnemonic))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify key was created
	_, err = os.Stat(outputPath)
	require.NoError(t, err)
	_, err = os.Stat(outputPath + ".pub")
	require.NoError(t, err)
}

func TestDeriveCmd_WithPassphrase(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	outputPath1 := filepath.Join(tempDir, "key1")
	outputPath2 := filepath.Join(tempDir, "key2")

	// Derive without passphrase
	cmd1 := Cmd()
	cmd1.SetArgs([]string{"--output", outputPath1, "--force"})
	cmd1.SetIn(strings.NewReader(mnemonic))
	buf1 := new(bytes.Buffer)
	cmd1.SetOut(buf1)
	cmd1.SetErr(buf1)
	require.NoError(t, cmd1.Execute())

	// Derive with passphrase
	cmd2 := Cmd()
	cmd2.SetArgs([]string{"--output", outputPath2, "--passphrase", "testpass", "--force"})
	cmd2.SetIn(strings.NewReader(mnemonic))
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetErr(buf2)
	require.NoError(t, cmd2.Execute())

	// Keys should be different
	pub1, _ := os.ReadFile(outputPath1 + ".pub")
	pub2, _ := os.ReadFile(outputPath2 + ".pub")
	assert.NotEqual(t, string(pub1), string(pub2))
}

func TestDeriveCmd_DefaultPath(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "id_ed25519")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath, "--force"})
	cmd.SetIn(strings.NewReader(mnemonic))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "m/44'/22'/0'/0'")
}

func TestDeriveCmd_OutputInfo(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "id_ed25519")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath, "--force"})
	cmd.SetIn(strings.NewReader(mnemonic))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "SSH key derived successfully")
	assert.Contains(t, output, "Fingerprint:")
	assert.Contains(t, output, "Private key:")
	assert.Contains(t, output, "Public key:")
}

func TestDeriveCmd_GenerateFlag(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "id_ed25519")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath, "--generate", "--force"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should have generated and displayed a mnemonic
	assert.Contains(t, output, "Generated mnemonic:")
	assert.Contains(t, output, "IMPORTANT")

	// Should have created the key
	_, err = os.Stat(outputPath)
	require.NoError(t, err)
}

func TestDeriveCmd_GenerateFlagShort(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "id_ed25519")

	cmd := Cmd()
	cmd.SetArgs([]string{"--output", outputPath, "-g", "--force"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	// Should have created the key
	_, err = os.Stat(outputPath)
	require.NoError(t, err)
}
