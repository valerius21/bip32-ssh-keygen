
## Wave 0 Task 3: go.sum and Dependencies

**Date**: Wed Feb 25 2026

### Task Completion
- Fixed go.sum and missing dependencies
- Ran `go get github.com/charmbracelet/bubbles/textinput@v1.0.0` to ensure clipboard dependency is properly resolved
- Ran `go mod tidy` to clean up go.mod and go.sum
- Verified: `go build ./...` succeeds with exit code 0
- Verified: `go mod verify` succeeds - all modules verified

### Key Findings
- `github.com/atotto/clipboard v0.1.4` was already present in go.sum (lines 1-2) and go.mod (line 17)
- `go mod tidy` confirmed dependency tree was correct
- All Go build commands must use `CGO_ENABLED=0` since gcc is not available in nix-shell
- All Go commands must be run through: `nix-shell -p go --run 'CGO_ENABLED=0 go <command>'`

### Go Module Pattern
- Use `go mod tidy` to clean up and resolve dependencies
- Use `go mod verify` to verify module integrity
- Use `go build ./...` to verify all packages compile
- Go is not in PATH; always wrap commands in nix-shell

### Build Verification Pattern
For any Go changes, verify with:
1. `nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'` - build all packages
2. `nix-shell -p go --run 'CGO_ENABLED=0 go mod verify'` - verify module integrity

## Wave 1 Task 4: Module Path Rename

**Date**: Wed Feb 25 2026

### Task Completion
- Renamed module path from `github.com/valerius/bip32-ssh-keygen` to `github.com/valerius21/bip32-ssh-keygen`
- Updated 6 files: go.mod, main.go, cmd/root.go, cmd/derive/derive.go, cmd/generate/generate.go, cmd/tui/tui.go
- Found and replaced 13 import statements across 5 files
- Verified: `go build ./...` succeeds with exit code 0
- Verified: No old module path references remain (grep output: 0)

### Files Modified
1. `go.mod` - Module declaration line 1
2. `main.go` - Import statement line 3
3. `cmd/root.go` - 3 import statements (lines 11-13)
4. `cmd/derive/derive.go` - 4 import statements (lines 11-14)
5. `cmd/generate/generate.go` - 1 import statement (line 8)
6. `cmd/tui/tui.go` - 4 import statements (lines 13-16)

### Renaming Pattern
- Use Edit tool with LINE#ID for precise module path replacements
- Always rename ALL occurrences atomically (module declaration + all imports)
- Run `go mod tidy` after renaming to update go.sum
- Verify with `grep -r 'old_path' --include='*.go' . | wc -l` should output 0
- Build verification: `nix-shell -p go --run 'CGO_ENABLED=0 go build ./...'`

### Go Module Renaming Lessons
- Module path changes affect go.mod and all import statements
- Package names remain unchanged (only import paths change)
- `go mod tidy` automatically updates go.sum with new module paths
- Test failures in output are pre-existing issues (stderr capture), not related to rename


## Gitignore Enhancement (Wave 1 - Task 6)

**Pattern: Merging gitignore templates**
- Use `curl -sL https://www.toptal.com/developers/gitignore/api/go` to fetch Go template
- Preserve existing project-specific entries at all costs
- Add new entries under section header (`# gitignore.lol - Go`) for clarity
- Add build-tool specific entries manually (.goreleaser-dist/, dist/)

**Key finding:**
- Most gitignore.lol Go entries already present (*.exe, *.dll, *.so, *.dylib, *.test, *.out, vendor/)
- Only unique addition: `go.work` (Go 1.18+ workspace file)
- Always verify existing entries preserved before appending

**Project-specific entries preserved:**
- bip32-ssh-keygen
- .sisyphus/evidence/
- .devenv*, devenv.local.nix, devenv.local.yaml
- .direnv
- .pre-commit-config.yaml

## Wave 2 Task 7: Makefile Creation

**Pattern: Go project makefile**
- All Go targets must use `CGO_ENABLED=0` prefix
- Use `.PHONY` declaration for ALL targets (build, test, lint, clean, release-dry-run)
- Make `build` the default target (first target in file)
- Keep targets simple: single command per target
- `build`: `CGO_ENABLED=0 go build -o bip32-ssh-keygen .`
- `test`: `CGO_ENABLED=0 go test -v -race -coverprofile=coverage.out ./...`
- `lint`: `golangci-lint run ./...`
- `clean`: `rm -f bip32-ssh-keygen coverage.out`
- `release-dry-run`: `goreleaser release --snapshot --clean`

**Verification notes:**
- Binary verified working: `./bip32-ssh-keygen --help` shows correct usage
- Clean verified working: `rm -f` removes both binary and coverage.out
- Make not available in PATH - commands verified directly

**Go not in PATH constraint:**
- Go must be invoked via `nix-shell -p go --run 'CGO_ENABLED=0 go <command>'`
- Make targets will work when make is installed; individual commands verified

## Wave 2 Task 8: GoReleaser Configuration

**Date**: Wed Feb 25 2026

