# bip32-ssh-keygen Full Project Overhaul

## TL;DR

> **Quick Summary**: Fix broken compilation from previous session, rename Go module to `github.com/valerius21/bip32-ssh-keygen`, add TUI/auto-generate features, achieve 100% test coverage, create Makefile + CI/CD + release workflows, fix devenv, and commit everything.
> 
> **Deliverables**:
> - Compiling, fully-tested Go CLI with TUI mode
> - Module path: `github.com/valerius21/bip32-ssh-keygen`
> - Makefile with standard targets
> - Working devenv.nix
> - .gitignore from gitignore.lol
> - GitHub Actions: CI (PR + push to master) + Release (on tag, linux/windows/macos)
> - .goreleaser.yaml for multi-platform builds
> - 100% test coverage
> - Clean git commit
> 
> **Estimated Effort**: Large
> **Parallel Execution**: YES — 6 waves
> **Critical Path**: Wave 0 (compile fixes) → Wave 1 (rename + env) → Wave 2 (tooling) → Wave 4 (verify) → Wave 5 (git)

---

## Context

### Original Request
User requested two batches of work:
1. **Code features**: TUI extension (bubbletea), derive command auto-generate mnemonic, 100% test coverage, reorganize cmd/ file structure, remove non-critical inline comments, expand docstrings.
2. **Project tooling**: Fix module path to `github.com/valerius21`, create Makefile, fix devenv.nix, create .gitignore from gitignore.lol, GitHub Actions workflows (CI + release), test with `act`, git add + commit.

