package path

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultPath is the default BIP44 derivation path for SSH keys.
const DefaultPath = "m/44'/22'/0'/0'"

// HardenedBit is the bit that indicates a hardened derivation.
const HardenedBit uint32 = 0x80000000

// Parse parses a BIP44 derivation path string into a slice of uint32 indices.
// It only supports hardened derivation paths as required by Ed25519 SLIP-0010.
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
