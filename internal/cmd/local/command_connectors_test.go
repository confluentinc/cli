package local

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListConnectors(t *testing.T) {
	req := require.New(t)

	out, err := mockLocalCommand("connectors", "list")
	req.NoError(err)
	req.Contains(out, buildTabbedList(connectors))
}
