package path

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    []uint32
		wantErr bool
		errMsg  string
	}{
		{
			name: "default path",
			path: "m/44'/22'/0'/0'",
			want: []uint32{0x8000002C, 0x80000016, 0x80000000, 0x80000000},
		},
		{
			name: "hardened indices",
			path: "m/44'/22'/1'/0'",
			want: []uint32{0x8000002C, 0x80000016, 0x80000001, 0x80000000},
		},
		{
			name: "single component",
			path: "m/0'",
			want: []uint32{0x80000000},
		},
		{
			name: "multiple components",
			path: "m/44'/0'/0'/0/0", // Non-hardened at end should fail
			wantErr: true,
			errMsg:  "non-hardened",
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "empty path",
		},
		{
			name:    "only m",
			path:    "m",
			wantErr: true,
			errMsg:  "invalid path prefix",
		},
		{
			name:    "m slash",
			path:    "m/",
			wantErr: true,
			errMsg:  "no components",
		},
		{
			name:    "invalid prefix",
			path:    "abc",
			wantErr: true,
			errMsg:  "invalid path prefix",
		},
		{
			name:    "non-hardened component",
			path:    "m/44'/22'/0/0",
			wantErr: true,
			errMsg:  "non-hardened",
		},
		{
			name:    "mixed hardened and non-hardened",
			path:    "m/44'/22/0'/0'",
			wantErr: true,
			errMsg:  "non-hardened",
		},
		{
			name:    "empty component",
			path:    "m/44'//0'/0'",
			wantErr: true,
			errMsg:  "empty component",
		},
		{
			name:    "invalid number",
			path:    "m/44'/abc'/0'/0'",
			wantErr: true,
			errMsg:  "invalid component",
		},
		{
			name:    "negative number",
			path:    "m/-1'/0'/0'/0'",
			wantErr: true,
			errMsg:  "invalid component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParse_DefaultPath(t *testing.T) {
	want := []uint32{0x8000002C, 0x80000016, 0x80000000, 0x80000000}
	got, err := Parse(DefaultPath)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestParse_HardenedBit(t *testing.T) {
	// Verify HardenedBit constant
	assert.Equal(t, uint32(0x80000000), HardenedBit)

	// Parse a path and verify all indices have the hardened bit set
	indices, err := Parse("m/0'/1'/2'/3'")
	require.NoError(t, err)

	for i, idx := range indices {
		assert.True(t, idx&HardenedBit != 0, "index %d should have hardened bit set", i)
		assert.Equal(t, uint32(i), idx^HardenedBit, "index value should be %d", i)
	}
}

func TestFormatPath(t *testing.T) {
	tests := []struct {
		name    string
		indices []uint32
		want    string
	}{
		{
			name:    "default path",
			indices: []uint32{0x8000002C, 0x80000016, 0x80000000, 0x80000000},
			want:    "m/44'/22'/0'/0'",
		},
		{
			name:    "single component",
			indices: []uint32{0x80000000},
			want:    "m/0'",
		},
		{
			name:    "multiple components",
			indices: []uint32{0x80000000, 0x80000001, 0x80000002},
			want:    "m/0'/1'/2'",
		},
		{
			name:    "empty indices",
			indices: []uint32{},
			want:    "m",
		},
		{
			name:    "large indices",
			indices: []uint32{0xffffffff},
			want:    "m/2147483647'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPath(tt.indices)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseAndFormatRoundTrip(t *testing.T) {
	tests := []string{
		"m/44'/22'/0'/0'",
		"m/0'",
		"m/44'/0'/0'/0'",
		"m/1'/2'/3'/4'",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			indices, err := Parse(path)
			require.NoError(t, err)

			reconstructed := FormatPath(indices)
			assert.Equal(t, path, reconstructed)
		})
	}
}

func TestFormatAndParseRoundTrip(t *testing.T) {
	tests := [][]uint32{
		{0x8000002C, 0x80000016, 0x80000000, 0x80000000},
		{0x80000000},
		{0x80000000, 0x80000001, 0x80000002},
	}

	for _, indices := range tests {
		t.Run(FormatPath(indices), func(t *testing.T) {
			path := FormatPath(indices)
			parsed, err := Parse(path)
			require.NoError(t, err)
			assert.True(t, reflect.DeepEqual(indices, parsed))
		})
	}
}
