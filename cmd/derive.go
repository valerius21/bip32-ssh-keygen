package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"github.com/valerius/bip32-ssh-keygen/internal/keygen"
	"github.com/valerius/bip32-ssh-keygen/internal/mnemonic"
	"github.com/valerius/bip32-ssh-keygen/internal/path"
	"github.com/valerius/bip32-ssh-keygen/internal/slip10"


	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	derivePath       string
	derivePassphrase string
	deriveOutput     string
	deriveForce      bool
)

var deriveCmd = &cobra.Command{
	Use:   "derive",
	Short: "Derive an Ed25519 SSH key from a BIP39 mnemonic",
	RunE: func(cmd *cobra.Command, args []string) error {
		var input string
		var err error

		if term.IsTerminal(int(os.Stdin.Fd())) && cmd.InOrStdin() == os.Stdin {

			fmt.Print("Enter mnemonic: ")
			byteMnemonic, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println() // New line after hidden input
			if err != nil {
				return fmt.Errorf("failed to read mnemonic: %w", err)
			}
			input = string(byteMnemonic)
		} else {
			reader := bufio.NewReader(os.Stdin)
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

func init() {
	deriveCmd.Flags().StringVar(&derivePath, "path", path.DefaultPath, "derivation path")
	deriveCmd.Flags().StringVar(&derivePassphrase, "passphrase", "", "BIP39 passphrase")
	deriveCmd.Flags().StringVar(&deriveOutput, "output", "id_ed25519", "output file path")
	deriveCmd.Flags().BoolVar(&deriveForce, "force", false, "overwrite existing files")
}

func init() {
	RootCmd.AddCommand(deriveCmd)
}
