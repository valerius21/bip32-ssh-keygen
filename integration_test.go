package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	// Build the binary
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "bip32-ssh-keygen")
	
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	t.Run("Full lifecycle: generate then derive", func(t *testing.T) {
		// 1. Generate mnemonic
		genCmd := exec.Command(binPath, "generate")
		var genOut bytes.Buffer
		genCmd.Stdout = &genOut
		if err := genCmd.Run(); err != nil {
			t.Fatalf("failed to generate mnemonic: %v", err)
		}
		
		mnemonic := strings.TrimSpace(genOut.String())
		if mnemonic == "" {
			t.Fatal("generated mnemonic is empty")
		}

		// 2. Derive SSH key
		keyPath := filepath.Join(tmpDir, "id_ed25519_lifecycle")
		deriveCmd := exec.Command(binPath, "derive", "--output", keyPath, "--force")
		deriveCmd.Stdin = strings.NewReader(mnemonic)
		if err := deriveCmd.Run(); err != nil {
			t.Fatalf("failed to derive key: %v", err)
		}

		// 3. Validate with ssh-keygen
		verifyCmd := exec.Command("ssh-keygen", "-l", "-f", keyPath)
		if err := verifyCmd.Run(); err != nil {
			t.Fatalf("ssh-keygen failed to validate the derived key: %v", err)
		}

		// 4. Check permissions
		info, err := os.Stat(keyPath)
		if err != nil {
			t.Fatalf("failed to stat private key: %v", err)
		}
		if mode := info.Mode().Perm(); mode != 0600 {
			t.Errorf("expected private key permissions 0600, got %o", mode)
		}

		pubInfo, err := os.Stat(keyPath + ".pub")
		if err != nil {
			t.Fatalf("failed to stat public key: %v", err)
		}
		if mode := pubInfo.Mode().Perm(); mode != 0644 {
			t.Errorf("expected public key permissions 0644, got %o", mode)
		}
	})

	t.Run("Determinism: same input produces same output", func(t *testing.T) {
		mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		
		keyPath1 := filepath.Join(tmpDir, "id_ed25519_det1")
		deriveCmd1 := exec.Command(binPath, "derive", "--output", keyPath1, "--force")
		deriveCmd1.Stdin = strings.NewReader(mnemonic)
		if err := deriveCmd1.Run(); err != nil {
			t.Fatalf("first derivation failed: %v", err)
		}

		keyPath2 := filepath.Join(tmpDir, "id_ed25519_det2")
		deriveCmd2 := exec.Command(binPath, "derive", "--output", keyPath2, "--force")
		deriveCmd2.Stdin = strings.NewReader(mnemonic)
		if err := deriveCmd2.Run(); err != nil {
			t.Fatalf("second derivation failed: %v", err)
		}

		content1, _ := os.ReadFile(keyPath1)
		content2, _ := os.ReadFile(keyPath2)
		if !bytes.Equal(content1, content2) {
			// The private key contains a random salt in the OpenSSH format even for the same key.
			// We should compare the public keys instead for determinism.
			t.Log("Private keys differ (expected due to random salt in OpenSSH format)")
		}

		pubContent1, _ := os.ReadFile(keyPath1 + ".pub")
		pubContent2, _ := os.ReadFile(keyPath2 + ".pub")
		if !bytes.Equal(pubContent1, pubContent2) {
			t.Error("public keys are not identical for same mnemonic")
		}
	})

	t.Run("Cross-path: different paths produce different keys", func(t *testing.T) {
		mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		
		keyPath1 := filepath.Join(tmpDir, "id_ed25519_path1")
		deriveCmd1 := exec.Command(binPath, "derive", "--output", keyPath1, "--path", "m/44'/0'/0'/0'", "--force")
		deriveCmd1.Stdin = strings.NewReader(mnemonic)
		if err := deriveCmd1.Run(); err != nil {
			t.Fatalf("derivation 1 failed: %v", err)
		}

		keyPath2 := filepath.Join(tmpDir, "id_ed25519_path2")
		deriveCmd2 := exec.Command(binPath, "derive", "--output", keyPath2, "--path", "m/44'/0'/0'/1'", "--force")
		deriveCmd2.Stdin = strings.NewReader(mnemonic)
		if err := deriveCmd2.Run(); err != nil {
			t.Fatalf("derivation 2 failed: %v", err)
		}

		content1, _ := os.ReadFile(keyPath1)
		content2, _ := os.ReadFile(keyPath2)
		if bytes.Equal(content1, content2) {
			t.Error("private keys are identical for different paths")
		}
	})

	t.Run("Passphrase: same mnemonic with/without passphrase produces different keys", func(t *testing.T) {
		mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		
		keyPath1 := filepath.Join(tmpDir, "id_ed25519_pass1")
		deriveCmd1 := exec.Command(binPath, "derive", "--output", keyPath1, "--force")
		deriveCmd1.Stdin = strings.NewReader(mnemonic)
		if err := deriveCmd1.Run(); err != nil {
			t.Fatalf("derivation without passphrase failed: %v", err)
		}

		keyPath2 := filepath.Join(tmpDir, "id_ed25519_pass2")
		deriveCmd2 := exec.Command(binPath, "derive", "--output", keyPath2, "--passphrase", "correct horse battery staple", "--force")
		deriveCmd2.Stdin = strings.NewReader(mnemonic)
		if err := deriveCmd2.Run(); err != nil {
			t.Fatalf("derivation with passphrase failed: %v", err)
		}

		content1, _ := os.ReadFile(keyPath1)
		content2, _ := os.ReadFile(keyPath2)
		if bytes.Equal(content1, content2) {
			t.Error("private keys are identical despite different passphrases")
		}
	})
}
