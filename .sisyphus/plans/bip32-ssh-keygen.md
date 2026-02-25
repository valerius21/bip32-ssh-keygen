# BIP32 SSH Keygen

## TL;DR

> **Quick Summary**: Go CLI tool that generates BIP39 mnemonics and derives deterministic Ed25519 SSH keys via SLIP-0010. Two subcommands: `generate` (new mnemonic) and `derive` (mnemonic → SSH key pair).
> 
> **Deliverables**:
> - `bip32-ssh-keygen` binary with Cobra CLI
> - `generate` subcommand — creates new BIP39 mnemonic phrases
> - `derive` subcommand — deterministic Ed25519 SSH key from mnemonic
> - devenv.nix for reproducible development environment
> - Full TDD test suite with SLIP-0010 test vectors
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 4 waves + final verification
> **Critical Path**: T2 → T3/T6 → T8 → T9

---

## Context

### Original Request
Build a Go application that uses a BIP32 seed phrase to generate deterministic Ed25519 SSH keys, with devenv for the development environment.

### Interview Summary
**Key Discussions**:
- **Language & Tooling**: Go with devenv (Nix) for reproducible dev shell
- **CLI Framework**: Cobra with subcommands (`generate`, `derive`)
- **Input Method**: Auto-detect TTY — interactive hidden prompt if terminal, read stdin if piped
- **Output**: Both private key + .pub file with ssh-keygen-style permissions (0600/0644)
- **Derivation Path**: Configurable `--path` flag, default `m/44'/22'/0'/0'` (22 = SSH port, memorable, unused in SLIP-0044)
- **BIP39 Passphrase**: Optional `--passphrase` flag for extra seed entropy
- **Multiple Keys**: Single key per invocation, index 0 default
- **Test Strategy**: TDD with `go test`, SLIP-0010 test vectors for crypto correctness

**Research Findings**:
- **MUST use SLIP-0010** for Ed25519 HD key derivation (BIP32 is secp256k1 only)
- SLIP-0010 uses `"ed25519 seed"` HMAC salt (not `"Bitcoin seed"`)
- Only **hardened** child derivation is supported for Ed25519
- `ssh.MarshalPrivateKey()` produces correct OpenSSH format (required for Ed25519)
- `ed25519.NewKeyFromSeed()` panics on non-32-byte input — must validate
- Reference implementations: `pinpox/mnemonic-ssh` (Go), `mikalv/anything2ed25519` (Go), `mikalv/hdpki` (Rust)

### Metis Review
**Identified Gaps** (addressed):
- **`tyler-smith/go-bip39` is deleted** → Replaced with `blinklabs-io/go-bip39` (maintained fork, v0.2.0, multi-language support)
- **`lyonnee/key25519` is a 4-star single-author crypto lib** → Replaced with `anyproto/go-slip10` (Metis-verified line-by-line)
- **BIP39 passphrase ≠ SSH key file encryption** → Clarified: `--passphrase` is BIP39 seed derivation only, NOT file encryption
- **No file collision handling** → Added: refuse overwrite by default, `--force` flag to overwrite
- **No public key comment convention** → Added: use derivation path as comment (e.g., `ssh-ed25519 AAAA... m/44'/22'/0'/0'`)
- **No word count configurability** → Added: `--words` flag for generate (12/15/18/21/24, default 24)

---

## Work Objectives

### Core Objective
Build a Go CLI tool that deterministically generates Ed25519 SSH key pairs from BIP39 mnemonic seed phrases using SLIP-0010 hierarchical derivation, with a secondary command to generate new mnemonics.

### Concrete Deliverables
- `bip32-ssh-keygen generate` — outputs a new BIP39 mnemonic to stdout
- `bip32-ssh-keygen derive` — reads mnemonic, writes OpenSSH Ed25519 key pair to files
- `devenv.nix` + `devenv.yaml` — reproducible Go dev environment
- Test suite with SLIP-0010 test vectors and BIP39 test vectors
- Go module with clean `internal/` package structure

### Definition of Done
- [ ] `bip32-ssh-keygen generate` produces valid 24-word BIP39 mnemonic
- [ ] `bip32-ssh-keygen derive` produces valid Ed25519 SSH key pair from mnemonic
- [ ] Same mnemonic + same path always produces identical key (determinism)
- [ ] Generated keys are accepted by `ssh-keygen -l -f` (format validation)
- [ ] All tests pass: `go test ./...`
- [ ] Build succeeds: `go build -o bip32-ssh-keygen .`

### Must Have
- SLIP-0010 derivation (NOT BIP32) with "ed25519 seed" HMAC salt
- Hardened-only derivation for all path components
- BIP39 mnemonic validation before derivation
- TTY auto-detection for interactive vs piped mnemonic input
- File permissions: 0600 private key, 0644 public key
- `--force` flag to overwrite existing key files
- Derivation path in public key comment
- SLIP-0010 test vectors for crypto correctness verification
- `ed25519.NewKeyFromSeed()` input validation (exactly 32 bytes)

### Must NOT Have (Guardrails)
- **No SSH agent integration** — out of scope for v1
- **No key import/export/backup** — just generation and derivation
- **No GUI** — CLI only
- **No config file** — flags only
- **No batch/multi-key derivation** — single key per invocation
- **No file encryption of output keys** — BIP39 passphrase is for seed derivation, not file encryption
- **No non-hardened derivation** — Ed25519 SLIP-0010 requires all-hardened
- **No BIP32 (secp256k1)** — SLIP-0010 only
- **No `FactomProject/go-bip32`** — secp256k1 only, wrong curve
- **No `lyonnee/key25519`** — insufficient trust (4-star single-author crypto lib)
- **No `tyler-smith/go-bip39`** — repository deleted

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: NO (greenfield)
- **Automated tests**: TDD (test-first)
- **Framework**: `go test` (stdlib)
- **TDD Flow**: Each core module follows RED (failing test) → GREEN (minimal impl) → REFACTOR

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Crypto modules**: Use Bash (`go test -v -run`) — run specific test cases, compare against known test vectors
- **CLI commands**: Use interactive_bash (tmux) — run binary, provide input, validate output
- **File output**: Use Bash — check file existence, permissions, format with `ssh-keygen -l -f`

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation — independent scaffolding):
├── Task 1: devenv setup (devenv.nix, devenv.yaml) [quick]
├── Task 2: Go module + Cobra scaffold + .gitignore [quick]

