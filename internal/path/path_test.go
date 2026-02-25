package path

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    []uint32
		wantErr bool
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
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "only m",
			path:    "m",
			wantErr: true,
		},
		{
			name:    "m slash",
			path:    "m/",
			wantErr: true,
		},
		{
			name:    "invalid prefix",
			path:    "abc",
			wantErr: true,
		},
		{
			name:    "non-hardened component",
			path:    "m/44'/22'/0/0",
			wantErr: true,
		},
		{
			name:    "mixed hardened and non-hardened",
			path:    "m/44'/22/0'/0'",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDefault(t *testing.T) {
	want := []uint32{0x8000002C, 0x80000016, 0x80000000, 0x80000000}
	got, err := Parse(DefaultPath)
	if err != nil {
		t.Fatalf("Parse(DefaultPath) failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(DefaultPath) got = %v, want %v", got, want)
	}
}

func TestRejectNonHardened(t *testing.T) {
	_, err := Parse("m/44'/22'/0/0")
	if err == nil {
		t.Error("Parse() should have failed for non-hardened path")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPath(tt.indices)
			if got != tt.want {
				t.Errorf("FormatPath() got = %s, want %s", got, tt.want)
			}
		})
	}
}
