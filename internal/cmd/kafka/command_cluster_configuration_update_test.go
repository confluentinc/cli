package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatUpdateOutput_Singular(t *testing.T) {
	output := formatUpdateOutput(map[string]string{"a": ""})
	require.Equal(t, `Successfully requested to update configuration "a".`, output)
}

func TestFormatUpdateOutput_Plural(t *testing.T) {
	output := formatUpdateOutput(map[string]string{"a": "", "b": ""})
	require.Equal(t, `Successfully requested to update configurations "a" and "b".`, output)
}
