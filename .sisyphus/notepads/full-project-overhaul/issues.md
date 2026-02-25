## Wave 0 — Task 1: Go Source Syntax Fixes

**Date:** 2026-02-25

### Fixes Applied

#### 1. internal/mnemonic/mnemonic.go (lines 4-6)
- **Issue:** Package doc comment continuation lines missing `//` prefix
- **Fix:** Added `//` prefix to lines 4-6
  ```go
  // Package mnemonic provides BIP39 mnemonic generation and validation.
  //
  // BIP39 defines a standard for creating mnemonic phrases that can be used
  // to generate deterministic cryptocurrency wallets. This package implements
  // the core BIP39 functionality for generating cryptographically secure
  // mnemonics and converting them to seeds.
  ```

#### 2. cmd/generate/generate.go (line 54)
- **Issue:** `fmt.Fprintln(os.Stderr, ...)` breaks cobra test compatibility
- **Fix:** Changed to `fmt.Fprintln(cmd.ErrOrStderr(), ...)`
  ```go
  fmt.Fprintln(cmd.ErrOrStderr(), "\nIMPORTANT: Write down these words and store them securely. They cannot be recovered.")
  ```

### Verification

✅ Both source files pass `go fmt` (syntax valid)
✅ `go fmt ./internal/mnemonic/mnemonic.go` - success
✅ `go fmt ./cmd/generate/generate.go` - success

### Notes

- Test files (mnemonic_test.go, generate_test.go) still have syntax errors but are out of scope for this task
- Wave 0 Task 1 completed successfully

