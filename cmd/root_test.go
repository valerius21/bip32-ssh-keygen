package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	if RootCmd == nil {
		t.Fatal("RootCmd should not be nil")
	}

	if RootCmd.Use != "bip32-ssh-keygen" {
		t.Errorf("expected Use 'bip32-ssh-keygen', got '%s'", RootCmd.Use)
	}

	if RootCmd.Version != "v0.1.0" {
		t.Errorf("expected Version 'v0.1.0', got '%s'", RootCmd.Version)
	}
}
