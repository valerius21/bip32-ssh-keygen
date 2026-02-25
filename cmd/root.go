// Package cmd provides the command-line interface for bip32-ssh-keygen.
// It implements commands for generating BIP39 mnemonics and deriving
// Ed25519 SSH keys using SLIP-0010 hierarchical deterministic key derivation.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/valerius21/bip32-ssh-keygen/cmd/derive"
	"github.com/valerius21/bip32-ssh-keygen/cmd/generate"
	"github.com/valerius21/bip32-ssh-keygen/cmd/tui"
)

// RootCmd is the root command for the bip32-ssh-keygen CLI.
// It serves as the entry point for all subcommands including generate,
// derive, and an interactive TUI mode.
var RootCmd = &cobra.Command{
	Use:     "bip32-ssh-keygen",
	Short:   "Generate deterministic Ed25519 SSH keys from BIP39 seed phrases",
	Long: `bip32-ssh-keygen generates deterministic Ed25519 SSH keys from BIP39 mnemonics.

It uses SLIP-0010 for hierarchical deterministic key derivation, which is
specifically designed for Ed25519 and only supports hardened derivation paths.

The default derivation path follows BIP44 conventions for SSH keys:
  m/44'/22'/0'/0'

Commands:
  generate    Create a new BIP39 mnemonic seed phrase
  derive      Derive an Ed25519 SSH key from a mnemonic
  tui         Launch an interactive terminal UI

Examples:
  # Generate a 24-word mnemonic
  bip32-ssh-keygen generate

  # Derive an SSH key (prompts for mnemonic)
  bip32-ssh-keygen derive

  # Launch interactive TUI
  bip32-ssh-keygen tui`,
	Version: "v0.2.0",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute runs the root command and handles any errors.
// It is the main entry point for the CLI application.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(generate.Cmd())
	RootCmd.AddCommand(derive.Cmd())
	RootCmd.AddCommand(tui.Cmd())
}
