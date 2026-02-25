package mnemonic

import (
	"fmt"

	"github.com/blinklabs-io/go-bip39"
)

// Generate creates a new BIP39 mnemonic with the specified word count.
// Supported word counts: 12, 15, 18, 21, 24.
func Generate(wordCount int) (string, error) {
	var entropyBitSize int
	switch wordCount {
	case 12:
		entropyBitSize = 128
	case 15:
		entropyBitSize = 160
	case 18:
		entropyBitSize = 192
	case 21:
		entropyBitSize = 224
	case 24:
		entropyBitSize = 256
	default:
		return "", fmt.Errorf("invalid word count: %d. Supported: 12, 15, 18, 21, 24", wordCount)
	}

	entropy, err := bip39.NewEntropy(entropyBitSize)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// Validate checks if the mnemonic is valid according to BIP39.
func Validate(mnemonic string) error {
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("invalid mnemonic")
	}
	return nil
}

// ToSeed converts a mnemonic and passphrase into a 64-byte seed.
func ToSeed(mnemonic string, passphrase string) []byte {
	return bip39.NewSeed(mnemonic, passphrase)
}
