# bip32-ssh-keygen

Generate deterministic Ed25519 SSH keys from BIP39 seed phrases.

![Demo](demo.gif)

## Features

- **Deterministic**: Same mnemonic always produces the same SSH key
- **Secure**: Uses BIP39 + SLIP-0010 for cryptographically sound key derivation
- **Flexible**: CLI commands or interactive TUI
- **Recoverable**: Memorize your seed phrase, regenerate your SSH key anywhere

## Installation

### Go

```bash
go install github.com/valerius21/bip32-ssh-keygen@latest
```

### From Source

```bash
git clone https://github.com/valerius21/bip32-ssh-keygen.git
cd bip32-ssh-keygen
go build -o bip32-ssh-keygen .
```

## Usage

### Generate a new mnemonic

```bash
bip32-ssh-keygen generate
```

Creates a new BIP39 mnemonic seed phrase. Defaults to 24 words.

```bash
# Generate a 12-word mnemonic (faster to write down)
bip32-ssh-keygen generate --words 12
```

### Derive an SSH key from a mnemonic

```bash
bip32-ssh-keygen derive
```

Derives an Ed25519 SSH key from a BIP39 mnemonic. Prompts for the mnemonic interactively (hidden input).

```bash
# Pipe mnemonic from a file
cat mnemonic.txt | bip32-ssh-keygen derive

# Custom output path
bip32-ssh-keygen derive --output ~/.ssh/server_key

# Custom derivation path
bip32-ssh-keygen derive --path "m/44'/22'/1'/0'"
```

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

This tool combines several standards for deterministic key generation:

- **BIP39** for mnemonic generation and seed derivation
- **SLIP-0010** for hierarchical deterministic key derivation (Ed25519-specific)
- **BIP44** derivation path conventions: `m/44'/22'/0'/0'`

The derivation path uses hardened indices only, as required by SLIP-0010 for Ed25519 keys.

### Derivation Path

Default path: `m/44'/22'/0'/0'`

| Index | Purpose |
|-------|---------|
| `44'` | BIP44 purpose constant |
| `22'` | SSH key type identifier |
| `0'` | Account index |
| `0'` | Key index |

You can derive multiple keys from the same mnemonic by varying the account or key index:

```bash
# Different account
bip32-ssh-keygen derive --path "m/44'/22'/1'/0'"

# Different key within account
bip32-ssh-keygen derive --path "m/44'/22'/0'/1'"
```

## Security

- **Seed phrase**: Store your mnemonic securely. Anyone with access to it can derive your SSH keys.
- **Private keys**: Derived keys are written with `0600` permissions (private key) and `0644` (public key).
- **Memory**: Mnemonic input is hidden in the terminal and not logged.
- **No cloud**: Everything runs locally. No data leaves your machine.

## Why?

**Problem**: SSH keys are typically stored as files. If you lose access to your machine, you lose your keys and need to update all servers.

**Solution**: Use a BIP39 seed phrase as your "master key". Memorize it, write it down, or store it in a password manager. You can regenerate your SSH key on any machine at any time.

This is the same approach used by cryptocurrency wallets - now available for SSH authentication.

## License

GPL-2.0
