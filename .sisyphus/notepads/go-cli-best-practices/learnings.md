# Go CLI Best Practices with Cobra - Learnings

Research Date: 2026-02-25

## 1. Command Structure Patterns

### Flat vs Nested Package Organization

**kubectl Pattern (Nested - Enterprise Scale)**
- Main: `pkg/cmd/cmd.go` - NewKubectlCommand() builds root command
- Subcommands: Each in own package (get, create, delete, etc.)
- Pattern: `pkg/cmd/<command>/<command>.go`
- Source: https://github.com/kubernetes/kubectl/blob/73138448cbc640f9551e68254ebff942309a9d9e/pkg/cmd/cmd.go#L164-L249

**gh CLI Pattern (Moderately Nested)**
- Main: `internal/ghcmd/cmd.go` - Main() orchestrates
- Subcommands: `pkg/cmd/<command>/` - Each command as package
- Pattern: `pkg/cmd/<command>/<command>.go`
- Source: https://github.com/cli/cli/blob/cf862d65df7f8ff528015e235c8cccd48cea286f/pkg/cmd/auth/auth.go#L16-L34

**Gum Pattern (Kong-based, Flat)**
- Uses Kong parser instead of Cobra
- Commands as top-level directories (choose, filter, input)
- Each command has: options.go, command.go, TUI logic
- Source: https://github.com/charmbracelet/gum/blob/06d72ec646276e3d0010818afbb10ac089f27c34/gum.go#L24-L22

### Recommendation
For Cobra apps: Follow kubectl's nested pattern for 10+ commands, gh's pattern for smaller apps.

## 2. cmd/ Directory Organization

### Standard kubectl Structure
```
pkg/cmd/
├── cmd.go                    # Root command construction
├── get/                      # Each subcommand as package
│   ├── get.go                # NewCmdGet(), Options, Complete/Validate/Run
│   ├── get_test.go           # Comprehensive tests
│   ├── get_flags.go          # Flag definitions
│   ├── customcolumn.go        # Supporting logic
│   └── sorter.go            # Supporting logic
├── create/                   # Grouped subcommands
│   ├── create.go             # Parent create command
│   ├── create_configmap.go  # Specific subcommands
│   └── create_deployment.go
└── util/                     # Shared utilities
    ├── helpers.go            # CheckErr, UsageErrorf
    └── ...
```

### Key Patterns from kubectl
- Each command has Options struct holding configuration
- Constructor: `NewCmd<Name>()` returns *cobra.Command
- Methods on Options: `Complete()`, `Validate()`, `Run()`
- Test file: `<command>_test.go` in same package
- Source: https://github.com/kubernetes/kubectl/blob/73138448cbc640f9551e68254ebff942309a9d9e/pkg/cmd/get/get.go#L55-L161

### Recommended Structure
```
cmd/
├── root.go                   # Root command definition
├── version.go                # Version command
└── <feature>/               # Feature packages
    ├── <feature>.go          # Command definition (NewCmdFeature)
    ├── <feature>_test.go     # Tests
    ├── options.go            # Options struct
    └── <support>.go         # Supporting logic
```

## 3. Testing Patterns for CLI Commands

### Table-Driven Tests (Kubectl Pattern)

```go
func TestGetObjects(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        wantErr     string
        setup       func(*testing.T, *GetOptions) func()
    }{
        {
            name:    "get pod by name",
            args:    []string{"pod", "test-pod"},
            wantErr: "",
        },
        {
            name:    "invalid resource type",
            args:    []string{"invalid"},
            wantErr: "error",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tf := cmdtesting.NewTestFactory()
            streams := genericiooptions.IOStreams{...}
            cmd := NewCmdGet("kubectl", tf, streams)
            // Execute and verify
        })
    }
}
```
Source: https://github.com/kubernetes/kubectl/blob/73138448cbc640f9551e68254ebff942309a9d9e/pkg/cmd/create/create_configmap_test.go#L30-L80

### Test Organization Patterns

