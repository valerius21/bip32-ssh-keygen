# bip32-ssh-keygen

Generate deterministic Ed25519 SSH keys from BIP39 seed phrases.

## Installation

```bash
go install github.com/valerius21/bip32-ssh-keygen@latest
```

## Usage

### Generate a new mnemonic

```bash
bip32-ssh-keygen generate
```

Creates a new BIP39 mnemonic seed phrase. Defaults to 24 words.

### Derive an SSH key from a mnemonic

```bash
bip32-ssh-keygen derive
```

Derives an Ed25519 SSH key from a BIP39 mnemonic. Prompts for the mnemonic interactively.

### Generate and derive in one step

```bash
bip32-ssh-keygen derive --generate --output ~/.ssh/id_ed25519
```

Generates a new mnemonic and immediately derives an SSH key from it.

### Interactive TUI

```bash
bip32-ssh-keygen tui
```

Launches an interactive terminal UI for generating mnemonics and deriving keys.

## How it works

This tool uses:

- **BIP39** for mnemonic generation and validation
- **SLIP-0010** for hierarchical deterministic key derivation (Ed25519-specific)
- **BIP44** derivation path conventions: `m/44'/22'/0'/0'`

The derivation path is hardened-only, as required by Ed25519.

## License

GPL-2.0
