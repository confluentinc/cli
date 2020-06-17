package local

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListConnectors(t *testing.T) {
	req := require.New(t)

	out, err := mockLocalCommand("services", "connect", "connector", "list")
	req.NoError(err)
	req.Contains(out, buildTabbedList(getConnectors()))
}

func TestExtractConfig(t *testing.T) {
	req := require.New(t)

	in := []byte("key1=val1\nkey2=val2\n#commented=val\n")

	out := map[string]string{
		"key1": "val1",
		"key2": "val2",
	}

	req.Equal(out, extractConfig(in))
}

func TestIsJSON(t *testing.T) {
	req := require.New(t)

	req.True(isJSON([]byte("{ \"key\": \"val\" }")))
	req.False(isJSON([]byte("Hello, World!")))
}
