package generate

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCmd_Validation(t *testing.T) {
	tests := []struct {
		name           string
		words          string
		expectedError  bool
		errorSubstring string
	}{
		{
			name:          "valid 12 words",
			words:         "12",
			expectedError: false,
		},
		{
			name:          "valid 15 words",
			words:         "15",
			expectedError: false,
		},
		{
			name:          "valid 18 words",
			words:         "18",
			expectedError: false,
		},
		{
			name:          "valid 21 words",
			words:         "21",
			expectedError: false,
		},
		{
			name:          "valid 24 words",
			words:         "24",
			expectedError: false,
		},
		{
			name:           "invalid 11 words",
			words:          "11",
			expectedError:  true,
			errorSubstring: "invalid word count",
		},
		{
			name:           "invalid 13 words",
			words:          "13",
			expectedError:  true,
			errorSubstring: "invalid word count",
		},
		{
			name:           "invalid zero words",
			words:          "0",
			expectedError:  true,
			errorSubstring: "invalid word count",
		},
		{
			name:           "invalid negative words",
			words:          "-1",
			expectedError:  true,
			errorSubstring: "invalid word count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Cmd()
			cmd.SetArgs([]string{"--words", tt.words})
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				require.NoError(t, err)
				output := buf.String()
				// First line should contain the mnemonic (rest is warning)
				lines := strings.Split(output, "\n")
				mnemonic := lines[0]
				words := strings.Fields(mnemonic)
				// Parse the word count from the test case for validation
				expectedCount := 24 // Default when no --words flag specified
				if tt.words != "" {
					fmt.Sscanf(tt.words, "%d", &expectedCount)
				}
				assert.Equal(t, expectedCount, len(words))
			}
		})
	}
}

func TestGenerateCmd_DefaultWords(t *testing.T) {
	cmd := Cmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	// First line should contain the mnemonic
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	mnemonic := lines[0]
	words := strings.Fields(mnemonic)

	assert.Equal(t, 24, len(words), "expected 24 words by default")
}

func TestGenerateCmd_WordCounts(t *testing.T) {
	wordCounts := []int{12, 15, 18, 21, 24}

	for _, count := range wordCounts {
		t.Run(fmt.Sprintf("%d_words", count), func(t *testing.T) {
			cmd := Cmd()
			cmd.SetArgs([]string{"--words", fmt.Sprintf("%d", count)})
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()
			require.NoError(t, err)

			output := strings.TrimSpace(buf.String())
			lines := strings.Split(output, "\n")
			mnemonic := lines[0]
			words := strings.Fields(mnemonic)

			assert.Equal(t, count, len(words))
		})
	}
}

func TestGenerateCmd_FlagShort(t *testing.T) {
	cmd := Cmd()
	cmd.SetArgs([]string{"-w", "12"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	mnemonic := lines[0]
	words := strings.Fields(mnemonic)

	assert.Equal(t, 12, len(words))
}

func TestGenerateCmd_OutputFormat(t *testing.T) {
	cmd := Cmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should contain mnemonic and warning message
	assert.Contains(t, output, "IMPORTANT")
	assert.Contains(t, output, "store them securely")
}