Wave 2 (Core modules — depend on T2, parallel with each other):
├── Task 3: SLIP-0010 implementation + TDD tests (internal/slip10/) [deep]
├── Task 4: BIP44 path parser + TDD tests (internal/path/) [quick]
├── Task 5: BIP39 mnemonic wrapper + TDD tests (internal/mnemonic/) [quick]
├── Task 6: SSH key generation + TDD tests (internal/keygen/) [deep]

Wave 3 (CLI commands — depend on Wave 2 modules):
├── Task 7: `generate` subcommand (cmd/generate.go) [quick]
├── Task 8: `derive` subcommand (cmd/derive.go) [deep]

Wave 4 (Integration — depends on Wave 3):
├── Task 9: End-to-end integration test + build verification [deep]

Wave FINAL (After ALL — 4 parallel reviewers):
├── Task F1: Plan compliance audit [oracle]
├── Task F2: Code quality review [unspecified-high]
├── Task F3: Real manual QA [unspecified-high]
├── Task F4: Scope fidelity check [deep]

Critical Path: T2 → T3 → T8 → T9 → F1-F4
Parallel Speedup: ~60% faster than sequential
Max Concurrent: 4 (Wave 2)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| T1   | —         | —      | 1    |
| T2   | —         | T3-T8  | 1    |
| T3   | T2        | T8, T9 | 2    |
| T4   | T2        | T8     | 2    |
| T5   | T2        | T7, T8 | 2    |
| T6   | T2        | T8, T9 | 2    |
| T7   | T2, T5    | T9     | 3    |
| T8   | T2-T6     | T9     | 3    |
| T9   | T7, T8    | F1-F4  | 4    |
| F1-F4| T9        | —      | FINAL|

### Agent Dispatch Summary

