package cmd

import (
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
