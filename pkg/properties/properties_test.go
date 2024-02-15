package properties

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLines_Empty(t *testing.T) {
	lines := parseLines("")
	require.Empty(t, lines)
}

func TestParseLines_Comment(t *testing.T) {
	lines := parseLines("#key=val")
	require.Empty(t, lines)
}

func TestParseLines_Basic(t *testing.T) {
	lines := parseLines("key=val")
	require.Equal(t, []string{"key=val"}, lines)
}

func TestParseLines_TrimSpace(t *testing.T) {
	lines := parseLines("  key=val  ")
	require.Equal(t, []string{"key=val"}, lines)
}

func TestParseLines_MultilineProperties(t *testing.T) {
	lines := parseLines("key=line1\\\nline2")
	require.Equal(t, []string{"key=line1line2"}, lines)
}

func TestParseLines_MultipleLines(t *testing.T) {
	lines := parseLines("key1=val1\nkey2=val2")
	require.Equal(t, []string{"key1=val1", "key2=val2"}, lines)
}

func TestToMap_Basic(t *testing.T) {
	m, err := ConfigSliceToMap([]string{"key=val"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val"}, m)
}

func TestToMap_Override(t *testing.T) {
	m, err := ConfigSliceToMap([]string{"key=val1", "key=val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val2"}, m)
}

func TestToMap_Error(t *testing.T) {
	_, err := ConfigSliceToMap([]string{"string without equal sign"})
	require.Error(t, err)
}

func TestConfigFlagToMap_Basic(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val"}, m)
}

func TestConfigFlagToMap_Override(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1", "key=val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val2"}, m)
}

func TestConfigFlagToMap_Error(t *testing.T) {
	_, err := ConfigFlagToMap([]string{"string without equal sign"})
	require.Error(t, err)
}

func TestConfigFlagToMap_ValueWithComma(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1", "val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1,val2"}, m)
}

func TestConfigFlagToMap_ValueWithEquals(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1=val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1=val2"}, m)
}

func TestCreateKeyValuePairsEmptyMap(t *testing.T) {
	m := make(map[string]string)
	require.Equal(t, "", CreateKeyValuePairs(m))
}

func TestCreateKeyValuePairsSingleKeyValue(t *testing.T) {
	m := make(map[string]string)
	m["k1"] = "v1"
	require.Equal(t, "\"k1\"=\"v1\"\n", CreateKeyValuePairs(m))
}

func TestCreateKeyValuePairsMultipleKeyValue(t *testing.T) {
	m := make(map[string]string)
	m["k1"] = "v1"
	m["k2"] = "v2"
	require.Equal(t, "\"k1\"=\"v1\"\n\"k2\"=\"v2\"\n", CreateKeyValuePairs(m))
}

func TestCreateKeyValuePairsKeysWithDotsAndSorts(t *testing.T) {
	m := make(map[string]string)
	m["link.mode"] = "BIDIRECTIONAL"
	m["connection.mode"] = "OUTBOUND"
	require.Equal(t, "\"connection.mode\"=\"OUTBOUND\"\n\"link.mode\"=\"BIDIRECTIONAL\"\n", CreateKeyValuePairs(m))
}

func TestConfigFlagToMapWithJavaPropertyParsing_Basic(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key=val"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val"}, m)
}

func TestConfigFlagToMapWithJavaPropertyParsing_Override(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key=val1", "key=val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val2"}, m)
}

func TestConfigFlagToMapWithJavaPropertyParsing_WithSpaceAsSeparator(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key val"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val"}, m)

	m2, err2 := ConfigFlagToMapWithJavaPropertyParsing([]string{"key val1 val2"})
	require.NoError(t, err2)
	require.Equal(t, map[string]string{"key": "val1 val2"}, m2)
}

func TestConfigFlagToMapWithJavaPropertyParsing_KeyOnly(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": ""}, m)

	m2, err2 := ConfigFlagToMapWithJavaPropertyParsing([]string{"key1", "key2"})
	require.NoError(t, err2)
	require.Equal(t, map[string]string{"key1": "", "key2": ""}, m2)
}

func TestConfigFlagToMapWithJavaPropertyParsing_Quote(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key=\"val\""})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "\"val\""}, m)
}

func TestConfigFlagToMapWithJavaPropertyParsing_MultipleValues(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key=val1 val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1 val2"}, m)
}

func TestConfigFlagToMapWithJavaPropertyParsing_CSV(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key=val1,val2,val3"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1,val2,val3"}, m)
}

func TestConfigFlagToMapWithJavaPropertyParsing_MultiLine(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key=val1\\\nval2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1val2"}, m)
}

func TestConfigFlagToMapWithJavaPropertyParsing_NextLine(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key1=val1\nkey2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key1": "val1", "key2": ""}, m)
}

func TestConfigFlagToMapWithJavaPropertyParsing_EqualInValue(t *testing.T) {
	m, err := ConfigFlagToMapWithJavaPropertyParsing([]string{"key=username=\"xyx\" password=\"123\""})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "username=\"xyx\" password=\"123\""}, m)
}