**Kubectl Approach:**
- One test file per command package: `get_test.go`
- Separate files for different components: `customcolumn_test.go`, `sorter_test.go`
- Uses table-driven tests extensively
- Test factory pattern for isolation: `cmdtesting.NewTestFactory()`
- Source: https://github.com/kubernetes/kubectl/blob/73138448cbc640f9551e68254ebff942309a9d9e/pkg/cmd/get/get_test.go#L17-T

**gh CLI Approach:**
- Simpler test files: `login_test.go`
- Focus on happy path and error cases
- Less comprehensive than kubectl

### Best Practices
1. Use table-driven tests for multiple scenarios
2. Separate test concerns into multiple files if complex
3. Use test factories/mocks for dependencies
4. Test both success and error paths
5. Verify exit codes for CLI errors
6. Mock stdio for testing output

## 4. Bubbletea Integration with Cobra

### Pattern 1: Direct Integration (Gum Style)

Gum doesn't use Cobra - uses Kong. But shows TUI pattern:

```go
func (o Options) Run() error {
    // Setup Bubbletea model
    m := model{
        options: o.Options,
        // ... model initialization
    }
    
    // Run TUI
    ctx, cancel := timeout.Context(o.Timeout)
    defer cancel()
    
    tm, err := tea.NewProgram(
        m,
        tea.WithOutput(os.Stderr),
        tea.WithContext(ctx),
    ).Run()
    
    if err != nil {
        return fmt.Errorf("unable to run: %w", err)
    }
    
    // Process result
    m = tm.(model)
    fmt.Println(m.selected)
    return nil
}
```
Source: https://github.com/charmbracelet/gum/blob/06d72ec646276e3d0010818afbb10ac089f27c34/choose/command.go#L152-L164

### Pattern 2: Cobra Wrapper

From community examples (Inngest, etc.):

```go
var interactive bool

var rootCmd = &cobra.Command{
    Use: "myapp",
    Run: func(cmd *cobra.Command, args []string) {
        if interactive {
            // Run Bubbletea TUI
            if err := runTUI(); err != nil {
                log.Fatal(err)
            }
        } else {
            // Run non-interactive logic
            if err := runNonInteractive(); err != nil {
                log.Fatal(err)
            }
        }
    },
}
```

### Integration Considerations

1. **Output Management**: TUI often writes to stderr, stdout for results
   - Gum: `tea.WithOutput(os.Stderr)` then `fmt.Println()` to stdout
   - Source: https://github.com/charmbracelet/gum/blob/06d72ec646276e3d0010818afbb10ac089f27c34/choose/command.go#L156-L180

2. **Timeout Handling**: Use context for cancellation
   ```go
   ctx, cancel := timeout.Context(o.Timeout)
   defer cancel()
   tea.WithContext(ctx)
   ```

3. **Error Handling**: Distinguish TUI errors from user cancellation
   ```go
   if errors.Is(err, tea.ErrInterrupted) {
       os.Exit(exit.StatusAborted)
   }
   ```
   Source: https://github.com/charmbracelet/gum/blob/06d72ec646276e3d0010818afbb10ac089f27c34/main.go#L81-L82

4. **Flag for Interaction Mode**: Allow disabling TUI for automation
   ```go
   var interactive bool
   cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")
   ```

### Best Practices
- Keep TUI logic in separate packages (choose/, filter/ in Gum)
- Use context for timeout and cancellation
- Write TUI to stderr, results to stdout (Unix philosophy)
- Provide flags to disable TUI for scripting
- Handle ErrInterrupted gracefully

## 5. Error Handling Patterns in CLI Tools

### kubectl Pattern: CheckErr Utility

```go
func CheckErr(err error) {
    checkErr(err, fatalErrHandler)
}

func checkErr(err error, handler func(string, int)) {
    if err == nil {
        return
    }
    
    switch {
    case errors.Is(err, ErrExit):
        // Silent exit
        handler("", DefaultErrorExitCode)
    case apierrors.IsInvalid(err):
        // User-friendly error for validation errors
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        handler("", DefaultErrorExitCode)
    default:
        // Generic error
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        handler("", DefaultErrorExitCode)
    }
}
```
Source: https://github.com/kubernetes/kubectl/blob/73138448cbc640f9551e68254ebff942309a9d9e/pkg/cmd/util/helpers.go#L111-L125