- **Wave 1**: **2** — T1 → `quick`, T2 → `quick`
- **Wave 2**: **4** — T3 → `deep`, T4 → `quick`, T5 → `quick`, T6 → `deep`
- **Wave 3**: **2** — T7 → `quick`, T8 → `deep`
- **Wave 4**: **1** — T9 → `deep`
- **FINAL**: **4** — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [ ] 1. devenv Setup

  **What to do**:
  - Create `devenv.yaml` pointing to devenv.nix
  - Create `devenv.nix` with:
    - Go toolchain (latest stable)
    - `gopls` for LSP
    - `golangci-lint` for linting
    - `delve` for debugging (optional, nice-to-have)
  - Verify `devenv shell` activates and `go version` works

  **Must NOT do**:
  - Do not install project Go dependencies (that's T2)
  - Do not create go.mod or any Go source files

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple config files, no complex logic
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - None relevant for Nix config

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: None (devenv is for developer convenience, not build dependency)
  - **Blocked By**: None (can start immediately)

  **References**:

  **External References**:
  - devenv docs: https://devenv.sh/languages/go/ — Go language support configuration
  - devenv.yaml format: https://devenv.sh/getting-started/ — Entry point config

  **WHY Each Reference Matters**:
  - The devenv Go docs show the exact `languages.go.enable = true` syntax and available options

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: devenv shell provides Go toolchain
    Tool: Bash
    Preconditions: devenv.nix and devenv.yaml exist in project root
    Steps:
      1. Run `devenv shell -- go version`
      2. Assert output contains `go1.` (any recent version)
      3. Run `devenv shell -- which gopls`
      4. Assert gopls binary path is returned (non-empty)
    Expected Result: Go and gopls available in devenv shell
    Failure Indicators: "command not found" or non-zero exit code
    Evidence: .sisyphus/evidence/task-1-devenv-shell.txt
  ```

  **Commit**: YES
  - Message: `chore: add devenv configuration for reproducible dev environment`
  - Files: `devenv.nix`, `devenv.yaml`
  - Pre-commit: `devenv shell -- go version`

- [ ] 2. Go Module + Cobra Scaffold

  **What to do**:
  - TDD: Write a test that imports the root command and verifies it exists
  - Run `go mod init github.com/user/bip32-ssh-keygen` (or appropriate module path)
  - Create `main.go` — thin entry point calling `cmd.Execute()`
  - Create `cmd/root.go` — Cobra root command with:
    - Name: `bip32-ssh-keygen`
    - Short description: "Generate deterministic Ed25519 SSH keys from BIP39 seed phrases"
    - Version flag (hardcode `v0.1.0` for now)
  - Create `.gitignore` with Go-standard ignores (binary, vendor, .env, etc.)
  - Run `go mod tidy` to resolve dependencies
  - Verify `go build .` produces a binary and `./bip32-ssh-keygen --help` works

  **Must NOT do**:
  - Do not implement `generate` or `derive` subcommands (those are T7, T8)
  - Do not add any internal packages yet
  - Do not add dependencies beyond cobra

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard Go project init with Cobra boilerplate
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: T3, T4, T5, T6, T7, T8 (all need go.mod)
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `pinpox/mnemonic-ssh` main.go — Reference for minimal Go SSH keygen entry point structure
  - `openpubkey/opkssh` main.go — Production Cobra CLI scaffold pattern (root + subcommands)

  **External References**:
  - Cobra user guide: https://github.com/spf13/cobra/blob/main/site/content/user_guide.md — Command creation patterns

  **WHY Each Reference Matters**:
  - opkssh shows production-grade Cobra setup: root command with Version, subcommand registration, thin main.go
  - The Cobra guide shows idiomatic `cmd/root.go` + `main.go` separation

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Binary builds and shows help
    Tool: Bash
    Preconditions: go.mod exists, cmd/root.go has root command
    Steps:
      1. Run `go build -o bip32-ssh-keygen .`
      2. Assert exit code 0 and binary file exists
      3. Run `./bip32-ssh-keygen --help`
      4. Assert output contains "Generate deterministic Ed25519 SSH keys"
      5. Run `./bip32-ssh-keygen --version`
      6. Assert output contains "v0.1.0"
    Expected Result: Binary builds, help text shown, version displayed
    Failure Indicators: Build errors, missing help text, wrong version
    Evidence: .sisyphus/evidence/task-2-build-help.txt

  Scenario: go test passes for root command
    Tool: Bash
    Preconditions: Test file exists for root command
    Steps:
      1. Run `go test ./cmd/... -v`
      2. Assert all tests pass
    Expected Result: Root command test passes
    Failure Indicators: Test failures or compilation errors
    Evidence: .sisyphus/evidence/task-2-root-test.txt
  ```

  **Commit**: YES
  - Message: `chore: scaffold Go module with Cobra CLI skeleton`
  - Files: `go.mod`, `go.sum`, `main.go`, `cmd/root.go`, `cmd/root_test.go`, `.gitignore`
  - Pre-commit: `go build . && go test ./...`

---

- [ ] 3. SLIP-0010 Ed25519 HD Key Derivation (TDD)

  **What to do**:
  - Create `internal/slip10/slip10.go` and `internal/slip10/slip10_test.go`
  - **RED**: Write tests FIRST using SLIP-0010 test vectors from the spec:
    - Test vector 1 (seed `000102030405060708090a0b0c0d0e0f`):
      - Master key: chain `873dff81c02f525623fd1fe5167eac3a55a049de3d314bb42ee227ffed37d508`
      - Master private: `2b4be7f19ee27bbf30c667b642d5f4aa69fd169872f8fc3059c08ebae2eb19e7`
      - Child m/0': chain `0b78a3226f915c082bf118f83618a618ab6dec793752624cbeb622acb562862d`
      - Child m/0': private `68e0fe46dfb67e368c75379acec591dad19df3cde26e63b93a8e45a1d7c7ad29`
    - Test determinism: same seed + same path → same key every time
    - Test hardened-only enforcement: reject non-hardened indices
  - **GREEN**: Implement `slip10.go`:
    - `type Key struct { PrivateKey [32]byte; ChainCode [32]byte }`
    - `func NewMasterKey(seed []byte) (*Key, error)` — HMAC-SHA512 with `"ed25519 seed"` salt
    - `func (k *Key) DeriveChild(index uint32) (*Key, error)` — hardened child derivation
    - Hardened index = `index | 0x80000000` (set bit 31)
    - Data format: `0x00 || parentPrivKey || index(32-bit BE)`
    - Validate: reject if `index < 0x80000000` (non-hardened)
  - **REFACTOR**: Clean up, ensure idiomatic Go error handling

  **Must NOT do**:
  - Do not implement BIP44 path parsing (that's T4)
  - Do not import any external crypto libraries — use only `crypto/hmac`, `crypto/sha512` from stdlib
  - Do not implement public key derivation (impossible for Ed25519 SLIP-0010)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Cryptographic implementation requires precision — test vectors must match exactly
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 6)
  - **Blocks**: T8 (derive command needs SLIP-0010), T9 (integration tests)
  - **Blocked By**: T2 (needs go.mod)

  **References**:

  **External References**:
  - SLIP-0010 spec: https://slips.readthedocs.io/en/latest/slip-0010/ — Master key generation algorithm, child derivation algorithm, test vectors
  - `anyproto/go-slip10` source: https://github.com/anyproto/go-slip10 — Verified-correct Go implementation to cross-reference (Metis reviewed line-by-line)
  - `lyonnee/key25519/bip32/bip32.go` lines 22-29 — Alternative implementation showing GenerateMasterKey with "ed25519 seed" salt

  **WHY Each Reference Matters**:
  - The SLIP-0010 spec has exact test vectors we MUST match byte-for-byte for correctness
  - anyproto/go-slip10 is the reference implementation to cross-check our logic against
  - key25519 shows a minimal implementation of the same algorithm for comparison

  **Acceptance Criteria**:
  - [ ] `go test ./internal/slip10/ -v` passes all SLIP-0010 test vectors
  - [ ] Master key from test vector seed matches expected hex values
  - [ ] Derived child keys match expected hex values at each depth

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: SLIP-0010 test vector 1 — master key derivation
    Tool: Bash
    Preconditions: internal/slip10/ package exists with tests
    Steps:
      1. Run `go test ./internal/slip10/ -v -run TestMasterKey`
      2. Assert test passes with exit code 0
      3. Verify output contains "PASS"
    Expected Result: Master key matches SLIP-0010 test vector 1
    Failure Indicators: Hex mismatch in assertion, test FAIL
    Evidence: .sisyphus/evidence/task-3-slip10-master.txt

  Scenario: SLIP-0010 test vector 1 — child derivation m/0'
    Tool: Bash
    Preconditions: DeriveChild implemented
    Steps:
      1. Run `go test ./internal/slip10/ -v -run TestChildDerivation`
      2. Assert derived key at m/0' matches expected hex
    Expected Result: Child key matches SLIP-0010 spec
    Failure Indicators: Wrong child key bytes
    Evidence: .sisyphus/evidence/task-3-slip10-child.txt

  Scenario: Non-hardened index rejected
    Tool: Bash
    Preconditions: DeriveChild validates index
    Steps:
      1. Run `go test ./internal/slip10/ -v -run TestRejectNonHardened`
      2. Assert error returned for index < 0x80000000
    Expected Result: Error with descriptive message about hardened-only requirement
    Failure Indicators: No error returned, or wrong error
    Evidence: .sisyphus/evidence/task-3-slip10-reject.txt
  ```

  **Commit**: YES
  - Message: `feat(slip10): implement SLIP-0010 Ed25519 HD key derivation with test vectors`
  - Files: `internal/slip10/slip10.go`, `internal/slip10/slip10_test.go`
  - Pre-commit: `go test ./internal/slip10/...`

- [ ] 4. BIP44 Derivation Path Parser (TDD)

  **What to do**:
  - Create `internal/path/path.go` and `internal/path/path_test.go`
  - **RED**: Write tests FIRST:
    - Parse `m/44'/22'/0'/0'` → `[]uint32{0x8000002C, 0x80000016, 0x80000000, 0x80000000}`
    - Parse `m/44'/22'/1'/0'` → correct hardened indices
    - Reject invalid paths: `""`, `"m"`, `"m/"`, `"abc"`, `"m/44/22"` (non-hardened should warn or error)
    - Reject paths with non-hardened components (Ed25519 requires all-hardened)
    - Parse single-component: `m/0'` → `[]uint32{0x80000000}`
  - **GREEN**: Implement:
    - `func Parse(path string) ([]uint32, error)` — parses BIP44 path string
    - `const DefaultPath = "m/44'/22'/0'/0'"` — default derivation path
    - Strip leading `m/`, split by `/`, for each component: strip trailing `'`, parse uint32, OR with 0x80000000
    - Validate all components are hardened (have `'` suffix)
    - `func FormatPath(indices []uint32) string` — reverse: indices back to string (for public key comment)
  - **REFACTOR**: Ensure clear error messages for each failure mode

  **Must NOT do**:
  - Do not support non-hardened derivation (Ed25519 requires all-hardened)
  - Do not integrate with SLIP-0010 (that coupling is in T8)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: String parsing with clear test cases, no crypto complexity
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 5, 6)
  - **Blocks**: T8 (derive command needs path parsing)
  - **Blocked By**: T2 (needs go.mod)

  **References**:

  **External References**:
  - BIP44 spec: https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki — Path format definition
  - SLIP-0010 hardened-only requirement: https://slips.readthedocs.io/en/latest/slip-0010/ — Why all components must be hardened

  **Pattern References**:
  - `volodymyrprokopyuk/go-wallet` hdwallet.go:83-98 — BIP32 path parsing with regex validation pattern

  **WHY Each Reference Matters**:
  - BIP44 spec defines the canonical `m/purpose'/coin'/account'/change/index` format
  - go-wallet shows regex-based path validation pattern in Go

  **Acceptance Criteria**:
  - [ ] `go test ./internal/path/ -v` passes all path parsing tests
  - [ ] Default path constant is `m/44'/22'/0'/0'`

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Parse default derivation path
    Tool: Bash
    Preconditions: internal/path/ package exists
    Steps:
      1. Run `go test ./internal/path/ -v -run TestParseDefault`
      2. Assert m/44'/22'/0'/0' parses to correct uint32 slice with hardened bits set
    Expected Result: [0x8000002C, 0x80000016, 0x80000000, 0x80000000]
    Failure Indicators: Wrong indices or missing hardened bit
    Evidence: .sisyphus/evidence/task-4-parse-default.txt

  Scenario: Reject non-hardened path components
    Tool: Bash
    Preconditions: Parse validates hardened-only
    Steps:
      1. Run `go test ./internal/path/ -v -run TestRejectNonHardened`
      2. Assert paths like "m/44'/22'/0/0" return error
    Expected Result: Error message indicating Ed25519 requires hardened derivation
    Failure Indicators: Non-hardened path accepted without error
    Evidence: .sisyphus/evidence/task-4-reject-nonhardened.txt
  ```

  **Commit**: YES
  - Message: `feat(path): implement BIP44 derivation path parser`
  - Files: `internal/path/path.go`, `internal/path/path_test.go`
  - Pre-commit: `go test ./internal/path/...`

---

- [ ] 5. BIP39 Mnemonic Wrapper (TDD)

  **What to do**:
  - Create `internal/mnemonic/mnemonic.go` and `internal/mnemonic/mnemonic_test.go`
  - Run `go get github.com/blinklabs-io/go-bip39` to add dependency
  - **RED**: Write tests FIRST:
    - Generate mnemonic: verify output is exactly 24 space-separated words (default)
    - Generate with --words 12: verify 12 words
    - Validate known-good mnemonic: returns true
    - Validate garbage string: returns false with descriptive error
    - Validate wrong checksum: returns false
    - Seed derivation: known mnemonic + empty passphrase → deterministic 64-byte seed
    - Seed derivation: known mnemonic + passphrase → DIFFERENT deterministic seed
  - **GREEN**: Implement:
    - `func Generate(wordCount int) (string, error)` — creates new BIP39 mnemonic (12/15/18/21/24 words)
    - `func Validate(mnemonic string) error` — validates mnemonic words + checksum
    - `func ToSeed(mnemonic string, passphrase string) []byte` — BIP39 PBKDF2 seed derivation
    - Map word count to entropy bits: 12→128, 15→160, 18→192, 21→224, 24→256
  - **REFACTOR**: Ensure error messages help users (e.g., "word 'xyz' not in BIP39 wordlist")

  **Must NOT do**:
  - Do not implement seed phrase storage or persistence
  - Do not reimplement BIP39 — wrap `blinklabs-io/go-bip39`
  - Do not use `tyler-smith/go-bip39` (deleted repo)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Thin wrapper around established library with clear test cases
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 6)
  - **Blocks**: T7 (generate command), T8 (derive command)
  - **Blocked By**: T2 (needs go.mod)

  **References**:

  **External References**:
  - `blinklabs-io/go-bip39`: https://github.com/blinklabs-io/go-bip39 — Maintained BIP39 fork (v0.2.0, Feb 2026, multi-language)
  - BIP39 spec: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki — Mnemonic generation + seed derivation algorithm
  - BIP39 test vectors: https://github.com/trezor/python-mnemonic/blob/master/vectors.json — Official test vectors for validation

  **Pattern References**:
  - `pinpox/mnemonic-ssh` main.go:22-30 — Shows BIP39 seed derivation pattern with `bip39.NewSeed(mnemonic, passphrase)`

  **WHY Each Reference Matters**:
  - blinklabs-io/go-bip39 is the maintained replacement for the deleted tyler-smith/go-bip39
  - BIP39 test vectors let us verify our wrapper produces correct seeds for known mnemonics
  - mnemonic-ssh shows the exact Go API pattern for NewSeed()

  **Acceptance Criteria**:
  - [ ] `go test ./internal/mnemonic/ -v` passes all mnemonic tests
  - [ ] Generated mnemonic is valid per BIP39 spec
  - [ ] Same mnemonic + passphrase always produces same seed (determinism)

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Generate and validate a 24-word mnemonic
    Tool: Bash
    Preconditions: internal/mnemonic/ package exists
    Steps:
      1. Run `go test ./internal/mnemonic/ -v -run TestGenerate24`
      2. Assert 24-word mnemonic generated
      3. Run `go test ./internal/mnemonic/ -v -run TestValidate`
      4. Assert known-good mnemonic validates successfully
    Expected Result: Mnemonic generation + validation works
    Failure Indicators: Wrong word count, validation rejection of valid mnemonic
    Evidence: .sisyphus/evidence/task-5-mnemonic-generate.txt

  Scenario: Reject invalid mnemonic
    Tool: Bash
    Steps:
      1. Run `go test ./internal/mnemonic/ -v -run TestInvalid`
      2. Assert garbage input returns descriptive error
    Expected Result: Clear error for invalid input
    Failure Indicators: Invalid mnemonic accepted
    Evidence: .sisyphus/evidence/task-5-mnemonic-invalid.txt
  ```

  **Commit**: YES
  - Message: `feat(mnemonic): implement BIP39 mnemonic generation and validation`
  - Files: `internal/mnemonic/mnemonic.go`, `internal/mnemonic/mnemonic_test.go`, `go.mod`, `go.sum`
  - Pre-commit: `go test ./internal/mnemonic/...`

