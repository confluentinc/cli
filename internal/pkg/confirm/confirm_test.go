package confirm

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDo(t *testing.T) {
	var out, in bytes.Buffer

	_, err := io.WriteString(&in, "blah\ny\n")
	require.NoError(t, err)
	v := Do(&out, &in, "should do?")
	require.True(t, v)
	require.Equal(t, "should do? (y/n): blah is not a valid choice\nshould do? (y/n): ", out.String())

	_, err = io.WriteString(&in, "no\n")
	require.NoError(t, err)
	v = Do(&out, &in, "should do?")
	require.False(t, v)
}
