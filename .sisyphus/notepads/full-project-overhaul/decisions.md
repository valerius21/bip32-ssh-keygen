
## Task 11 Decisions

### derive.go os.Stderr usage:
**Decision**: Keep using `os.Stderr` instead of refactoring to `cmd.ErrOrStderr()`
**Reason**: Task explicitly forbids refactoring TTY detection for testability. Tests focus on functional behavior (key file creation) rather than output capture.

### TUI View methods:
**Decision**: Do not attempt to test View() methods
**Reason**: Task explicitly forbids installing teatest framework. Focus on logic coverage (Update, handleEnter, Init, performDerivation).

### Execute() function testability:
**Decision**: Accept 0% coverage for cmd/root Execute()
**Reason**: The function wraps RootCmd.Execute() with error handling and os.Exit(1). Testing this requires mocking os.Exit which isn't practical. Tested via RootCmd.Execute() directly instead.

### derive TTY path coverage:
**Decision**: Accept 80.4% coverage instead of 85% target
**Reason**: TTY detection path (~20% of code) can't be covered without a real TTY device. The non-TTY stdin path is well-tested.

### generate validation tests:
**Decision**: Parse first line of output for word count validation
**Reason**: Output includes mnemonic (first line) followed by warning messages. Need to split by newline before counting words.