- [ ] 6. SSH Key Generation from Ed25519 Seed (TDD)

  **What to do**:
  - Create `internal/keygen/keygen.go` and `internal/keygen/keygen_test.go`
  - **RED**: Write tests FIRST:
    - Known 32-byte seed → deterministic Ed25519 keypair → OpenSSH private key PEM → public key authorized_keys format
    - Verify private key PEM starts with `-----BEGIN OPENSSH PRIVATE KEY-----`
    - Verify public key starts with `ssh-ed25519 `
    - Verify public key contains the comment (derivation path)
    - Verify `ssh.ParsePrivateKey()` can re-parse the output (round-trip)
    - Test with different seeds → different keys
    - Test panic guard: reject non-32-byte seed gracefully (don't let `ed25519.NewKeyFromSeed` panic)
  - **GREEN**: Implement:
    - `type SSHKeyPair struct { PrivateKeyPEM []byte; PublicKey []byte; Fingerprint string }`
    - `func Generate(seed [32]byte, comment string) (*SSHKeyPair, error)`:
      1. `ed25519.NewKeyFromSeed(seed[:])` → private key
      2. `ssh.NewPublicKey(privateKey.Public())` → SSH public key
      3. `ssh.MarshalPrivateKey(privateKey, comment)` → PEM block
      4. `pem.EncodeToMemory(pemBlock)` → private key bytes
      5. `ssh.MarshalAuthorizedKey(pubKey)` → public key bytes
      6. Append comment to public key line
    - `func WriteKeyPair(pair *SSHKeyPair, outputPath string, force bool) error`:
      1. Check if files exist, refuse if not `force`
      2. Write private key to `outputPath` with mode 0600
      3. Write public key to `outputPath + ".pub"` with mode 0644
  - **REFACTOR**: Validate seed length before calling `NewKeyFromSeed`

  **Must NOT do**:
  - Do not encrypt the private key file (passphrase is BIP39 only, not file encryption)
  - Do not integrate with SLIP-0010 derivation (that coupling is in T8)
  - Do not add SSH agent support

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: SSH key format precision matters — malformed keys will be silently rejected by SSH
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 5)
  - **Blocks**: T8 (derive command), T9 (integration tests)
  - **Blocked By**: T2 (needs go.mod for golang.org/x/crypto/ssh)

  **References**:

  **External References**:
  - `golang.org/x/crypto/ssh` docs: https://pkg.go.dev/golang.org/x/crypto/ssh — MarshalPrivateKey, NewPublicKey, MarshalAuthorizedKey API
  - `golang.org/x/crypto/ed25519` docs: https://pkg.go.dev/golang.org/x/crypto/ed25519 — NewKeyFromSeed (panics on wrong size!)

  **Pattern References**:
  - `gravitational/teleport` api/utils/keys/privatekey.go:148-152 — Production OpenSSH Ed25519 key marshaling pattern (shows why PKCS#8 doesn't work)
  - `mikalv/anything2ed25519` main.go:29-48 — Simple seed→Ed25519→SSH key generation
  - `mikalv/anything2ed25519` main.go:129-130 — File permission pattern (0600 private, 0644 public)

  **WHY Each Reference Matters**:
  - Teleport is a production SSH system — their Ed25519 marshaling code is battle-tested
  - anything2ed25519 shows the minimal correct path from seed bytes to SSH key files
  - The golang.org/x/crypto docs define the exact function signatures and panic conditions

  **Acceptance Criteria**:
  - [ ] `go test ./internal/keygen/ -v` passes all keygen tests
  - [ ] Private key PEM is valid OpenSSH format
  - [ ] Public key is valid authorized_keys format
  - [ ] Round-trip: marshal → parse succeeds

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Generate SSH key from known seed and verify format
    Tool: Bash
    Preconditions: internal/keygen/ package exists
    Steps:
      1. Run `go test ./internal/keygen/ -v -run TestGenerate`
      2. Assert private key starts with "-----BEGIN OPENSSH PRIVATE KEY-----"
      3. Assert public key starts with "ssh-ed25519 "
      4. Assert round-trip parsing succeeds
    Expected Result: Valid OpenSSH Ed25519 key pair generated
    Failure Indicators: Wrong PEM header, parsing failure, missing comment
    Evidence: .sisyphus/evidence/task-6-keygen-format.txt

  Scenario: File write with correct permissions
    Tool: Bash
    Preconditions: WriteKeyPair implemented
    Steps:
      1. Run `go test ./internal/keygen/ -v -run TestWriteKeyPair`
      2. Assert output file has 0600 permissions
      3. Assert .pub file has 0644 permissions
    Expected Result: Key files written with correct Unix permissions
    Failure Indicators: Wrong file permissions, file not created
    Evidence: .sisyphus/evidence/task-6-keygen-perms.txt

  Scenario: Reject overwrite without --force
    Tool: Bash
    Steps:
      1. Run `go test ./internal/keygen/ -v -run TestRefuseOverwrite`
      2. Assert error returned when output file already exists and force=false
    Expected Result: Clear error about existing file, suggests --force
    Failure Indicators: File silently overwritten
    Evidence: .sisyphus/evidence/task-6-keygen-overwrite.txt
  ```

  **Commit**: YES
  - Message: `feat(keygen): implement Ed25519 SSH key generation with OpenSSH marshaling`
  - Files: `internal/keygen/keygen.go`, `internal/keygen/keygen_test.go`, `go.mod`, `go.sum`
  - Pre-commit: `go test ./internal/keygen/...`

---

- [ ] 7. `generate` Subcommand

  **What to do**:
  - Create `cmd/generate.go` and `cmd/generate_test.go`
  - **RED**: Write tests FIRST:
    - Execute generate command → outputs valid mnemonic to stdout
    - Default: 24 words
    - With `--words 12`: 12 words
    - Invalid `--words 13`: error (not a valid BIP39 word count)
  - **GREEN**: Implement Cobra subcommand:
    - Command name: `generate`
    - Short: "Generate a new BIP39 mnemonic seed phrase"
    - Flags:
      - `--words` (int, default 24): number of mnemonic words (12/15/18/21/24)
    - Logic:
      1. Validate --words is one of [12, 15, 18, 21, 24]
      2. Call `mnemonic.Generate(words)`
      3. Print mnemonic to stdout (one line, space-separated)
      4. Print to stderr: "\nIMPORTANT: Write down these words and store them securely. They cannot be recovered."
  - Register subcommand with root in `cmd/root.go`

  **Must NOT do**:
  - Do not save mnemonic to a file
  - Do not auto-derive keys (that's the `derive` command)
  - Do not prompt for passphrase (irrelevant for generation)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple command wiring to existing mnemonic package
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Task 8)
  - **Blocks**: T9 (integration tests)
  - **Blocked By**: T2 (Cobra scaffold), T5 (mnemonic package)

  **References**:

  **Pattern References**:
  - `cmd/root.go` (from T2) — Root command to register this subcommand with
  - `internal/mnemonic/mnemonic.go` (from T5) — Generate() function to call
  - `volodymyrprokopyuk/go-wallet` cli.go:259-268 — Cobra mnemonic subcommand pattern

  **WHY Each Reference Matters**:
  - root.go is where this command gets registered via `rootCmd.AddCommand()`
  - mnemonic.go provides the Generate() function this command calls
  - go-wallet shows the idiomatic Cobra subcommand pattern for mnemonic generation

  **Acceptance Criteria**:
  - [ ] `./bip32-ssh-keygen generate` outputs 24 words to stdout
  - [ ] `./bip32-ssh-keygen generate --words 12` outputs 12 words
  - [ ] `./bip32-ssh-keygen generate --words 13` returns error

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Generate 24-word mnemonic
    Tool: Bash
    Preconditions: Binary built with generate subcommand
    Steps:
      1. Run `go build -o bip32-ssh-keygen . && ./bip32-ssh-keygen generate 2>/dev/null`
      2. Capture stdout output
      3. Count words: `echo "$output" | wc -w`
      4. Assert word count is exactly 24
    Expected Result: 24 space-separated BIP39 words on stdout
    Failure Indicators: Wrong word count, words not from BIP39 wordlist
    Evidence: .sisyphus/evidence/task-7-generate-24.txt

  Scenario: Generate 12-word mnemonic
    Tool: Bash
    Steps:
      1. Run `./bip32-ssh-keygen generate --words 12 2>/dev/null`
      2. Assert output has exactly 12 words
    Expected Result: 12 BIP39 words
    Failure Indicators: Wrong word count
    Evidence: .sisyphus/evidence/task-7-generate-12.txt

  Scenario: Reject invalid word count
    Tool: Bash
    Steps:
      1. Run `./bip32-ssh-keygen generate --words 13 2>&1`
      2. Assert exit code is non-zero
      3. Assert stderr contains error about valid word counts
    Expected Result: Error message listing valid options (12, 15, 18, 21, 24)
    Failure Indicators: No error, or mnemonic generated
    Evidence: .sisyphus/evidence/task-7-generate-invalid.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add generate subcommand for BIP39 mnemonic creation`
  - Files: `cmd/generate.go`, `cmd/generate_test.go`, `cmd/root.go`
  - Pre-commit: `go test ./cmd/... && go build .`

- [ ] 8. `derive` Subcommand

  **What to do**:
  - Create `cmd/derive.go` and `cmd/derive_test.go`
  - This is the core integration point — connects ALL internal packages
  - **RED**: Write tests FIRST:
    - Known mnemonic + default path → deterministic SSH key files
    - Same mnemonic + same path = identical output (determinism)
    - Different path = different key
    - Invalid mnemonic → descriptive error
    - Missing mnemonic (no stdin, no TTY) → error
    - Existing output file without --force → error
    - Existing output file with --force → overwritten
  - **GREEN**: Implement Cobra subcommand:
    - Command name: `derive`
    - Short: "Derive an Ed25519 SSH key from a BIP39 mnemonic"
    - Flags:
      - `--path` (string, default `m/44'/22'/0'/0'`): derivation path
      - `--passphrase` (string, default ""): BIP39 passphrase for seed derivation
      - `--output` (string, default "id_ed25519"): output file path (without .pub extension)
      - `--force` (bool, default false): overwrite existing files
    - Logic:
      1. Read mnemonic (auto-detect TTY):
         - If stdin is a terminal: prompt with hidden input (use `golang.org/x/term`)
         - If stdin is piped: read from stdin, trim whitespace
      2. Validate mnemonic via `mnemonic.Validate(input)`
      3. Derive seed: `mnemonic.ToSeed(input, passphrase)`
      4. Parse path: `path.Parse(pathFlag)`
      5. Create SLIP-0010 master key: `slip10.NewMasterKey(seed)`
      6. Derive child keys along path: loop `key.DeriveChild(index)` for each path component
      7. Generate SSH key: `keygen.Generate(derivedKey.PrivateKey, pathFlag)`
      8. Write key files: `keygen.WriteKeyPair(pair, output, force)`
      9. Print to stderr: fingerprint, public key path, derivation path used
  - Register subcommand with root in `cmd/root.go`

  **Must NOT do**:
  - Do not encrypt the output key files
  - Do not add SSH agent integration
  - Do not support non-hardened derivation (path parser already rejects these)
  - Do not store the mnemonic anywhere

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Integration of all internal packages + TTY detection + error handling for all edge cases
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Task 7)
  - **Blocks**: T9 (integration tests)
  - **Blocked By**: T2 (Cobra), T3 (SLIP-0010), T4 (path), T5 (mnemonic), T6 (keygen)

  **References**:

  **Pattern References**:
  - `cmd/root.go` (from T2) — Root command for subcommand registration
  - `internal/slip10/slip10.go` (from T3) — NewMasterKey() + DeriveChild() API
  - `internal/path/path.go` (from T4) — Parse() + DefaultPath constant
  - `internal/mnemonic/mnemonic.go` (from T5) — Validate() + ToSeed() API
  - `internal/keygen/keygen.go` (from T6) — Generate() + WriteKeyPair() API
  - `mikalv/anything2ed25519` main.go:86-91 — TTY auto-detection pattern (`os.Stdin.Stat()` checking `ModeCharDevice`)
  - `pinpox/mnemonic-ssh` main.go:61-77 — Interactive mnemonic prompt pattern

  **External References**:
  - `golang.org/x/term` docs: https://pkg.go.dev/golang.org/x/term — `term.ReadPassword()` for hidden input

  **WHY Each Reference Matters**:
  - Each internal package reference defines the API contract this command wires together
  - anything2ed25519 shows the exact Go pattern for TTY detection (checking stdin stat for ModeCharDevice)
  - x/term provides cross-platform hidden password input

  **Acceptance Criteria**:
  - [ ] `go test ./cmd/ -v -run TestDerive` passes all derivation tests
  - [ ] Piped mnemonic: `echo "words..." | ./bip32-ssh-keygen derive` produces key files
  - [ ] Same mnemonic + path produces identical key on repeated runs

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Derive SSH key from piped mnemonic
    Tool: Bash
    Preconditions: Binary built, all internal packages working
    Steps:
      1. Run `./bip32-ssh-keygen generate 2>/dev/null` and capture mnemonic
      2. Run `echo "$mnemonic" | ./bip32-ssh-keygen derive --output /tmp/test_bip32_key --force`
      3. Assert exit code 0
      4. Assert /tmp/test_bip32_key exists with permissions 0600
      5. Assert /tmp/test_bip32_key.pub exists with permissions 0644
      6. Run `ssh-keygen -l -f /tmp/test_bip32_key`
      7. Assert output contains "256" and "ED25519"
    Expected Result: Valid Ed25519 SSH key pair generated from mnemonic
    Failure Indicators: Missing files, wrong permissions, ssh-keygen rejects key
    Evidence: .sisyphus/evidence/task-8-derive-piped.txt

  Scenario: Verify determinism — same input produces same key
    Tool: Bash
    Steps:
      1. Generate mnemonic: `mnemonic=$(./bip32-ssh-keygen generate 2>/dev/null)`
      2. Derive first time: `echo "$mnemonic" | ./bip32-ssh-keygen derive --output /tmp/key1 --force`
      3. Derive second time: `echo "$mnemonic" | ./bip32-ssh-keygen derive --output /tmp/key2 --force`
      4. Run `diff /tmp/key1 /tmp/key2`
      5. Assert diff output is empty (files identical)
      6. Run `diff /tmp/key1.pub /tmp/key2.pub`
      7. Assert diff output is empty
    Expected Result: Both derivations produce byte-identical key files
    Failure Indicators: Any diff output (non-determinism)
    Evidence: .sisyphus/evidence/task-8-derive-determinism.txt

  Scenario: Invalid mnemonic rejected
    Tool: Bash
    Steps:
      1. Run `echo "invalid garbage words here" | ./bip32-ssh-keygen derive --output /tmp/bad_key 2>&1`
      2. Assert exit code is non-zero
      3. Assert stderr contains error about invalid mnemonic
      4. Assert /tmp/bad_key does NOT exist
    Expected Result: Descriptive error, no key files created
    Failure Indicators: Key files created from invalid input
    Evidence: .sisyphus/evidence/task-8-derive-invalid.txt

  Scenario: Refuse overwrite without --force
    Tool: Bash
    Steps:
      1. Create dummy file: `touch /tmp/existing_key`
      2. Run `echo "valid mnemonic..." | ./bip32-ssh-keygen derive --output /tmp/existing_key 2>&1`
      3. Assert exit code non-zero
      4. Assert error mentions existing file and --force flag
    Expected Result: Error about existing file, key not overwritten
    Failure Indicators: File silently overwritten
    Evidence: .sisyphus/evidence/task-8-derive-noforce.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add derive subcommand for mnemonic-to-SSH-key derivation`
  - Files: `cmd/derive.go`, `cmd/derive_test.go`, `cmd/root.go`, `go.mod`, `go.sum`
  - Pre-commit: `go test ./... && go build .`

---

- [ ] 9. End-to-End Integration Test + Build Verification

  **What to do**:
  - Create `integration_test.go` (top-level, tests the full binary)
  - Write integration tests that exercise the complete pipeline:
    1. **Full lifecycle test**: `generate` → capture mnemonic → `derive` with that mnemonic → verify SSH key
    2. **Determinism test**: derive twice with same input → byte-identical output
    3. **Cross-path test**: same mnemonic + different `--path` → different keys
    4. **Passphrase test**: same mnemonic + passphrase vs no passphrase → different keys
    5. **ssh-keygen validation**: run `ssh-keygen -l -f <output>` to verify key format
    6. **Build verification**: `go build .`, `go vet ./...`, `go test ./...` all pass
  - Tests should use `os/exec` to run the actual binary (black-box testing)
  - Use `t.TempDir()` for test output files

  **Must NOT do**:
  - Do not modify any internal packages
  - Do not add new features — only test existing functionality

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Integration tests must exercise real OS interactions (file I/O, process execution, SSH tooling)
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (sequential — after T7, T8)
  - **Blocks**: F1-F4 (final verification)
  - **Blocked By**: T7 (generate command), T8 (derive command)

  **References**:

  **Pattern References**:
  - All `cmd/*.go` (from T7, T8) — The commands being integration-tested
  - All `internal/*/` (from T3-T6) — The modules wired together by the commands

  **External References**:
  - Go testing docs: https://pkg.go.dev/testing — t.TempDir(), os/exec for subprocess testing
  - SLIP-0010 test vectors: https://slips.readthedocs.io/en/latest/slip-0010/ — Known-good derivation outputs for end-to-end validation

  **WHY Each Reference Matters**:
  - The commands are the black-box interfaces we test through
  - SLIP-0010 vectors provide ground truth to verify the full pipeline produces correct keys

  **Acceptance Criteria**:
  - [ ] `go test -v -run TestIntegration` passes all integration tests
  - [ ] `go build .` succeeds with zero warnings
  - [ ] `go vet ./...` reports no issues
  - [ ] Full test suite: `go test ./...` — all pass

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Full lifecycle — generate mnemonic then derive SSH key
    Tool: Bash
    Preconditions: Binary built successfully
    Steps:
      1. Build: `go build -o bip32-ssh-keygen .`
      2. Generate: `mnemonic=$(./bip32-ssh-keygen generate 2>/dev/null)`
      3. Derive: `echo "$mnemonic" | ./bip32-ssh-keygen derive --output /tmp/e2e_key --force`
      4. Validate: `ssh-keygen -l -f /tmp/e2e_key`
      5. Assert output contains "256" and "ED25519"
      6. Check permissions: `stat -c '%a' /tmp/e2e_key` = "600"
      7. Check pub permissions: `stat -c '%a' /tmp/e2e_key.pub` = "644"
      8. Check pub format: `head -1 /tmp/e2e_key.pub` starts with "ssh-ed25519 "
    Expected Result: Complete pipeline works end-to-end
    Failure Indicators: Any step fails or produces wrong output
    Evidence: .sisyphus/evidence/task-9-e2e-lifecycle.txt

  Scenario: Passphrase changes derived key
    Tool: Bash
    Steps:
      1. Generate: `mnemonic=$(./bip32-ssh-keygen generate 2>/dev/null)`
      2. Derive without passphrase: `echo "$mnemonic" | ./bip32-ssh-keygen derive --output /tmp/nopass --force`
      3. Derive with passphrase: `echo "$mnemonic" | ./bip32-ssh-keygen derive --passphrase "test123" --output /tmp/withpass --force`
      4. Run `diff /tmp/nopass.pub /tmp/withpass.pub`
      5. Assert diff output is NON-empty (keys are different)
    Expected Result: Passphrase produces a completely different key
    Failure Indicators: Keys are identical (passphrase not applied)
    Evidence: .sisyphus/evidence/task-9-e2e-passphrase.txt

  Scenario: Full test suite passes
    Tool: Bash
    Steps:
      1. Run `go vet ./...`
      2. Assert exit code 0
      3. Run `go test ./... -v`
      4. Assert all tests pass
      5. Count total tests: assert > 15
    Expected Result: All unit and integration tests pass
    Failure Indicators: Any test failure or vet warning
    Evidence: .sisyphus/evidence/task-9-full-suite.txt
  ```

  **Commit**: YES
  - Message: `test: add end-to-end integration tests`
  - Files: `integration_test.go`
  - Pre-commit: `go test ./...`

