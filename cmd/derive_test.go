package cmd

import (
	"crypto/ed25519"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"

)

func TestDeriveCommand(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	tempDir, err := os.MkdirTemp("", "derive-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputBase := filepath.Join(tempDir, "id_ed25519")

	t.Run("Deterministic output", func(t *testing.T) {
		// Reset flags for each run
		deriveForce = true
		deriveOutput = outputBase
		derivePath = "m/44'/22'/0'/0'"
		derivePassphrase = ""

		// First derivation
		cmd1 := RootCmd
		cmd1.SetArgs([]string{"derive", "--output", outputBase, "--force"})
		cmd1.SetIn(strings.NewReader(mnemonic))
		if err := cmd1.Execute(); err != nil {
			t.Fatalf("first execution failed: %v", err)
		}

		priv1, _ := os.ReadFile(outputBase)
		pub1, _ := os.ReadFile(outputBase + ".pub")

		// Second derivation
		os.Remove(outputBase)
		os.Remove(outputBase + ".pub")

		// Reset flags again
		deriveForce = true
		deriveOutput = outputBase

		cmd2 := RootCmd
		cmd2.SetArgs([]string{"derive", "--output", outputBase, "--force"})
		cmd2.SetIn(strings.NewReader(mnemonic))
		if err := cmd2.Execute(); err != nil {
			t.Fatalf("second execution failed: %v", err)
		}

		priv2, _ := os.ReadFile(outputBase)
		pub2, _ := os.ReadFile(outputBase + ".pub")

		// Round-trip verification: Parse the generated private key
		parsed1, err := ssh.ParseRawPrivateKey(priv1)
		if err != nil {
			t.Fatalf("failed to parse priv1: %v", err)
		}
		parsed2, err := ssh.ParseRawPrivateKey(priv2)
		if err != nil {
			t.Fatalf("failed to parse priv2: %v", err)
		}

		k1 := parsed1.(*ed25519.PrivateKey)
		k2 := parsed2.(*ed25519.PrivateKey)

		if !bytes.Equal(*k1, *k2) {
			t.Error("functionally, private keys are not identical")
		}
		if !bytes.Equal(pub1, pub2) {
			t.Error("public keys are not identical")
		}
	})

	t.Run("Invalid mnemonic", func(t *testing.T) {
		deriveForce = true
		deriveOutput = filepath.Join(tempDir, "bad")
		cmd := RootCmd
		cmd.SetArgs([]string{"derive", "--output", deriveOutput, "--force"})
		cmd.SetIn(strings.NewReader("invalid words"))
		err := cmd.Execute()
		if err == nil {
			t.Error("expected error for invalid mnemonic, got nil")
		}
	})

	t.Run("Existing file without force", func(t *testing.T) {
		existingFile := filepath.Join(tempDir, "existing")
		os.WriteFile(existingFile, []byte("data"), 0600)

		deriveForce = false
		deriveOutput = existingFile
		cmd := RootCmd
		cmd.SetArgs([]string{"derive", "--output", existingFile})
		cmd.SetIn(strings.NewReader(mnemonic))
		err := cmd.Execute()
		if err == nil {
			t.Error("expected error when file exists and force is false, got nil")
		}
	})
}
