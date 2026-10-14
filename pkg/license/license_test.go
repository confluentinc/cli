package license

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveLicense(t *testing.T) {
	const fileContents = "header.file-payload.signature"

	dir := t.TempDir()
	file := filepath.Join(dir, "license.jwt")
	require.NoError(t, os.WriteFile(file, []byte(fileContents+"\n"), 0600))

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "inline license", input: "header.inline-payload.signature", want: "header.inline-payload.signature"},
		{name: "inline license trimmed", input: "  header.inline-payload.signature\n", want: "header.inline-payload.signature"},
		{name: "file path reads contents", input: file, want: fileContents},
		{name: "empty", input: "", wantErr: true},
		{name: "whitespace only", input: "   ", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveLicense(test.input)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			// A file path must resolve to the file's contents, never the path itself.
			require.Equal(t, test.want, got)
		})
	}
}
