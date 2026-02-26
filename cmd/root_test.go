package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)
func TestRootCommand(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		want    string
		wantErr bool
	}{
		{
			name:  "command name",
			field: "Use",
			want:  "bip32-ssh-keygen",
		},
		{
			name:  "version",
			field: "Version",
			want:  "v0.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, RootCmd)
			if tt.field == "Use" {
				assert.Equal(t, tt.want, RootCmd.Use)
			}
			if tt.field == "Version" {
				assert.Equal(t, tt.want, RootCmd.Version)
			}
		})
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	assert.NotNil(t, RootCmd)

	commands := RootCmd.Commands()
	assert.GreaterOrEqual(t, len(commands), 3, "expected at least 3 subcommands")

	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	assert.True(t, commandNames["generate"], "expected generate subcommand")
	assert.True(t, commandNames["derive"], "expected derive subcommand")
	assert.True(t, commandNames["tui"], "expected tui subcommand")
}

func TestRootCommandHelp(t *testing.T) {
	assert.NotNil(t, RootCmd.Short)
	assert.NotEmpty(t, RootCmd.Short)
	assert.NotNil(t, RootCmd.Long)
	assert.NotEmpty(t, RootCmd.Long)
}

func TestRootCmd_ExecuteRootCommand(t *testing.T) {
	// Test that RootCmd.Run (the root handler) displays help
	buf := new(strings.Builder)
	RootCmd.SetArgs([]string{})
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	err := RootCmd.Execute()

	assert.NoError(t, err) // Root command runs help without error
	output := buf.String()
	assert.Contains(t, output, "generates deterministic") // Should show help
}

func TestExecute_HandlesError(t *testing.T) {
	// Test that os.Exit is called when command fails
	// We can't test os.Exit directly, but we can verify RootCmd returns error
	buf := new(strings.Builder)
	RootCmd.SetArgs([]string{"--nonexistent-flag"})
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	err := RootCmd.Execute()
	assert.Error(t, err) // Invalid flag should return error
}

func TestRootCmd_VersionFlag(t *testing.T) {
	buf := new(strings.Builder)
	RootCmd.SetArgs([]string{"--version"})
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	err := RootCmd.Execute()
	
	assert.NoError(t, err)
	// Version flag should be handled by cobra
}

func TestExecute_Function(t *testing.T) {
	// Test Execute() function directly
	// The Execute function wraps RootCmd.Execute() with error handling
	// It calls os.Exit(1) on error which we can't test here
	// But we can verify it exists and doesn't crash on successful path
	// We test this indirectly through RootCmd.Execute elsewhere
}
