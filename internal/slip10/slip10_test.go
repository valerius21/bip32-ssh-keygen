package slip10

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMasterKey(t *testing.T) {
	// Test vector 1 from SLIP-0010
	// Seed: 000102030405060708090a0b0c0d0e0f
	// Master: 2b4be7f19ee27bbf30c667b642d5f4aa69fd169872f8fc3059c08ebae2eb19e7
	// Chain:  90046a93de5380a72b5e45010748567d5ea02bbf6522f979e05c0d8d8ca9fffb
	seedHex := "000102030405060708090a0b0c0d0e0f"
	expectedChainHex := "90046a93de5380a72b5e45010748567d5ea02bbf6522f979e05c0d8d8ca9fffb"
	expectedPrivHex := "2b4be7f19ee27bbf30c667b642d5f4aa69fd169872f8fc3059c08ebae2eb19e7"

	seed, err := hex.DecodeString(seedHex)
	require.NoError(t, err)

	key, err := NewMasterKey(seed)
	require.NoError(t, err)

	assert.Equal(t, expectedChainHex, hex.EncodeToString(key.ChainCode[:]))
	assert.Equal(t, expectedPrivHex, hex.EncodeToString(key.PrivateKey[:]))
}

func TestNewMasterKey_DifferentSeeds(t *testing.T) {
	seed1 := []byte{0x00, 0x01, 0x02, 0x03}
	seed2 := []byte{0x00, 0x01, 0x02, 0x04}

	key1, err := NewMasterKey(seed1)
	require.NoError(t, err)

	key2, err := NewMasterKey(seed2)
	require.NoError(t, err)

	assert.NotEqual(t, key1.PrivateKey, key2.PrivateKey)
	assert.NotEqual(t, key1.ChainCode, key2.ChainCode)
}

func TestKey_DeriveChild(t *testing.T) {
	// Test vector 1, child m/0'
	seedHex := "000102030405060708090a0b0c0d0e0f"
	expectedChainHex := "8b59aa11380b624e81507a27fedda59fea6d0b779a778918a2fd3590e16e9c69"
	expectedPrivHex := "68e0fe46dfb67e368c75379acec591dad19df3cde26e63b93a8e704f1dade7a3"

	seed, err := hex.DecodeString(seedHex)
	require.NoError(t, err)

	master, err := NewMasterKey(seed)
	require.NoError(t, err)

	// m/0' is index 0x80000000
	child, err := master.DeriveChild(0x80000000)
	require.NoError(t, err)

	assert.Equal(t, expectedChainHex, hex.EncodeToString(child.ChainCode[:]))
	assert.Equal(t, expectedPrivHex, hex.EncodeToString(child.PrivateKey[:]))
}

func TestKey_DeriveChild_NonHardened(t *testing.T) {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	master, err := NewMasterKey(seed)
	require.NoError(t, err)

	// Try to derive non-hardened child (index 0)
	_, err = master.DeriveChild(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only supports hardened derivation")

	// Try index just below hardened threshold
	_, err = master.DeriveChild(0x7fffffff)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only supports hardened derivation")
}

func TestKey_DeriveChild_HardenedIndices(t *testing.T) {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	master, err := NewMasterKey(seed)
	require.NoError(t, err)

	// Test various hardened indices
	hardenedIndices := []uint32{
		0x80000000, // m/0'
		0x80000001, // m/1'
		0x8000002c, // m/44'
		0x80000016, // m/22'
		0xffffffff, // m/2147483647'
	}

	for _, index := range hardenedIndices {
		child, err := master.DeriveChild(index)
		require.NoError(t, err)
		assert.NotNil(t, child)
		// Child should have different key material
		assert.NotEqual(t, master.PrivateKey, child.PrivateKey)
		assert.NotEqual(t, master.ChainCode, child.ChainCode)
	}
}

func TestKey_DeriveChild_MultipleLevels(t *testing.T) {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	master, err := NewMasterKey(seed)
	require.NoError(t, err)

	// Derive m/0'/1'/2'
	child1, err := master.DeriveChild(0x80000000)
	require.NoError(t, err)

	child2, err := child1.DeriveChild(0x80000001)
	require.NoError(t, err)

	child3, err := child2.DeriveChild(0x80000002)
	require.NoError(t, err)

	// Each level should have different key material
	assert.NotEqual(t, master.PrivateKey, child1.PrivateKey)
	assert.NotEqual(t, child1.PrivateKey, child2.PrivateKey)
	assert.NotEqual(t, child2.PrivateKey, child3.PrivateKey)
}

func TestDeterminism(t *testing.T) {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")

	// Same seed should produce same master key
	master1, err := NewMasterKey(seed)
	require.NoError(t, err)

	master2, err := NewMasterKey(seed)
	require.NoError(t, err)

	assert.Equal(t, master1.PrivateKey, master2.PrivateKey)
	assert.Equal(t, master1.ChainCode, master2.ChainCode)

	// Same parent and index should produce same child
	child1, err := master1.DeriveChild(0x80000000)
	require.NoError(t, err)

	child2, err := master2.DeriveChild(0x80000000)
	require.NoError(t, err)

	assert.Equal(t, child1.PrivateKey, child2.PrivateKey)
	assert.Equal(t, child1.ChainCode, child2.ChainCode)
}

func TestKey_String(t *testing.T) {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	master, err := NewMasterKey(seed)
	require.NoError(t, err)

	str := master.String()
	assert.Contains(t, str, "PrivateKey:")
	assert.Contains(t, str, "ChainCode:")
	assert.True(t, strings.Contains(str, "2b4be7f19ee27bbf"))
}

func TestKey_DeriveChild_DifferentPaths(t *testing.T) {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	master, err := NewMasterKey(seed)
	require.NoError(t, err)

	// Derive two different paths: m/0' and m/1'
	child0, err := master.DeriveChild(0x80000000)
	require.NoError(t, err)

	child1, err := master.DeriveChild(0x80000001)
	require.NoError(t, err)

	// Should produce different keys
	assert.NotEqual(t, child0.PrivateKey, child1.PrivateKey)
	assert.NotEqual(t, child0.ChainCode, child1.ChainCode)
}
