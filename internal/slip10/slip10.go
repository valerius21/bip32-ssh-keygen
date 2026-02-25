// Package slip10 implements SLIP-0010 hierarchical deterministic key derivation for Ed25519.
//
// SLIP-0010 is a specification for hierarchical deterministic (HD) wallets that
// extends BIP-0032 to support Ed25519 curves. Unlike BIP-0032, SLIP-0010 only
// supports hardened key derivation for Ed25519, which provides stronger security
// guarantees by preventing public key derivation from parent public keys.
//
// This package implements the core SLIP-0010 operations:
//   - Master key generation from a seed
//   - Hardened child key derivation
//
// All derivation indices must be hardened (>= 0x80000000).
package slip10

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
)

// Key represents a SLIP-0010 Ed25519 HD key.
//
// A key consists of a 32-byte private key and a 32-byte chain code.
// The private key is used for cryptographic operations, while the chain
// code is used in the HMAC-SHA512 derivation process for child keys.
type Key struct {
	PrivateKey [32]byte
	ChainCode  [32]byte
}

const (
	curveSeed = "ed25519 seed"
	hardened  = 0x80000000
)

// NewMasterKey creates a new master key from a seed using SLIP-0010.
//
// The seed should be 64 bytes (typically the output of PBKDF2 from a BIP39
// mnemonic). The master key is derived using HMAC-SHA512 with the constant
// "ed25519 seed" as the key.
//
// The first 32 bytes of the HMAC output become the private key, and the
// last 32 bytes become the chain code.
//
// Returns an error if the HMAC computation fails.
func NewMasterKey(seed []byte) (*Key, error) {
	h := hmac.New(sha512.New, []byte(curveSeed))
	h.Write(seed)
	intermediate := h.Sum(nil)

	key := &Key{}
	copy(key.PrivateKey[:], intermediate[:32])
	copy(key.ChainCode[:], intermediate[32:])

	return key, nil
}

// DeriveChild derives a hardened child key from the parent key.
//
// SLIP-0010 for Ed25519 only supports hardened derivation (indices >= 0x80000000).
// Non-hardened indices will result in an error.
//
// The child key is derived using HMAC-SHA512 with the parent's chain code as
// the key. The data format is:
//   0x00 || parentPrivateKey || index(32-bit big-endian)
//
// The first 32 bytes of the HMAC output become the child private key, and
// the last 32 bytes become the child chain code.
//
// Returns an error if the index is not hardened or if HMAC computation fails.
func (k *Key) DeriveChild(index uint32) (*Key, error) {
	if index < hardened {
		return nil, errors.New("SLIP-0010 Ed25519 only supports hardened derivation (index >= 0x80000000)")
	}

	h := hmac.New(sha512.New, k.ChainCode[:])

	data := make([]byte, 1+32+4)
	data[0] = 0x00
	copy(data[1:33], k.PrivateKey[:])
	binary.BigEndian.PutUint32(data[33:], index)

	h.Write(data)
	intermediate := h.Sum(nil)

	child := &Key{}
	copy(child.PrivateKey[:], intermediate[:32])
	copy(child.ChainCode[:], intermediate[32:])

	return child, nil
}

// String returns a string representation of the key for debugging.
//
// The output format is "PrivateKey: <hex>, ChainCode: <hex>".
// This should not be used in production code as it exposes sensitive key material.
func (k *Key) String() string {
	return fmt.Sprintf("PrivateKey: %x, ChainCode: %x", k.PrivateKey, k.ChainCode)
}
