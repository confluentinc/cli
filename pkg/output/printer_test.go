package output

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLinkRegexp(t *testing.T) {
	require.Equal(t, " \n", linkRegexp.ReplaceAllString("https://docs.confluent.io/current/cli/index.html \n", ""))
	require.Equal(t, " and", linkRegexp.ReplaceAllString("https://docs.docker.com/engine/install/ and", ""))
}
