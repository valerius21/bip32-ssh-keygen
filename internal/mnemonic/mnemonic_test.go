package mnemonic

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name      string
		wordCount int
		wantWords int
	}{
		{"Default 24 words", 24, 24},
		{"12 words", 12, 12},
		{"15 words", 15, 15},
		{"18 words", 18, 18},
		{"21 words", 21, 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := Generate(tt.wordCount)
			if err != nil {
				t.Fatalf("Generate(%d) error: %v", tt.wordCount, err)
			}
			words := strings.Fields(m)
			if len(words) != tt.wantWords {
				t.Errorf("Generate(%d) got %d words, want %d", tt.wordCount, len(words), tt.wantWords)
			}
			if err := Validate(m); err != nil {
				t.Errorf("Generated mnemonic is invalid: %v", err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Known good mnemonic from BIP39 test vectors
	good := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	if err := Validate(good); err != nil {
		t.Errorf("Validate(good) error: %v", err)
	}

	bad := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
	if err := Validate(bad); err == nil {
		t.Error("Validate(bad) expected error, got nil")
	}

	garbage := "this is not a mnemonic at all"
	if err := Validate(garbage); err == nil {
		t.Error("Validate(garbage) expected error, got nil")
	}
}

func TestToSeed(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	
	// Vector from BIP39: mnemonic + empty passphrase
	// Mnemonic: abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about
	// Seed: 5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4
	expectedSeedHex := "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"
	
	seed := ToSeed(mnemonic, "")
	gotSeedHex := hex.EncodeToString(seed)
	
	if gotSeedHex != expectedSeedHex {
		t.Errorf("ToSeed(mnemonic, \"\")\ngot  %s\nwant %s", gotSeedHex, expectedSeedHex)
	}
	
	// Verify passphrase changes seed
	seedWithPass := ToSeed(mnemonic, "TREZOR")
	if hex.EncodeToString(seedWithPass) == gotSeedHex {
		t.Error("ToSeed with passphrase should produce different seed")
	}
}
