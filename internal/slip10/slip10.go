package slip10

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
)

// Key represents a SLIP-0010 Ed25519 HD key.
type Key struct {
	PrivateKey [32]byte
	ChainCode  [32]byte
}

const (
	curveSeed = "ed25519 seed"
	hardened  = 0x80000000
)

// NewMasterKey creates a new master key from a seed using SLIP-0010.
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
// SLIP-0010 for Ed25519 only supports hardened derivation.
func (k *Key) DeriveChild(index uint32) (*Key, error) {
	if index < hardened {
		return nil, errors.New("SLIP-0010 Ed25519 only supports hardened derivation (index >= 0x80000000)")
	}

	h := hmac.New(sha512.New, k.ChainCode[:])

	// Data format: 0x00 || parentPrivKey || index(32-bit BE)
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

// String returns a string representation of the key (for debugging).
func (k *Key) String() string {
	return fmt.Sprintf("PrivateKey: %x, ChainCode: %x", k.PrivateKey, k.ChainCode)
}
