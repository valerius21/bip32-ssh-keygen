package cmd

import (
	"fmt"
	"os"

	"github.com/valerius/bip32-ssh-keygen/internal/mnemonic"
	"github.com/spf13/cobra"
)

var words int

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new BIP39 mnemonic seed phrase",
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
		fmt.Fprintln(os.Stderr, "\nIMPORTANT: Write down these words and store them securely. They cannot be recovered.")

		return nil
	},
}

func init() {
	generateCmd.Flags().IntVarP(&words, "words", "w", 24, "number of words (12/15/18/21/24)")
	RootCmd.AddCommand(generateCmd)
}
