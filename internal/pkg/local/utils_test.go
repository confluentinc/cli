package local

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildTabbedList(t *testing.T) {
	req := require.New(t)

	arr := []string{"a", "b"}
	out := "  a\n  b\n"
	req.Equal(out, BuildTabbedList(arr))
}

func TestExtractConfig(t *testing.T) {
	req := require.New(t)

	in := []byte("key1=val1\nkey2=val2\n#commented=val\n")

	out := map[string]string{
		"key1": "val1",
		"key2": "val2",
	}

	req.Equal(out, ExtractConfig(in))
}
