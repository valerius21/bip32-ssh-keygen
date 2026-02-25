// Package path provides BIP44 derivation path parsing and formatting.
//
// BIP44 defines a hierarchical deterministic wallet structure with a standardized
// path format: m / purpose' / coin_type' / account' / change / address_index
//
// This package specifically handles derivation paths for Ed25519 SSH keys,
// which require hardened derivation (all indices must be hardened).
// The default path follows BIP44 conventions with coin type 22 (MONA) repurposed
// for SSH keys: m/44'/22'/0'/0'
package path

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultPath is the default BIP44 derivation path for SSH keys.
//
// The path m/44'/22'/0'/0' follows BIP44 conventions:
//   - 44': BIP44 purpose (hardened)
//   - 22': Coin type for SSH keys (hardened)
//   - 0': Account 0 (hardened)
//   - 0': External chain (hardened)
const DefaultPath = "m/44'/22'/0'/0'"

// HardenedBit is the bit mask that indicates a hardened derivation index.
//
// In BIP32/SLIP-0010, hardened indices are indicated by setting the
// most significant bit (0x80000000). When parsing paths, indices with
// the apostrophe suffix (e.g., 44') are converted to their hardened
// representation by ORing with this bit.
const HardenedBit uint32 = 0x80000000

// Parse parses a BIP44 derivation path string into a slice of uint32 indices.
//
// The path must follow the format "m/<index>/<index>/..." where each index
// is a non-negative integer optionally followed by an apostrophe (') to
// indicate hardened derivation.
//
// For Ed25519 SSH keys, all indices must be hardened (have the ' suffix).
// Non-hardened indices will result in an error.
//
// Examples of valid paths:
//   - "m/44'/22'/0'/0'" (default path)
//   - "m/44'/22'/0'/1'" (alternate key)
//   - "m/0'" (single hardened index)
//
// Returns an error if the path is empty, has an invalid prefix,
// contains non-hardened components, or has malformed indices.
func Parse(path string) ([]uint32, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	if !strings.HasPrefix(path, "m/") {
		return nil, fmt.Errorf("invalid path prefix: must start with 'm/'")
	}

	parts := strings.Split(path[2:], "/")
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		return nil, fmt.Errorf("invalid path: no components found")
	}

	indices := make([]uint32, len(parts))
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid path: empty component at index %d", i)
		}

		if !strings.HasSuffix(part, "'") {
			return nil, fmt.Errorf("non-hardened component '%s' at index %d: Ed25519 requires hardened derivation", part, i)
		}

		valStr := part[:len(part)-1]
		val, err := strconv.ParseUint(valStr, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid component '%s' at index %d: %w", part, i, err)
		}

		indices[i] = uint32(val) | HardenedBit
	}

	return indices, nil
}

// FormatPath converts a slice of uint32 indices back into a BIP44 path string.
//
// The returned path will have the format "m/<index>/<index>/..." where
// hardened indices (those with HardenedBit set) are displayed with an
// apostrophe suffix.
//
// This is the inverse of Parse. Given the output of Parse, FormatPath
// will reconstruct the original path string.
//
// Example:
//   indices := []uint32{0x8000002C, 0x80000016}
//   FormatPath(indices) // returns "m/44'/22'"
func FormatPath(indices []uint32) string {
	var sb strings.Builder
	sb.WriteString("m")
	for _, index := range indices {
		sb.WriteString("/")
		if index >= HardenedBit {
			sb.WriteString(strconv.FormatUint(uint64(index^HardenedBit), 10))
			sb.WriteString("'")
		} else {
			sb.WriteString(strconv.FormatUint(uint64(index), 10))
		}
	}
	return sb.String()
}
