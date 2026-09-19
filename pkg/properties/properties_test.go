package properties

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLines_Empty(t *testing.T) {
	lines := ParseLines("")
	require.Empty(t, lines)
}

func TestParseLines_Comment(t *testing.T) {
	lines := ParseLines("#key=val")
	require.Empty(t, lines)
}

func TestParseLines_Basic(t *testing.T) {
	lines := ParseLines("key=val")
	require.Equal(t, []string{"key=val"}, lines)
}

func TestParseLines_TrimSpace(t *testing.T) {
	lines := ParseLines("  key=val  ")
	require.Equal(t, []string{"key=val"}, lines)
}

func TestParseLines_MultilineProperties(t *testing.T) {
	lines := ParseLines("key=line1\\\nline2")
	require.Equal(t, []string{"key=line1line2"}, lines)
}

func TestParseLines_MultipleLines(t *testing.T) {
	lines := ParseLines("key1=val1\nkey2=val2")
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

func TestConfigFlagToMap_NewLine(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1\\nval2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1\nval2"}, m)
}

func TestConfigFlagToMap_CarriageReturn(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1\\rval2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1\rval2"}, m)
}

func TestConfigFlagToMap_Tab(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1\\tval2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1\tval2"}, m)
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

// Regression: behavior must match the StringSlice path for non-JSON configs. StringArray delivers a
// comma-list as ONE element (pflag does not split), so the parser must split it itself.
func TestConfigArrayToMap_CommaSeparatedList(t *testing.T) {
	m, err := GetMapFromArray([]string{"cleanup.policy=compact,compression.type=gzip"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"cleanup.policy": "compact", "compression.type": "gzip"}, m)
}

func TestConfigArrayToMap_ValueWithComma(t *testing.T) {
	m, err := GetMapFromArray([]string{"cleanup.policy=delete,compact"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"cleanup.policy": "delete,compact"}, m)
}

func TestConfigArrayToMap_MultipleFlags(t *testing.T) {
	m, err := GetMapFromArray([]string{"a=1", "b=2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"a": "1", "b": "2"}, m)
}

func TestConfigArrayToMap_Override(t *testing.T) {
	m, err := GetMapFromArray([]string{"retention.ms=1", "retention.ms=2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"retention.ms": "2"}, m)
}

func TestConfigArrayToMap_ValueWithEquals(t *testing.T) {
	m, err := GetMapFromArray([]string{"key=a=b"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "a=b"}, m)
}

func TestConfigArrayToMap_UnescapesNonRawValues(t *testing.T) {
	m, err := GetMapFromArray([]string{`foo=a\nb`})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"foo": "a\nb"}, m)
}

func TestConfigArrayToMap_RawValueKeyPreservesJSON(t *testing.T) {
	json := `{"schema":"{\"type\":\"record\",\"name\":\"TestRecord\",\"doc\":\"Basic test.\\na=b.\",\"fields\":[{\"name\":\"field1\",\"type\":[\"null\",\"string\"]}]}"}`
	m, err := GetMapFromArray([]string{"confluent.value.association=" + json}, "confluent.value.association")
	require.NoError(t, err)
	require.Equal(t, json, m["confluent.value.association"])
}

func TestConfigArrayToMap_JSONWithCommasThenNormalConfig(t *testing.T) {
	m, err := GetMapFromArray(
		[]string{`confluent.key.association={"subject":"x","lifecycle":"STRONG"},retention.ms=500`},
		"confluent.key.association",
	)
	require.NoError(t, err)
	require.Equal(t, `{"subject":"x","lifecycle":"STRONG"}`, m["confluent.key.association"])
	require.Equal(t, "500", m["retention.ms"])
}

func TestConfigArrayToMap_MalformedJSONReturnsError(t *testing.T) {
	_, err := GetMapFromArray([]string{`confluent.value.association={"broken`}, "confluent.value.association")
	require.Error(t, err)
}

// The file path is unchanged and shared with GetMap.
func TestGetMapFromArray_FilePreservesJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topic.properties")
	// The file path stores jsonValueKeys verbatim: a JSON value containing "\n" and "=" must not be un-escaped.
	json := `{"schema":"{\"type\":\"record\",\"name\":\"TestRecord\",\"doc\":\"Basic test.\\na=b.\",\"fields\":[{\"name\":\"field1\",\"type\":[\"null\",\"string\"]}]}"}`
	require.NoError(t, os.WriteFile(path, []byte("confluent.value.association="+json+"\n"), 0o600))

	m, err := GetMapFromArray([]string{path}, "confluent.value.association")
	require.NoError(t, err)
	require.Equal(t, json, m["confluent.value.association"])
}
