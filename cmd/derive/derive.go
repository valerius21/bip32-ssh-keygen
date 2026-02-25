// Package derive implements the derive command for generating SSH keys from BIP39 mnemonics.
package derive

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/valerius21/bip32-ssh-keygen/internal/keygen"
	"github.com/valerius21/bip32-ssh-keygen/internal/mnemonic"
	"github.com/valerius21/bip32-ssh-keygen/internal/path"
	"github.com/valerius21/bip32-ssh-keygen/internal/slip10"
	"golang.org/x/term"
)

// Cmd returns the derive command.
// The derive command creates an Ed25519 SSH key pair from a BIP39 mnemonic
// using SLIP-0010 hierarchical deterministic key derivation.
func Cmd() *cobra.Command {
	var (
		derivePath       string
		derivePassphrase string
		deriveOutput     string
		deriveForce      bool
		generateMnemonic bool
	)

	cmd := &cobra.Command{
		Use:   "derive",
		Short: "Derive an Ed25519 SSH key from a BIP39 mnemonic",
		Long: `Derive generates an Ed25519 SSH key pair from a BIP39 mnemonic using
SLIP-0010 hierarchical deterministic key derivation.

If no mnemonic is provided via stdin, a new mnemonic will be automatically
generated and displayed. The derived SSH key is written to the specified
output path with appropriate file permissions (0600 for private key,
0644 for public key).

The derivation path must use hardened indices (ending with ') as required
by SLIP-0010 for Ed25519. The default path follows BIP44 conventions for SSH keys.

Examples:
  # Derive with interactive mnemonic input (hidden)
  bip32-ssh-keygen derive

  # Derive from piped mnemonic
  echo "mnemonic words..." | bip32-ssh-keygen derive

  # Derive with custom path and output
  bip32-ssh-keygen derive --path "m/44'/22'/0'/1'" --output ~/.ssh/my_key

  # Auto-generate mnemonic and derive key
  bip32-ssh-keygen derive --generate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var input string
			var err error

			// Check if we should generate a new mnemonic
			if generateMnemonic {
				m, err := mnemonic.Generate(24)
				if err != nil {
					return fmt.Errorf("failed to generate mnemonic: %w", err)
				}
				input = m
				fmt.Fprintln(cmd.OutOrStdout(), "Generated mnemonic:")
				fmt.Fprintln(cmd.OutOrStdout(), m)
				fmt.Fprintln(cmd.OutOrStdout(), "\nIMPORTANT: Write down these words and store them securely.")
			} else if term.IsTerminal(int(os.Stdin.Fd())) && cmd.InOrStdin() == os.Stdin {
				// Interactive mode: prompt for mnemonic with hidden input
				fmt.Print("Enter mnemonic: ")
				byteMnemonic, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Println()
				if err != nil {
					return fmt.Errorf("failed to read mnemonic: %w", err)
				}
				input = string(byteMnemonic)
			} else {
				// Non-interactive mode: read from stdin
				reader := bufio.NewReader(cmd.InOrStdin())
				input, err = reader.ReadString('\n')
				if err != nil && err.Error() != "EOF" {
					return fmt.Errorf("failed to read mnemonic from stdin: %w", err)
				}
			}

			input = strings.TrimSpace(input)
			if input == "" {
				return fmt.Errorf("mnemonic cannot be empty")
			}

			if err := mnemonic.Validate(input); err != nil {
				return fmt.Errorf("invalid mnemonic: %w", err)
			}

			seed := mnemonic.ToSeed(input, derivePassphrase)

			indices, err := path.Parse(derivePath)
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}

			key, err := slip10.NewMasterKey(seed)
			if err != nil {
				return fmt.Errorf("failed to create master key: %w", err)
			}

			for _, index := range indices {
				key, err = key.DeriveChild(index)
				if err != nil {
					return fmt.Errorf("failed to derive child key: %w", err)
				}
			}

			pair, err := keygen.Generate(key.PrivateKey, derivePath)
			if err != nil {
				return fmt.Errorf("failed to generate SSH key: %w", err)
			}

			if err := keygen.WriteKeyPair(pair, deriveOutput, deriveForce); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "SSH key derived successfully\n")
			fmt.Fprintf(os.Stderr, "Fingerprint: %s\n", pair.Fingerprint)
			fmt.Fprintf(os.Stderr, "Path: %s\n", derivePath)
			fmt.Fprintf(os.Stderr, "Private key: %s\n", deriveOutput)
			fmt.Fprintf(os.Stderr, "Public key:  %s.pub\n", deriveOutput)

			return nil
		},
	}

	cmd.Flags().StringVar(&derivePath, "path", path.DefaultPath, "Derivation path (must use hardened indices)")
	cmd.Flags().StringVar(&derivePassphrase, "passphrase", "", "BIP39 passphrase (optional)")
	cmd.Flags().StringVar(&deriveOutput, "output", "id_ed25519", "Output file path for the SSH key")
	cmd.Flags().BoolVar(&deriveForce, "force", false, "Overwrite existing key files")
	cmd.Flags().BoolVarP(&generateMnemonic, "generate", "g", false, "Auto-generate a mnemonic if none provided")

	return cmd
}