### Interview Summary
**Key Discussions**:
- Previous session partially implemented code changes but introduced compilation errors
- The codebase currently does NOT compile — 4 blocking errors must be fixed first
- devenv.nix has iocraft-0.7.16 error (open issue cachix/devenv#2524)
- Module path rename from `github.com/valerius` to `github.com/valerius21` affects 13 import statements + go.mod

**Research Findings**:
- `internal/mnemonic/mnemonic.go` lines 4-6: missing `//` prefix in doc comment → syntax error
- `cmd/generate/generate_test.go` and `internal/mnemonic/mnemonic_test.go`: duplicate import blocks
- `cmd/tui/tui_test.go` lines 57-58, 185-186: `cmd.(tea.QuitMsg)` type assertion is wrong (tea.Quit returns a `tea.Cmd` func, not `QuitMsg`)
- `cmd/generate/generate.go` line 54: uses raw `os.Stderr` bypassing cobra's capture
- `cmd/generate/generate_test.go` line 97: `assert.Len(t, words, 1)` is wrong — 24-word mnemonic has 24 elements
- go.sum missing entry for `github.com/atotto/clipboard` (bubbles/textinput dep)
- devenv iocraft fix: pin devenv version or disable tasks subsystem
- gitignore.lol: `curl -sL https://www.toptal.com/developers/gitignore/api/go`
- GoReleaser: goreleaser-action@v7 with goreleaser v2+, CGO_ENABLED=0
- act: available via `nix-shell -p act`, test with `act -l` and `act push`
- Go CI: actions/setup-go@v5 + golangci-lint-action@v6

### Metis Review
**Identified Gaps** (addressed):
- Wave 0 mandatory before any other work — codebase doesn't compile
- Module rename must be atomic (all 13 imports + go.mod in one operation)
- CGO_ENABLED=0 required for all Go commands (no gcc in PATH)
- Go binary not in PATH — must use `nix-shell -p go` for all commands
- `generate.go` writes to raw `os.Stderr` bypassing cobra test capture
- GitHub repo `valerius21/bip32-ssh-keygen` may not exist — needs creation
- devenv iocraft error IS reproducible (confirmed from terminal output)
- Existing .gitignore should be enhanced, not replaced

---

## Work Objectives

### Core Objective
Transform bip32-ssh-keygen from a broken partial implementation into a fully-tested, properly-structured Go CLI with CI/CD, release automation, and clean dev environment.

### Concrete Deliverables
- All Go source files compile with `CGO_ENABLED=0 go build ./...`
- All tests pass with `CGO_ENABLED=0 go test ./...`
- Module path: `github.com/valerius21/bip32-ssh-keygen`
- `cmd/` organized as `cmd/generate/`, `cmd/derive/`, `cmd/tui/`
- TUI mode via `bip32-ssh-keygen tui`
- `bip32-ssh-keygen derive --generate` auto-creates mnemonic
- Makefile with targets: `build`, `test`, `lint`, `clean`, `release-dry-run`
- `.goreleaser.yaml` for linux/windows/macos × amd64/arm64
- `.github/workflows/ci.yml` triggered on PR + push to master
- `.github/workflows/release.yml` triggered on tag push
- Working `devenv.nix` (no iocraft error)
- Enhanced `.gitignore` with gitignore.lol content
- 100% test coverage (or as close as practical for TTY-dependent code)
- Clean git commit of all changes

### Definition of Done
- [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'` exits 0
- [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go test ./...'` — all pass
- [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go test -coverprofile=cov.out ./... && go tool cover -func=cov.out'` shows ≥95% per package
- [ ] `grep -r "github.com/valerius/" --include="*.go" . | wc -l` outputs 0
- [ ] `nix-shell -p goreleaser --run 'goreleaser check'` exits 0
- [ ] `.github/workflows/ci.yml` and `.github/workflows/release.yml` exist and parse as valid YAML
- [ ] `make build && make test && make clean` all succeed
- [ ] `git status` shows clean working tree after final commit

### Must Have
- All 4 compilation blockers fixed before any other work
- Module path is `github.com/valerius21/bip32-ssh-keygen` everywhere
- Makefile with at least: build, test, lint, clean
- CI workflow runs on PR + push to master
- Release workflow runs on tag push, builds for linux/windows/macos
- Working devenv shell

### Must NOT Have (Guardrails)
- Do NOT add homebrew taps, scoop manifests, Docker images, or signing to GoReleaser
- Do NOT add codecov uploads or security scanning to CI workflow
- Do NOT install `teatest` framework or expand TUI test coverage beyond fixing existing broken tests
- Do NOT refactor `derive.go` TTY detection logic just for coverage metrics
- Do NOT add docstrings to test files — only exported functions/types in non-test files
- Do NOT replace existing .gitignore content — only append/merge
- Do NOT add `install`, `docker-build`, or elaborate phony targets to Makefile
- Do NOT use `golangci-lint` in CI (keep it local via devenv only)

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (go test, testify)
- **Automated tests**: Tests-after (existing tests need fixing, coverage gaps need closing)
- **Framework**: go test + testify/assert + testify/require

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Go build/test**: Use Bash — `nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'`
- **Coverage**: Use Bash — `nix-shell -p go --run 'CGO_ENABLED=0 go test -coverprofile=cov.out ./...'`
- **GoReleaser**: Use Bash — `nix-shell -p goreleaser --run 'goreleaser check'`
- **YAML validation**: Use Bash — `python3 -c "import yaml; yaml.safe_load(open('file.yml'))"`
- **Workflow testing**: Use Bash — `nix-shell -p act --run 'act -l'`

### IMPORTANT: Go Build Environment
**Every task that invokes Go must use this pattern:**
```bash
nix-shell -p go --run 'CGO_ENABLED=0 go <command>'
```
Go is NOT in PATH. CGO requires gcc which is NOT available. Both constraints are non-negotiable.

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 0 (MANDATORY FIRST — fix compilation, 3 parallel):
├── Task 1: Fix Go source syntax errors [quick]
├── Task 2: Fix test file bugs [quick]
└── Task 3: Fix go.sum + missing deps [quick]

Wave 1 (After Wave 0 — rename + environment, 3 parallel):
├── Task 4: Atomic module path rename [quick]
├── Task 5: Fix devenv.nix [quick]
└── Task 6: Enhance .gitignore from gitignore.lol [quick]

Wave 2 (After Wave 1 — tooling + CI/CD, 3 parallel):
├── Task 7: Create Makefile [quick]
├── Task 8: Create .goreleaser.yaml [quick]
└── Task 9: Create GitHub workflows [quick]

Wave 3 (After Wave 1 — coverage, 3 parallel):
├── Task 10: Close coverage gaps in internal packages [unspecified-high]
├── Task 11: Close coverage gaps in cmd packages [unspecified-high]
└── Task 12: Update integration_test.go [quick]

Wave 4 (After Wave 2+3 — verification, 2 parallel):
├── Task 13: Test GH workflows with act [unspecified-high]
└── Task 14: Full verification + coverage report [deep]

Wave 5 (After ALL — git, 1 task):
└── Task 15: Git add + commit [quick]

Wave FINAL (After ALL tasks — independent review, 4 parallel):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)

Critical Path: T1-3 → T4 → T7-9 → T13-14 → T15 → F1-F4
Parallel Speedup: ~60% faster than sequential
Max Concurrent: 3 (Waves 0-3)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1-3  | — | 4-6 | 0 |
| 4 | 1-3 | 7-12 | 1 |
| 5 | 1-3 | 13 | 1 |
| 6 | 1-3 | 15 | 1 |
| 7 | 4 | 13, 15 | 2 |
| 8 | 4 | 13, 15 | 2 |
| 9 | 4 | 13, 15 | 2 |
| 10 | 4 | 14 | 3 |
| 11 | 4 | 14 | 3 |
| 12 | 4 | 14 | 3 |
| 13 | 7-9 | 15 | 4 |
| 14 | 10-12 | 15 | 4 |
| 15 | 13-14 | F1-F4 | 5 |
| F1-F4 | 15 | — | FINAL |

### Agent Dispatch Summary

- **Wave 0**: 3 tasks → `quick` × 3
- **Wave 1**: 3 tasks → `quick` × 3
- **Wave 2**: 3 tasks → `quick` × 3
- **Wave 3**: 3 tasks → `unspecified-high` × 2, `quick` × 1
- **Wave 4**: 2 tasks → `unspecified-high` × 1, `deep` × 1
- **Wave 5**: 1 task → `quick` (+ `git-master` skill)
- **FINAL**: 4 tasks → `oracle` × 1, `unspecified-high` × 2, `deep` × 1

---

## TODOs

### Wave 0 — Fix Compilation (MANDATORY FIRST)

- [x] 1. Fix Go source syntax errors

  **What to do**:
  - Fix `internal/mnemonic/mnemonic.go` lines 4-6: add missing `//` prefix to doc comment continuation lines. Lines currently read `to generate deterministic...`, `the core BIP39 functionality...`, `mnemonics and converting them to seeds.` — each needs `//` prepended.
  - Fix `cmd/generate/generate.go` line 54: change `fmt.Fprintln(os.Stderr, ...)` to `fmt.Fprintln(cmd.ErrOrStderr(), ...)` so the stderr warning is capturable by cobra's test harness.
  - Verify both files parse: `nix-shell -p go --run 'CGO_ENABLED=0 go vet ./internal/mnemonic/ ./cmd/generate/'`

  **Must NOT do**:
  - Do not change any logic or test files
  - Do not modify any other files

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 0 (with Tasks 2, 3)
  - **Blocks**: Tasks 4-6
  - **Blocked By**: None

  **References**:
  - `internal/mnemonic/mnemonic.go:4-6` — lines missing `//` prefix in doc comment
  - `cmd/generate/generate.go:54` — `os.Stderr` should be `cmd.ErrOrStderr()`

  **Acceptance Criteria**:
  - [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go vet ./internal/mnemonic/'` exits 0
  - [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go vet ./cmd/generate/'` exits 0

  **QA Scenarios:**
  ```
  Scenario: mnemonic.go compiles
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go vet ./internal/mnemonic/'
      2. Assert exit code 0
    Expected Result: No syntax errors
    Evidence: .sisyphus/evidence/task-1-mnemonic-vet.txt

  Scenario: generate.go compiles and uses cmd.ErrOrStderr()
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go vet ./cmd/generate/'
      2. grep 'os.Stderr' cmd/generate/generate.go — should return empty (no raw stderr)
      3. grep 'cmd.ErrOrStderr' cmd/generate/generate.go — should return match
    Expected Result: No raw os.Stderr usage in generate.go
    Evidence: .sisyphus/evidence/task-1-generate-vet.txt
  ```

  **Commit**: NO (groups with Wave 0)

- [x] 2. Fix test file bugs

  **What to do**:
  - Fix `cmd/generate/generate_test.go`: remove duplicate import block (lines 3-10 likely duplicated). Ensure single clean import block with `bytes`, `fmt`, `strings`, `testing`, `testify/assert`, `testify/require`.
  - Fix `cmd/generate/generate_test.go` line ~97: change `assert.Len(t, words, 1)` to proper validation — the default 24-word mnemonic has 24 words, not 1. The line should check `len(words) > 0` or validate against expected count from the output.
  - Fix `internal/mnemonic/mnemonic_test.go`: same duplicate import block issue. Ensure single clean import with `encoding/hex`, `fmt`, `strings`, `testing`, `testify/assert`, `testify/require`.
  - Fix `cmd/tui/tui_test.go` lines 57-58 and 185-186: replace `_, isQuit := cmd.(tea.QuitMsg)` / `assert.True(t, isQuit)` with `assert.NotNil(t, cmd)` — `tea.Quit` returns a `tea.Cmd` (function), NOT a `tea.QuitMsg` value. Type-asserting against `tea.QuitMsg` always yields false.
  - Fix `integration_test.go`: add missing `fmt` import, update version reference from `v0.1.0` to `v0.2.0`.
  - Verify: `nix-shell -p go --run 'CGO_ENABLED=0 go vet ./...'`

  **Must NOT do**:
  - Do not add new test cases — only fix existing broken ones
  - Do not modify non-test files

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 0 (with Tasks 1, 3)
  - **Blocks**: Tasks 4-6
  - **Blocked By**: None

  **References**:
  - `cmd/generate/generate_test.go:3-10` — duplicate import block to merge
  - `cmd/generate/generate_test.go:97` — `assert.Len(t, words, 1)` is wrong
  - `internal/mnemonic/mnemonic_test.go:3-10` — duplicate import block to merge
  - `cmd/tui/tui_test.go:57-58,185-186` — `tea.QuitMsg` type assertion is fundamentally wrong
  - `integration_test.go` — needs `fmt` import added
  - Bubbletea docs: `tea.Quit` returns `tea.Cmd` (a function type), not `tea.QuitMsg`

  **Acceptance Criteria**:
  - [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go vet ./...'` exits 0 (all test files parse)
  - [ ] No duplicate import blocks in any test file
  - [ ] No `tea.QuitMsg` type assertions in tui_test.go

  **QA Scenarios:**
  ```
  Scenario: All test files compile
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go vet ./...'
      2. Assert exit code 0
    Expected Result: Zero compilation errors across all packages
    Evidence: .sisyphus/evidence/task-2-vet-all.txt

  Scenario: No duplicate imports remain
    Tool: Bash (grep)
    Steps:
      1. grep -c '^import (' cmd/generate/generate_test.go — should be exactly 1
      2. grep -c '^import (' internal/mnemonic/mnemonic_test.go — should be exactly 1
    Expected Result: Each file has exactly 1 import block
    Evidence: .sisyphus/evidence/task-2-imports-check.txt
  ```

  **Commit**: NO (groups with Wave 0)

- [x] 3. Fix go.sum and missing dependencies

  **What to do**:
  - Run `nix-shell -p go --run 'CGO_ENABLED=0 go get github.com/charmbracelet/bubbles/textinput@v1.0.0'` to add missing `github.com/atotto/clipboard` transitive dependency to go.sum.
  - Run `nix-shell -p go --run 'CGO_ENABLED=0 go mod tidy'` to clean up go.mod and go.sum.
  - Verify: `nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'` exits 0.

  **Must NOT do**:
  - Do not change any Go source files
  - Do not add new direct dependencies

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 0 (with Tasks 1, 2)
  - **Blocks**: Tasks 4-6
  - **Blocked By**: None

  **References**:
  - `go.mod` — current dependency list
  - `go.sum` — missing entry for `github.com/atotto/clipboard`
  - Error: `missing go.sum entry for module providing package github.com/atotto/clipboard`

  **Acceptance Criteria**:
  - [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'` exits 0
  - [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go mod verify'` exits 0

  **QA Scenarios:**
  ```
  Scenario: Full project builds
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'
      2. Assert exit code 0
    Expected Result: All packages build successfully
    Evidence: .sisyphus/evidence/task-3-build.txt
  ```

  **Commit**: NO (groups with Wave 0)

### Wave 1 — Module Rename + Environment

- [x] 4. Atomic module path rename

  **What to do**:
  - Change `go.mod` line 1 from `module github.com/valerius/bip32-ssh-keygen` to `module github.com/valerius21/bip32-ssh-keygen`.
  - Find ALL import statements referencing `github.com/valerius/bip32-ssh-keygen` and replace with `github.com/valerius21/bip32-ssh-keygen`. There are ~13 imports across 5 files: `main.go`, `cmd/root.go`, `cmd/derive/derive.go`, `cmd/generate/generate.go`, `cmd/tui/tui.go`.
  - Use `ast_grep_replace` or global find-and-replace to do this atomically.
  - Run `nix-shell -p go --run 'CGO_ENABLED=0 go mod tidy'` after rename.
  - Verify: `nix-shell -p go --run 'CGO_ENABLED=0 go build ./... && CGO_ENABLED=0 go test ./...'`

  **Must NOT do**:
  - Do not partially rename (all or nothing)
  - Do not change package names (only import paths)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 5, 6)
  - **Blocks**: Tasks 7-12
  - **Blocked By**: Tasks 1-3

  **References**:
  - `go.mod:1` — module declaration to rename
  - `main.go:3` — imports cmd package
  - `cmd/root.go:10-12` — imports derive, generate, tui packages
  - `cmd/derive/derive.go:11-15` — imports internal packages
  - `cmd/generate/generate.go:8` — imports mnemonic package
  - `cmd/tui/tui.go:13-16` — imports internal packages

  **Acceptance Criteria**:
  - [ ] `grep -r 'github.com/valerius/' --include='*.go' . | wc -l` outputs 0
  - [ ] `grep 'github.com/valerius21/bip32-ssh-keygen' go.mod` finds the module line
  - [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'` exits 0
  - [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go test ./...'` — all pass

  **QA Scenarios:**
  ```
  Scenario: No old module path remains
    Tool: Bash (grep)
    Steps:
      1. grep -r 'github.com/valerius/' --include='*.go' .
      2. Assert empty output (exit code 1 = no matches)
    Expected Result: Zero occurrences of old module path
    Evidence: .sisyphus/evidence/task-4-no-old-path.txt

  Scenario: Full build + test after rename
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'
      2. nix-shell -p go --run 'CGO_ENABLED=0 go test ./...'
      3. Assert both exit 0
    Expected Result: All packages build and tests pass with new module path
    Evidence: .sisyphus/evidence/task-4-build-test.txt
  ```

  **Commit**: NO (groups with Wave 1)

- [x] 5. Fix devenv.nix

  **What to do**:
  - The current devenv.nix triggers the iocraft-0.7.16 error (cachix/devenv#2524) when building the shell. This was confirmed in the terminal output.
  - Clean up devenv.nix: remove all boilerplate comments (`# https://devenv.sh/basics/`, commented-out examples, `env.GREET`, `scripts.hello`, the `enterShell` hello/git commands, the `enterTest` section).
  - Keep only: `packages` list (gopls, golangci-lint, delve, plus add `act` and `goreleaser`), `languages.go.enable = true`.
  - To fix the iocraft error, try one of these approaches (in order of preference):
    1. Add `devenv.tasks = {};` to explicitly disable the tasks subsystem.
    2. If that doesn't work, pin devenv to a pre-iocraft version by adding to `devenv.yaml`: `devenv: url: github:cachix/devenv/<known-good-rev>`
  - Verify: `devenv shell -- go version` should succeed.

  **Must NOT do**:
  - Do not remove `languages.go.enable = true`
  - Do not add unnecessary packages

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 4, 6)
  - **Blocks**: Task 13
  - **Blocked By**: Tasks 1-3

  **References**:
  - `devenv.nix` — current config (full of boilerplate comments)
  - `devenv.yaml` — input configuration
  - cachix/devenv#2524 — iocraft-0.7.16 error description
  - The error trace from earlier: `error: A hash was specified for iocraft-0.7.16, but there is no corresponding git dependency`

  **Acceptance Criteria**:
  - [ ] `devenv shell -- go version` exits 0 and shows `go1.25`
  - [ ] No boilerplate comments remain in devenv.nix
  - [ ] devenv.nix contains gopls, golangci-lint, delve, act, goreleaser in packages

  **QA Scenarios:**
  ```
  Scenario: devenv shell works
    Tool: Bash
    Steps:
      1. devenv shell -- go version
      2. Assert exit 0, output contains 'go1.25'
    Expected Result: Working devenv shell with Go available
    Evidence: .sisyphus/evidence/task-5-devenv-shell.txt

  Scenario: devenv.nix is clean
    Tool: Bash (grep)
    Steps:
      1. grep -c 'https://devenv.sh' devenv.nix — should be 0
      2. grep -c 'GREET' devenv.nix — should be 0
    Expected Result: No boilerplate remains
    Evidence: .sisyphus/evidence/task-5-devenv-clean.txt
  ```

  **Commit**: NO (groups with Wave 1)

- [x] 6. Enhance .gitignore from gitignore.lol

  **What to do**:
  - Fetch the Go gitignore template: `curl -sL https://www.toptal.com/developers/gitignore/api/go`
  - Merge the fetched content with the existing `.gitignore` — do NOT replace. The existing file has project-specific entries (`bip32-ssh-keygen`, `.sisyphus/evidence/`, `.devenv*`, `.direnv`, `.pre-commit-config.yaml`) that must be preserved.
  - Strategy: read existing .gitignore, fetch gitignore.lol content, append any new entries from gitignore.lol that aren't already present. Add under a `# gitignore.lol - Go` header.
  - Also add: `.goreleaser-dist/`, `dist/` (GoReleaser output), `.github/act/` if needed.

  **Must NOT do**:
  - Do NOT replace existing .gitignore — only append/merge
  - Do NOT remove project-specific entries

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 4, 5)
  - **Blocks**: Task 15
  - **Blocked By**: Tasks 1-3

  **References**:
  - `.gitignore` — existing content (already has Go basics + devenv/direnv entries)
  - `https://www.toptal.com/developers/gitignore/api/go` — gitignore.lol Go template

  **Acceptance Criteria**:
  - [ ] .gitignore contains `*.exe`, `*.test`, `*.out`, `go.work` entries from gitignore.lol
  - [ ] .gitignore still contains `bip32-ssh-keygen`, `.sisyphus/evidence/`, `.devenv*`
  - [ ] .gitignore contains `dist/` for GoReleaser output

  **QA Scenarios:**
  ```
  Scenario: gitignore.lol content present
    Tool: Bash (grep)
    Steps:
      1. grep 'go.work' .gitignore — should match
      2. grep 'dist/' .gitignore — should match
    Expected Result: New entries from gitignore.lol are present
    Evidence: .sisyphus/evidence/task-6-gitignore.txt

  Scenario: Project-specific entries preserved
    Tool: Bash (grep)
    Steps:
      1. grep 'bip32-ssh-keygen' .gitignore — should match
      2. grep '.sisyphus/evidence' .gitignore — should match
    Expected Result: Existing project entries still present
    Evidence: .sisyphus/evidence/task-6-gitignore-preserved.txt
  ```

  **Commit**: NO (groups with Wave 1)

### Wave 2 — Tooling + CI/CD

- [ ] 7. Create Makefile

  **What to do**:
  - Create a `Makefile` at project root with these targets:
    - `build`: `CGO_ENABLED=0 go build -o bip32-ssh-keygen .`
    - `test`: `CGO_ENABLED=0 go test -v -race -coverprofile=coverage.out ./...`
    - `lint`: `golangci-lint run ./...`
    - `clean`: `rm -f bip32-ssh-keygen coverage.out`
    - `release-dry-run`: `goreleaser release --snapshot --clean`
  - Include `.PHONY` declarations for all targets.
  - `build` should be the default target.
  - Use `$(shell which go)` or just `go` (assumes Go in PATH via devenv).

  **Must NOT do**:
  - No `install`, `docker-build`, `docker-push` targets
  - No elaborate multi-stage builds
  - Maximum 5 targets

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 8, 9)
  - **Blocks**: Tasks 13, 15
  - **Blocked By**: Task 4

  **References**:
  - Standard Go Makefile patterns
  - GoReleaser: `goreleaser release --snapshot --clean` for dry-run

  **Acceptance Criteria**:
  - [ ] `make build` produces `bip32-ssh-keygen` binary, exits 0
  - [ ] `make test` runs all tests, exits 0
  - [ ] `make clean` removes binary and coverage.out
  - [ ] `make lint` runs golangci-lint (exits 0 or with lint-only warnings)

  **QA Scenarios:**
  ```
  Scenario: Makefile targets work
    Tool: Bash
    Steps:
      1. nix-shell -p go -p golangci-lint --run 'make build'
      2. ls -la bip32-ssh-keygen — binary exists
      3. nix-shell -p go --run 'make test'
      4. nix-shell -p go --run 'make clean'
      5. ls bip32-ssh-keygen — should not exist
    Expected Result: All targets execute successfully
    Evidence: .sisyphus/evidence/task-7-makefile.txt
  ```

  **Commit**: NO (groups with Wave 2)

- [ ] 8. Create .goreleaser.yaml

  **What to do**:
  - Create `.goreleaser.yaml` at project root with:
    - `project_name: bip32-ssh-keygen`
    - `builds`: single build entry, `env: [CGO_ENABLED=0]`, `goos: [linux, windows, darwin]`, `goarch: [amd64, arm64]`
    - `archives`: tar.gz default, zip override for windows, name template `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}`
    - `checksum`: sha256 checksums file
    - `changelog`: sort asc, exclude `docs:`, `test:`, `chore:`
    - `release`: github with owner `valerius21`
  - Verify: `nix-shell -p goreleaser --run 'goreleaser check'`

  **Must NOT do**:
  - No homebrew taps, scoop manifests, Docker images, snapcraft, or signing
  - No nfpms (package managers)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 9)
  - **Blocks**: Tasks 13, 15
  - **Blocked By**: Task 4

  **References**:
  - GoReleaser docs: https://goreleaser.com/customization/
  - goreleaser-action@v7 expects `.goreleaser.yaml` or `.goreleaser.yml`

  **Acceptance Criteria**:
  - [ ] `nix-shell -p goreleaser --run 'goreleaser check'` exits 0
  - [ ] `.goreleaser.yaml` contains `linux`, `windows`, `darwin` in goos
  - [ ] `.goreleaser.yaml` contains `amd64`, `arm64` in goarch

  **QA Scenarios:**
  ```
  Scenario: GoReleaser config is valid
    Tool: Bash
    Steps:
      1. nix-shell -p goreleaser --run 'goreleaser check'
      2. Assert exit 0
    Expected Result: Config passes GoReleaser validation
    Evidence: .sisyphus/evidence/task-8-goreleaser-check.txt
  ```

  **Commit**: NO (groups with Wave 2)

- [ ] 9. Create GitHub workflows (CI + Release)

  **What to do**:
  - Create `.github/workflows/ci.yml`:
    - Triggers: `push` to `main`/`master` branches, `pull_request`
    - Jobs: `test` (checkout, setup-go with `go-version-file: go.mod`, `CGO_ENABLED=0 go build ./...`, `CGO_ENABLED=0 go test -v -race ./...`)
    - Permissions: `contents: read`
    - Runs on `ubuntu-latest`
  - Create `.github/workflows/release.yml`:
    - Triggers: `push` with `tags: ['v*']`
    - Jobs: `release` (checkout with `fetch-depth: 0`, setup-go, goreleaser-action@v7 with `args: release --clean`)
    - Permissions: `contents: write`
    - Env: `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`
  - Verify both files are valid YAML.

  **Must NOT do**:
  - No codecov uploads, security scanning, or multi-OS matrix in CI
  - No golangci-lint in CI (local only via devenv)
  - No Docker login or signing in release

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 8)
  - **Blocks**: Tasks 13, 15
  - **Blocked By**: Task 4

  **References**:
  - GitHub Actions Go CI docs: https://docs.github.com/en/actions/automating-builds-and-tests/building-and-testing-go
  - GoReleaser CI docs: https://goreleaser.com/ci/actions/
  - `goreleaser/goreleaser-action@v7` — latest release action
  - `actions/setup-go@v5` — setup Go with `go-version-file: go.mod`

  **Acceptance Criteria**:
  - [ ] `.github/workflows/ci.yml` exists and is valid YAML
  - [ ] CI triggers on `push` to main/master and `pull_request`
  - [ ] `.github/workflows/release.yml` exists and is valid YAML
  - [ ] Release triggers on tag push (`v*`)
  - [ ] Release uses goreleaser-action@v7

  **QA Scenarios:**
  ```
  Scenario: CI workflow is valid YAML
    Tool: Bash
    Steps:
      1. python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"
      2. Assert exit 0
    Expected Result: Valid YAML, no parse errors
    Evidence: .sisyphus/evidence/task-9-ci-yaml.txt

  Scenario: Release workflow is valid YAML
    Tool: Bash
    Steps:
      1. python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))"
      2. Assert exit 0
    Expected Result: Valid YAML, no parse errors
    Evidence: .sisyphus/evidence/task-9-release-yaml.txt
  ```

  **Commit**: NO (groups with Wave 2)

### Wave 3 — Test Coverage

- [ ] 10. Close coverage gaps in internal packages

  **What to do**:
  - Run `nix-shell -p go --run 'CGO_ENABLED=0 go test -coverprofile=cov.out ./internal/... && go tool cover -func=cov.out'` to identify uncovered lines.
  - **keygen** (currently 83.3%): Likely missing coverage for error branches in `Generate()` (SSH public key creation error, marshal error) and `WriteKeyPair()` error paths. Add tests that exercise all branches.
  - **path** (currently 96.4%): Likely missing coverage for the non-hardened branch in `FormatPath()`. Add test case with non-hardened index (e.g., `[]uint32{5}` → `m/5`).
  - **mnemonic**: Verify coverage after syntax fix. Should be close to 100% already.
  - **slip10**: Already at 100% — verify still 100%.
  - Target: ≥95% per package, ideally 100%.

  **Must NOT do**:
  - Do not change implementation code to improve testability
  - Do not add mock frameworks

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 11, 12)
  - **Blocks**: Task 14
  - **Blocked By**: Task 4

  **References**:
  - `internal/keygen/keygen.go:44` — `Generate()` has error branches at lines ~47, ~52 that need coverage
  - `internal/keygen/keygen.go:85` — `WriteKeyPair()` has multiple error paths
  - `internal/path/path.go:99` — `FormatPath()` has a non-hardened branch at line ~100
  - `internal/keygen/keygen_test.go` — existing tests to extend
  - `internal/path/path_test.go` — existing tests to extend

  **Acceptance Criteria**:
  - [ ] `internal/keygen` coverage ≥95%
  - [ ] `internal/path` coverage = 100%
  - [ ] `internal/mnemonic` coverage ≥95%
  - [ ] `internal/slip10` coverage = 100%

  **QA Scenarios:**
  ```
  Scenario: Internal package coverage meets targets
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go test -coverprofile=cov.out ./internal/...'
      2. nix-shell -p go --run 'go tool cover -func=cov.out'
      3. Verify each package meets target
    Expected Result: All internal packages ≥95%
    Evidence: .sisyphus/evidence/task-10-internal-coverage.txt
  ```

  **Commit**: NO (groups with Wave 3)

- [ ] 11. Close coverage gaps in cmd packages

  **What to do**:
  - Run coverage for cmd packages: `nix-shell -p go --run 'CGO_ENABLED=0 go test -coverprofile=cov.out ./cmd/...'`
  - **cmd** (root): `root_test.go` exists — verify coverage. Add test for `Execute()` function if not covered.
  - **cmd/generate**: Tests exist but had bugs (fixed in Task 2). Verify coverage after fixes. Add any missing test cases for all word counts and error paths.
  - **cmd/derive**: Tests exist — verify coverage. The `--generate` flag path needs testing (already partially tested). Verify all error paths are covered.
  - **cmd/tui**: Tests exist (model, navigation, screen transitions). TUI test coverage is inherently limited — don't try to achieve 100% for bubbletea View() rendering. Focus on logic coverage (handleEnter, Update, Init).
  - Target: ≥90% for generate/derive logic, ≥70% for tui (View rendering is hard to unit-test).

  **Must NOT do**:
  - Do not install `teatest` framework
  - Do not try to achieve 100% coverage for TUI View() methods
  - Do not refactor derive.go TTY detection for testability

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10, 12)
  - **Blocks**: Task 14
  - **Blocked By**: Task 4

  **References**:
  - `cmd/root.go` + `cmd/root_test.go` — root command tests
  - `cmd/generate/generate.go` + `cmd/generate/generate_test.go` — generate tests (fixed in Task 2)
  - `cmd/derive/derive.go` + `cmd/derive/derive_test.go` — derive tests with --generate flag
  - `cmd/tui/tui.go` + `cmd/tui/tui_test.go` — TUI logic tests

  **Acceptance Criteria**:
  - [ ] `cmd/generate` coverage ≥90%
  - [ ] `cmd/derive` coverage ≥85%
  - [ ] `cmd/tui` coverage ≥70%
  - [ ] `cmd` (root) has working tests

  **QA Scenarios:**
  ```
  Scenario: Cmd package coverage meets targets
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go test -coverprofile=cov.out ./cmd/...'
      2. nix-shell -p go --run 'go tool cover -func=cov.out'
    Expected Result: generate ≥90%, derive ≥85%, tui ≥70%
    Evidence: .sisyphus/evidence/task-11-cmd-coverage.txt
  ```

  **Commit**: NO (groups with Wave 3)

- [ ] 12. Update integration_test.go for new structure

  **What to do**:
  - Update `integration_test.go` to work with the new module path and cmd structure.
  - Add `fmt` import (needed by test cases that use `fmt.Sprintf`).
  - Update version check from `v0.1.0` to `v0.2.0`.
  - Add test case for `--generate` flag: build binary, run `derive --generate --output <path> --force`, verify key files created and mnemonic printed to stdout.
  - Add test case for `tui` subcommand: just verify `--help` works (can't interactively test TUI in CI).
  - Verify: `nix-shell -p go --run 'CGO_ENABLED=0 go test -v -run TestIntegration .'`

  **Must NOT do**:
  - Do not try to interactively test the TUI
  - Do not add flaky tests that depend on timing

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10, 11)
  - **Blocks**: Task 14
  - **Blocked By**: Task 4

  **References**:
  - `integration_test.go` — current integration tests (needs fmt import, version update)
  - The binary is built via `go build -o binPath .` in the test

  **Acceptance Criteria**:
  - [ ] `nix-shell -p go --run 'CGO_ENABLED=0 go test -v -run TestIntegration .'` — all pass
  - [ ] Test for `--generate` flag exists and passes
  - [ ] Test for `tui --help` exists and passes

  **QA Scenarios:**
  ```
  Scenario: Integration tests pass
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go test -v -run TestIntegration .'
      2. Assert all subtests pass
    Expected Result: All integration test scenarios pass
    Evidence: .sisyphus/evidence/task-12-integration.txt
  ```

  **Commit**: NO (groups with Wave 3)

### Wave 4 — Verification

- [ ] 13. Test GitHub workflows with `act`

  **What to do**:
  - Use `act` to test the CI workflow locally:
    - `nix-shell -p act --run 'act -l'` — list available workflows
    - `nix-shell -p act --run 'act push --dryrun'` — dry-run the CI workflow
    - If dry-run succeeds, try actual run: `nix-shell -p act --run 'act push'`
  - For the release workflow, use dry-run only (needs GITHUB_TOKEN): `nix-shell -p act --run 'act push --workflows .github/workflows/release.yml --dryrun'`
  - If `act` has issues with Go version or Docker images, document the issue and mark as known limitation. `act` is not a perfect replica of GitHub Actions.
  - Minimum acceptance: `act -l` lists both workflows, dry-run doesn't show YAML errors.

  **Must NOT do**:
  - Do not spend more than 15 minutes debugging `act` issues
  - Do not modify workflows just to make `act` happy if they're valid for GitHub

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Task 14)
  - **Blocks**: Task 15
  - **Blocked By**: Tasks 7-9

  **References**:
  - `.github/workflows/ci.yml` — CI workflow to test
  - `.github/workflows/release.yml` — Release workflow to test
  - `act` docs: https://github.com/nektos/act
  - Common flags: `act -l`, `act push`, `act --dryrun`

  **Acceptance Criteria**:
  - [ ] `nix-shell -p act --run 'act -l'` lists both ci and release workflows
  - [ ] `act push --dryrun` succeeds (no YAML errors)
  - [ ] Any `act` limitations are documented

  **QA Scenarios:**
  ```
  Scenario: act lists workflows
    Tool: Bash
    Steps:
      1. nix-shell -p act --run 'act -l'
      2. Assert output contains 'ci' and 'release' (or workflow names)
    Expected Result: Both workflows listed
    Evidence: .sisyphus/evidence/task-13-act-list.txt
  ```

  **Commit**: NO (groups with Wave 4)

- [ ] 14. Full verification + coverage report

  **What to do**:
  - Run full build: `nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'`
  - Run full test suite: `nix-shell -p go --run 'CGO_ENABLED=0 go test -v -coverprofile=coverage.out ./...'`
  - Generate coverage report: `nix-shell -p go --run 'go tool cover -func=coverage.out'`
  - Run vet: `nix-shell -p go --run 'CGO_ENABLED=0 go vet ./...'`
  - Validate goreleaser: `nix-shell -p goreleaser --run 'goreleaser check'`
  - Validate Makefile: `nix-shell -p go --run 'make build && make clean'`
  - Capture all outputs as evidence.
  - If any check fails, report the failure clearly with exact error.

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Task 13)
  - **Blocks**: Task 15
  - **Blocked By**: Tasks 10-12

  **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `go test ./...` — all pass, 0 failures
  - [ ] Coverage report shows ≥90% overall
  - [ ] `go vet ./...` exits 0
  - [ ] `goreleaser check` exits 0
  - [ ] `make build && make clean` exits 0

  **QA Scenarios:**
  ```
  Scenario: Complete project health check
    Tool: Bash
    Steps:
      1. nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'
      2. nix-shell -p go --run 'CGO_ENABLED=0 go test -v -coverprofile=coverage.out ./...'
      3. nix-shell -p go --run 'go tool cover -func=coverage.out'
      4. nix-shell -p go --run 'CGO_ENABLED=0 go vet ./...'
      5. nix-shell -p goreleaser --run 'goreleaser check'
    Expected Result: All checks pass, coverage ≥90%
    Evidence: .sisyphus/evidence/task-14-full-verification.txt
  ```

  **Commit**: NO (groups with Wave 4)

### Wave 5 — Git

- [ ] 15. Git add + commit

  **What to do**:
  - Stage all changed, added, and deleted files: `git add -A`
  - Review staged changes: `git status` and `git diff --cached --stat`
  - Create a single atomic commit: `git commit -m 'feat: full project overhaul — TUI, CI/CD, coverage, module rename'`
  - If a GitHub repo for `valerius21/bip32-ssh-keygen` exists, update remote and push. If not, create it with `gh repo create valerius21/bip32-ssh-keygen --public --source=. --push`.
  - Verify: `git log --oneline -1` shows the new commit.

  **Must NOT do**:
  - Do not force push
  - Do not amend existing commits
  - Do not push if tests fail

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`git-master`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 13-14

  **References**:
  - Current git state: 5 existing commits, 20+ modified/untracked files
  - Remote: may need updating from `valerius` to `valerius21`

  **Acceptance Criteria**:
  - [ ] `git status` shows clean working tree
  - [ ] `git log --oneline -1` shows the overhaul commit
  - [ ] All changes are committed (no unstaged files)

  **QA Scenarios:**
  ```
  Scenario: Clean git state
    Tool: Bash
    Steps:
      1. git status --porcelain
      2. Assert empty output (clean tree)
    Expected Result: All changes committed
    Evidence: .sisyphus/evidence/task-15-git-status.txt
  ```

  **Commit**: YES
  - Message: `feat: full project overhaul — TUI, CI/CD, coverage, module rename`
  - Files: all changed/added files
  - Pre-commit: `nix-shell -p go --run 'CGO_ENABLED=0 go test ./...'`
## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Rejection → fix → re-run.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `nix-shell -p go --run 'CGO_ENABLED=0 go vet ./...'` + `nix-shell -p go --run 'CGO_ENABLED=0 go test ./...'`. Review all changed files for: `as any`/`@ts-ignore` (N/A for Go), empty catches, fmt.Println in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Vet [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Build the binary. Run `bip32-ssh-keygen generate` and verify output. Run `echo "abandon..." | bip32-ssh-keygen derive --output /tmp/test_key --force` and verify key files. Run `bip32-ssh-keygen derive --generate --output /tmp/auto_key --force` and verify auto-generate. Run `make build`, `make test`, `make clean`. Verify `goreleaser check` passes. Save evidence to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

After all tasks complete and final verification passes:
- Single atomic commit: `feat: full project overhaul — TUI, CI/CD, coverage, module rename`
- Files: all changed/added files
- Pre-commit: `nix-shell -p go --run 'CGO_ENABLED=0 go test ./...'`

---

## Success Criteria

### Verification Commands
```bash
nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'           # Expected: exit 0
nix-shell -p go --run 'CGO_ENABLED=0 go test -v ./...'         # Expected: all PASS
nix-shell -p go --run 'CGO_ENABLED=0 go vet ./...'             # Expected: exit 0
grep -r "github.com/valerius/" --include="*.go" . | wc -l      # Expected: 0
nix-shell -p goreleaser --run 'goreleaser check'                # Expected: exit 0
make build && make test && make clean                            # Expected: all exit 0
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"      # Expected: no error
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" # Expected: no error
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass
- [ ] Coverage ≥95% per package
- [ ] Clean git commit
