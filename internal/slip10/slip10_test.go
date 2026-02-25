package slip10

import (
	"encoding/hex"
	"testing"
)

func TestMasterKey(t *testing.T) {
	// Test vector 1 (Seed: 000102030405060708090a0b0c0d0e0f)
	// Master: 2b4be7f19ee27bbf30c667b642d5f4aa69fd169872f8fc3059c08ebae2eb19e7
	// Chain:  90046a93de5380a72b5e45010748567d5ea02bbf6522f979e05c0d8d8ca9fffb
	seedHex := "000102030405060708090a0b0c0d0e0f"
	expectedChainHex := "90046a93de5380a72b5e45010748567d5ea02bbf6522f979e05c0d8d8ca9fffb"
	expectedPrivHex := "2b4be7f19ee27bbf30c667b642d5f4aa69fd169872f8fc3059c08ebae2eb19e7"

	seed, _ := hex.DecodeString(seedHex)
	key, err := NewMasterKey(seed)
	if err != nil {
		t.Fatalf("Failed to create master key: %v", err)
	}

	if hex.EncodeToString(key.ChainCode[:]) != expectedChainHex {
		t.Errorf("Master chain code mismatch\nexpected: %s\ngot:      %s", expectedChainHex, hex.EncodeToString(key.ChainCode[:]))
	}
	if hex.EncodeToString(key.PrivateKey[:]) != expectedPrivHex {
		t.Errorf("Master private key mismatch\nexpected: %s\ngot:      %s", expectedPrivHex, hex.EncodeToString(key.PrivateKey[:]))
	}
}

func TestChildDerivation(t *testing.T) {
	// Test vector 1, child m/0'
	seedHex := "000102030405060708090a0b0c0d0e0f"
	expectedChainHex := "8b59aa11380b624e81507a27fedda59fea6d0b779a778918a2fd3590e16e9c69"
	expectedPrivHex := "68e0fe46dfb67e368c75379acec591dad19df3cde26e63b93a8e704f1dade7a3"

	seed, _ := hex.DecodeString(seedHex)
	master, _ := NewMasterKey(seed)

	// m/0' is index 0x80000000
	child, err := master.DeriveChild(0x80000000)
	if err != nil {
		t.Fatalf("Failed to derive child m/0': %v", err)
	}

	if hex.EncodeToString(child.ChainCode[:]) != expectedChainHex {
		t.Errorf("Child m/0' chain code mismatch\nexpected: %s\ngot:      %s", expectedChainHex, hex.EncodeToString(child.ChainCode[:]))
	}
	if hex.EncodeToString(child.PrivateKey[:]) != expectedPrivHex {
		t.Errorf("Child m/0' private key mismatch\nexpected: %s\ngot:      %s", expectedPrivHex, hex.EncodeToString(child.PrivateKey[:]))
	}
}

func TestRejectNonHardened(t *testing.T) {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	master, _ := NewMasterKey(seed)

	_, err := master.DeriveChild(0)
	if err == nil {
		t.Error("Expected error for non-hardened index 0, got nil")
	}

	_, err = master.DeriveChild(0x7fffffff)
	if err == nil {
		t.Error("Expected error for non-hardened index 0x7fffffff, got nil")
	}
}

func TestDeterminism(t *testing.T) {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	master1, _ := NewMasterKey(seed)
	master2, _ := NewMasterKey(seed)

	if master1.PrivateKey != master2.PrivateKey || master1.ChainCode != master2.ChainCode {
		t.Error("Master keys from same seed are not identical")
	}

	child1, _ := master1.DeriveChild(0x80000000)
	child2, _ := master2.DeriveChild(0x80000000)

	if child1.PrivateKey != child2.PrivateKey || child1.ChainCode != child2.ChainCode {
		t.Error("Child keys from same parent and index are not identical")
	}
}
