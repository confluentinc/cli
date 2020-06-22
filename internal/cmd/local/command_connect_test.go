package local

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJSON(t *testing.T) {
	req := require.New(t)

	req.True(isJSON([]byte("{ \"key\": \"val\" }")))
	req.False(isJSON([]byte("Hello, World!")))
}