---
## Final Verification Wave

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `go build ./...` + `go test ./...`. Review all files for: unchecked errors, panic-prone code, hardcoded secrets, unused imports, empty error handling. Check for AI slop: excessive comments, over-abstraction, generic variable names.
  Output: `Build [PASS/FAIL] | Vet [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Build binary. Run `generate` — verify mnemonic validity. Run `derive` with that mnemonic — verify SSH key works. Run `derive` again with same input — verify identical output (determinism). Test interactive mode and piped mode. Test `--force` overwrite. Test invalid mnemonic. Test `ssh-keygen -l -f` on output. Save evidence to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT Have" compliance (no SSH agent, no GUI, no config file, no batch mode). Flag unaccounted files.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **After T2**: `chore: scaffold Go module with Cobra CLI skeleton`
- **After T3**: `feat(slip10): implement SLIP-0010 Ed25519 HD key derivation with test vectors`
- **After T4**: `feat(path): implement BIP44 derivation path parser`
- **After T5**: `feat(mnemonic): implement BIP39 mnemonic generation and validation`
- **After T6**: `feat(keygen): implement Ed25519 SSH key generation with OpenSSH marshaling`
- **After T7**: `feat(cli): add generate subcommand for BIP39 mnemonic creation`
- **After T8**: `feat(cli): add derive subcommand for mnemonic-to-SSH-key derivation`
- **After T9**: `test: add end-to-end integration tests`
- **After T1** (whenever it completes): `chore: add devenv configuration for reproducible dev environment`

---

## Success Criteria

### Verification Commands
```bash
go build -o bip32-ssh-keygen .                    # Expected: binary created, exit 0
go test ./...                                       # Expected: all tests pass
go vet ./...                                        # Expected: no issues
./bip32-ssh-keygen generate                         # Expected: 24-word mnemonic on stdout
echo "mnemonic words..." | ./bip32-ssh-keygen derive --output /tmp/test_key  # Expected: key files created
ssh-keygen -l -f /tmp/test_key                      # Expected: 256 SHA256:... (ED25519)
stat -c '%a' /tmp/test_key                          # Expected: 600
stat -c '%a' /tmp/test_key.pub                      # Expected: 644
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass
- [ ] Determinism verified: same input → same output across invocations
- [ ] Binary builds cleanly with zero warnings
