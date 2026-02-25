package mnemonic

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.NoError(t, err)
			words := strings.Fields(m)
			assert.Equal(t, tt.wantWords, len(words))
			assert.NoError(t, Validate(m))
		})
	}
}

func TestGenerate_InvalidWordCount(t *testing.T) {
	invalidCounts := []int{0, 1, 11, 13, 14, 16, 17, 19, 20, 22, 23, 25, 100}

	for _, count := range invalidCounts {
		t.Run(fmt.Sprintf("invalid_%d", count), func(t *testing.T) {
			_, err := Generate(count)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid word count")
		})
	}
}

func TestGenerate_NegativeWordCount(t *testing.T) {
	_, err := Generate(-1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid word count")
}

func TestValidate_Valid(t *testing.T) {
	good := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	assert.NoError(t, Validate(good))
}

func TestValidate_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
	}{
		{
			name:     "bad checksum",
			mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
		},
		{
			name:     "garbage",
			mnemonic: "this is not a mnemonic at all",
		},
		{
			name:     "empty",
			mnemonic: "",
		},
		{
			name:     "single word",
			mnemonic: "abandon",
		},
		{
			name:     "wrong word count",
			mnemonic: "abandon abandon abandon abandon abandon abandon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.mnemonic)
			assert.Error(t, err)
		})
	}
}

func TestToSeed_EmptyPassphrase(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	expectedSeedHex := "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"

	seed := ToSeed(mnemonic, "")
	gotSeedHex := hex.EncodeToString(seed)

	assert.Equal(t, expectedSeedHex, gotSeedHex)
	assert.Equal(t, 64, len(seed))
}

func TestToSeed_WithPassphrase(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	seed1 := ToSeed(mnemonic, "")
	seed2 := ToSeed(mnemonic, "passphrase")
	seed3 := ToSeed(mnemonic, "passphrase")

	// Same mnemonic + passphrase should produce same seed
	assert.Equal(t, seed2, seed3)

	// Different passphrase should produce different seed
	assert.NotEqual(t, seed1, seed2)
}

func TestToSeed_DifferentPassphrases(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	tests := []string{"", "a", "longer passphrase", "TREZOR", "Correct Horse Battery Staple"}
	seeds := make(map[string]bool)

	for _, passphrase := range tests {
		seed := ToSeed(mnemonic, passphrase)
		seedHex := hex.EncodeToString(seed)
		// Each passphrase should produce a unique seed
		assert.False(t, seeds[seedHex], "duplicate seed for passphrase: %q", passphrase)
		seeds[seedHex] = true
	}
}

func TestGenerate_Unique(t *testing.T) {
	// Generate multiple mnemonics and verify they're unique
	mnemonics := make(map[string]bool)
	for i := 0; i < 10; i++ {
		m, err := Generate(24)
		require.NoError(t, err)
		assert.False(t, mnemonics[m], "generated duplicate mnemonic")
		mnemonics[m] = true
	}
}

func TestGenerate_AllWordCounts(t *testing.T) {
	wordCounts := []int{12, 15, 18, 21, 24}

	for _, count := range wordCounts {
		t.Run(fmt.Sprintf("%d_words", count), func(t *testing.T) {
			m, err := Generate(count)
			require.NoError(t, err)
			words := strings.Fields(m)
			assert.Equal(t, count, len(words))
			// Verify each word is lowercase
			for _, word := range words {
				assert.Equal(t, strings.ToLower(word), word)
			}
		})
	}
}
