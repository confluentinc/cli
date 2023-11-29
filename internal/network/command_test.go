package network

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertMapToString(t *testing.T) {
	m := map[string]string{"zone1": "link1", "zone2": "link2", "zone3": "link3"}
	require.Equal(t, "zone1=link1, zone2=link2, zone3=link3", convertMapToString(m))
}