### UsageErrorf Pattern

```go
func UsageErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
    msg := fmt.Sprintf(format, args...)
    return fmt.Errorf("%s\nSee '%s -h' for help and examples", msg, cmd.CommandPath())
}
```
Usage: `return cmdutil.UsageErrorf(cmd, "NAME is required")`
Source: https://github.com/kubernetes/kubectl/blob/73138448cbc640f9551e68254ebff942309a9d9e/pkg/cmd/util/helpers.go#L334-L337

### Command Execution Pattern (kubectl)

```go
cmd := &cobra.Command{
    Use: "get [TYPE] [NAME]",
    Run: func(cmd *cobra.Command, args []string) {
        cmdutil.CheckErr(o.Complete(f, cmd, args))
        cmdutil.CheckErr(o.Validate())
        cmdutil.CheckErr(o.Run(f, args))
    },
}
```
Source: https://github.com/kubernetes/kubectl/blob/73138448cbc640f9551e68254ebff942309a9d9e/pkg/cmd/get/get.go#L172-L174

### gh CLI Pattern: Exit Codes

```go
const (
    exitOK      exitCode = 0
    exitError   exitCode = 1
    exitCancel  exitCode = 2
    exitAuth    exitCode = 4
    exitPending exitCode = 8
)

func Main() exitCode {
    rootCmd, err := root.NewCmdRoot(...)
    if err != nil {
        return exitError
    }
    
    if err := rootCmd.ExecuteContextC(ctx); err != nil {
        var noResultsError cmdutil.NoResultsError
        var authError *root.AuthError
        
        if errors.Is(err, cmdutil.SilentError) {
            return exitError
        } else if errors.Is(err, cmdutil.PendingError) {
            return exitPending
        } else if errors.As(err, &authError) {
            return exitAuth
        } else if errors.As(err, &noResultsError) {
            return exitOK  // No results is not failure
        }
        
        return exitError
    }
    return exitOK
}
```
Source: https://github.com/cli/cli/blob/cf862d65df7f8ff528015e235c8cccd48cea286f/internal/ghcmd/cmd.go#L32-L140

### Error Handling Best Practices

1. **Use CheckErr for fatal errors**: Centralized error handling with exit
2. **Return errors from non-fatal operations**: Validate() returns errors
3. **UsageErrorf for argument errors**: Shows help message
4. **Meaningful exit codes**: Different codes for different failures
5. **Distinguish user errors vs system errors**: User errors get friendly messages
6. **Log debug info with klog**: Use V() levels for debug output
7. **Aggregate errors**: Use utilerrors.NewAggregate() for multiple errors

### Error Type Examples

```go
// Usage errors (show help)
return cmdutil.UsageErrorf(cmd, "NAME is required")

// Validation errors (user-friendly)
return fmt.Errorf("invalid value for --flag: %v", value)

// System errors (technical)
return fmt.Errorf("failed to connect: %w", err)

// Silent exit (no error message)
return cmdutil.ErrExit
```

## Key Takeaways

1. **Structure**: Use nested packages (kubectl) for complex CLIs, moderate nesting for simpler ones
2. **Organization**: One package per command with Options struct pattern
3. **Testing**: Table-driven tests, test factories, comprehensive coverage
4. **TUI Integration**: Keep TUI in separate packages, use context for timeout, handle interrupts
5. **Error Handling**: CheckErr utility, meaningful exit codes, user-friendly messages

## Resources

- kubectl: https://github.com/kubernetes/kubectl
- gh CLI: https://github.com/cli/cli
- Gum: https://github.com/charmbracelet/gum
- Cobra Docs: https://cobra.dev/docs/
- Bubbletea: https://github.com/charmbracelet/bubbletea
