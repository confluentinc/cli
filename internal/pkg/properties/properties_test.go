package properties

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLines_Empty(t *testing.T) {
	t.Parallel()

	lines := parseLines("")
	require.Empty(t, lines)
}

func TestParseLines_Comment(t *testing.T) {
	t.Parallel()

	lines := parseLines("#key=val")
	require.Empty(t, lines)
}

func TestParseLines_Basic(t *testing.T) {
	t.Parallel()

	lines := parseLines("key=val")
	require.Equal(t, []string{"key=val"}, lines)
}

func TestParseLines_TrimSpace(t *testing.T) {
	t.Parallel()

	lines := parseLines("  key=val  ")
	require.Equal(t, []string{"key=val"}, lines)
}

func TestParseLines_MultilineProperties(t *testing.T) {
	t.Parallel()

	lines := parseLines("key=line1\\\nline2")
	require.Equal(t, []string{"key=line1line2"}, lines)
}

func TestParseLines_MultipleLines(t *testing.T) {
	t.Parallel()

	lines := parseLines("key1=val1\nkey2=val2")
	require.Equal(t, []string{"key1=val1", "key2=val2"}, lines)
}

func TestToMap_Basic(t *testing.T) {
	t.Parallel()

	m, err := toMap([]string{"key=val"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val"}, m)
}

func TestToMap_Override(t *testing.T) {
	t.Parallel()

	m, err := toMap([]string{"key=val1", "key=val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val2"}, m)
}

func TestToMap_Error(t *testing.T) {
	t.Parallel()

	_, err := toMap([]string{"string without equal sign"})
	require.Error(t, err)
}

func TestConfigFlagToMap_Basic(t *testing.T) {
	t.Parallel()

	m, err := ConfigFlagToMap([]string{"key=val"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val"}, m)
}

func TestConfigFlagToMap_Override(t *testing.T) {
	t.Parallel()

	m, err := ConfigFlagToMap([]string{"key=val1", "key=val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val2"}, m)
}

func TestConfigFlagToMap_Error(t *testing.T) {
	t.Parallel()

	_, err := ConfigFlagToMap([]string{"string without equal sign"})
	require.Error(t, err)
}

func TestConfigFlagToMap_ValueWithComma(t *testing.T) {
	t.Parallel()

	m, err := ConfigFlagToMap([]string{"key=val1", "val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1,val2"}, m)
}

func TestConfigFlagToMap_ValueWithEquals(t *testing.T) {
	t.Parallel()

	m, err := ConfigFlagToMap([]string{"key=val1=val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1=val2"}, m)
}