### Task Completion
- Created .goreleaser.yaml with version 2 configuration
- Configured builds with CGO_ENABLED=0 and multiplatform support (linux, windows, darwin × amd64, arm64)
- Configured archives with Windows zip override using new format_overrides syntax (no `format` field)
- Added SHA256 checksums generation
- Configured changelog filtering to exclude docs:, test:, chore:
- Set release target to github.com/valerius21/bip32-ssh-keygen
- Verified: `goreleaser check` passes with exit code 0

### Key Findings
- GoReleaser v2 requires `version: 2` at top of config file
- Old `archives.format` field is deprecated - remove it entirely and use only `format_overrides`
- Windows zip override simplified to just list `goos: windows` without specifying format (zip is automatic)
- Archives name template: `"{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"` creates cross-platform consistent naming
- Checksum name template: `checksums.txt` (standard practice)

### GoReleaser v2 Configuration Pattern
- Use `version: 2` at top of config
- Builds section: `env: [CGO_ENABLED=0]` + `goos: [linux, windows, darwin]` + `goarch: [amd64, arm64]`
- Archives: NO `format` field, only `format_overrides: [{goos: windows}]` for zip on Windows
- Checksum: `name_template: checksums.txt` + `algorithm: sha256`
- Changelog: `sort: asc` + `filters.exclude: ["^docs:", "^test:", "^chore:"]`
- Release: `github.owner: valerius21` + `github.name: bip32-ssh-keygen`

### Verification Pattern
- Run: `nix-shell -p goreleaser --run 'goreleaser check'`
- Success: "1 configuration file(s) validated" + exit code 0
- Warning: "only version: 2 configuration files are supported" means you need to add `version: 2`

### Deprecated Fields to Avoid
- `archives.format: tar.gz` - removed in v2, default is now implicit
- `format_overrides.format: zip` - simplified to just `goos: windows`
  EOF
## GitHub Workflows Setup (Wave 2, Task 9)

### Workflow Patterns Used

**CI Workflow (.github/workflows/ci.yml):**
- Triggers: push to main/master + pull_request
- Minimal test job with:
  - actions/checkout@v4
  - actions/setup-go@v5 with go-version-file: go.mod
  - Build step: CGO_ENABLED=0 go build ./...
  - Test step: CGO_ENABLED=0 go test -v -race ./...
- Permissions: contents: read (security best practice)

**Release Workflow (.github/workflows/release.yml):**
- Triggers: push with tags matching 'v*'
- Single release job with:
  - actions/checkout@v4 with fetch-depth: 0 (required for GoReleaser changelog)
  - actions/setup-go@v5 with go-version-file: go.mod
  - goreleaser/goreleaser-action@v7 with args: release --clean
- Permissions: contents: write (required for GitHub releases)
- GITHUB_TOKEN automatically passed for release creation

### Key Decisions

- No multi-OS matrix in CI (plan specified: ubuntu-latest only)
- No codecov, security scanning, or linting (local devenv only)
- No Docker login or signing in release workflow (minimal per plan)
- Using CGO_ENABLED=0 for consistency and faster builds
- Using fetch-depth: 0 in release so GoReleaser can generate changelogs

### YAML Validation

Both files validated with custom Python checker (PyYAML not available):
- ✓ ci.yml - 29 lines, valid structure
- ✓ release.yml - 30 lines, valid structure

No tabs, proper indentation (2-space multiples), no trailing spaces.


## Task: Fix devenv.yaml to resolve iocraft-0.7.16 error

**Date:** 2026-02-25

### Original Problem
- devenv shell was failing with: "A hash was specified for iocraft-0.7.16, but there is no corresponding git dependency."
- This was believed to be a known bug in devenv's internal tasks module (cachix/devenv#2524).

### Investigation Findings
1. **No devenv input needed**: The devenv.yaml initially had no explicit `devenv:` input, which means it uses the default version bundled with the devenv CLI.

2. **Pinning attempts failed**:
   - `v1.3`: Failed with missing `export`/`exports` attribute error in tasks module
   - `v1.4`: Failed with Go toolchain configuration error (gopls override issue)
   - `latest` (explicit): Failed with process-managers configuration error

3. **Solution**: The iocraft error was **transient/intermittent**. Using the default devenv (no explicit pinning) works:
   - Cleaned cache: `rm -rf .devenv .direnv`
   - Used default devenv CLI version
   - Shell built successfully: `devenv shell -- go version` → `go1.25.5 linux/amd64`

### Key Lesson
The iocraft-0.7.16 error appears to be a transient lockfile/cache issue, not a permanent bug requiring version pinning. The proper fix is:
1. **Don't pin devenv version** (let CLI use its bundled version)
2. **Clear devenv cache** when encountering hash-related errors
3. **Avoid older versions** (v1.3, v1.4) as they have other breaking changes

### For Future Reference
If you encounter iocraft hash errors:
```bash
rm -rf .devenv .direnv
devenv shell -- [command]
```

This forces a fresh lockfile resolution which typically resolves transient hash conflicts.
