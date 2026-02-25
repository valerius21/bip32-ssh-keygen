package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedWords  int
		expectedError  bool
		errorSubstring string
	}{
		{
			name:          "default 24 words",
			args:          []string{"generate"},
			expectedWords: 24,
			expectedError: false,
		},
		{
			name:          "explicit 12 words",
			args:          []string{"generate", "--words", "12"},
			expectedWords: 12,
			expectedError: false,
		},
		{
			name:           "invalid word count 13",
			args:           []string{"generate", "--words", "13"},
			expectedError:  true,
			errorSubstring: "invalid word count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RootCmd.SetIn(strings.NewReader("")) // Reset stdin for each test
			RootCmd.SetIn(strings.NewReader("")) // Reset stdin for each test
			RootCmd.SetArgs(nil)                 // Reset args for each test
			buf := new(bytes.Buffer)
			RootCmd.SetOut(buf)
			RootCmd.SetErr(buf)
			RootCmd.SetArgs(tt.args)

			err := RootCmd.Execute()

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				assert.NoError(t, err)
				output := strings.TrimSpace(buf.String())
				// The output might contain the warning message on stderr if we don't separate them,
				// but for now let's just check the first line or the word count of the mnemonic.
				lines := strings.Split(output, "\n")
				mnemonic := lines[0]
				words := strings.Fields(mnemonic)
				assert.Equal(t, tt.expectedWords, len(words))
			}
		})
	}
}
