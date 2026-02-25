// Package mnemonic provides BIP39 mnemonic generation and validation.
//
// BIP39 defines a standard for creating mnemonic phrases that can be used
// to generate deterministic cryptocurrency wallets. This package implements
// the core BIP39 functionality for generating cryptographically secure
// mnemonics and converting them to seeds.
//
// The package supports standard BIP39 word counts:
//   - 12 words (128 bits of entropy)
//   - 15 words (160 bits of entropy)
//   - 18 words (192 bits of entropy)
//   - 21 words (224 bits of entropy)
//   - 24 words (256 bits of entropy)
package mnemonic

import (
	"fmt"

	"github.com/blinklabs-io/go-bip39"
)

// Generate creates a new BIP39 mnemonic with the specified word count.
//
// The wordCount parameter must be one of the supported values: 12, 15, 18, 21, or 24.
// Each word count corresponds to a specific entropy bit size:
//   - 12 words: 128 bits of entropy
//   - 15 words: 160 bits of entropy
//   - 18 words: 192 bits of entropy
//   - 21 words: 224 bits of entropy
//   - 24 words: 256 bits of entropy
//
// The function generates cryptographically secure random entropy and converts
// it to a mnemonic phrase using the BIP39 word list. The generated mnemonic
// includes a checksum for validation.
//
// Returns an error if the word count is not supported or if entropy generation fails.
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

// Validate checks if a mnemonic is valid according to BIP39.
//
// A valid mnemonic must:
//   - Contain only words from the BIP39 word list
//   - Have a word count of 12, 15, 18, 21, or 24
//   - Have a valid checksum
//
// Returns an error if the mnemonic is invalid or malformed.
func Validate(mnemonic string) error {
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("invalid mnemonic")
	}
	return nil
}

// ToSeed converts a mnemonic and optional passphrase into a 64-byte seed.
//
// The seed is generated using the PBKDF2 key derivation function with:
//   - SHA-512 as the hash function
//   - 2048 iterations
//   - The mnemonic as the password
//   - "mnemonic" + passphrase as the salt
//
// If no passphrase is provided, an empty string should be passed.
// The same mnemonic with different passphrases will produce different seeds.
func ToSeed(mnemonic string, passphrase string) []byte {
	return bip39.NewSeed(mnemonic, passphrase)
}
