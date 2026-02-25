// Package generate implements the generate command for creating BIP39 mnemonics.
package generate

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valerius/bip32-ssh-keygen/internal/mnemonic"
)

// Cmd returns the generate command.
// The generate command creates a new BIP39 mnemonic seed phrase with
// a configurable word count (12, 15, 18, 21, or 24 words).
func Cmd() *cobra.Command {
	var words int

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a new BIP39 mnemonic seed phrase",
		Long: `Generate creates a cryptographically secure BIP39 mnemonic seed phrase.

The mnemonic can be used with the 'derive' command to generate deterministic
SSH keys. Supported word counts are 12, 15, 18, 21, and 24 words.
Higher word counts provide more entropy and security.

The generated mnemonic should be written down and stored securely.
It cannot be recovered if lost.`,
		Example: `  # Generate a 24-word mnemonic (default)
  bip32-ssh-keygen generate

  # Generate a 12-word mnemonic
  bip32-ssh-keygen generate --words 12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			validCounts := []int{12, 15, 18, 21, 24}
			isValid := false
			for _, c := range validCounts {
				if words == c {
					isValid = true
					break
				}
			}

			if !isValid {
				return fmt.Errorf("invalid word count: %d. Supported: 12, 15, 18, 21, 24", words)
			}

			m, err := mnemonic.Generate(words)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), m)
			fmt.Fprintln(cmd.ErrOrStderr(), "\nIMPORTANT: Write down these words and store them securely. They cannot be recovered.")

			return nil
		},
	}

	cmd.Flags().IntVarP(&words, "words", "w", 24, "Number of words (12/15/18/21/24)")

	return cmd
}
