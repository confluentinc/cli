package properties

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestConfigFlagToMap_WithoutEqual(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key val"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val"}, m)
}

func TestConfigFlagToMap_ValueWithComma(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1,val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1,val2"}, m)
}

func TestConfigFlagToMap_ValueWithEquals(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1=val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1=val2"}, m)
}

func TestConfigFlagToMap_KeyOnly(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": ""}, m)
}

func TestConfigFlagToMap_Quote(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=\"val\""})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "\"val\""}, m)
}

func TestConfigFlagToMap_MultipleValues(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1 val2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1 val2"}, m)
}

func TestConfigFlagToMap_MultiLine(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=val1\\\nval2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1val2"}, m)
}

func TestConfigFlagToMap_NextLine(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key1=val1\nkey2"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key1": "val1", "key2": ""}, m)
}

func TestConfigFlagToMap_UsernamePassword(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key=username=\"xyx\" password=\"123\""})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "username=\"xyx\" password=\"123\""}, m)
}

func TestConfigFlagToMap_TrimWhiteSpaces(t *testing.T) {
	m, err := ConfigFlagToMap([]string{"key= val "})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val"}, m)
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

func TestGetMap_ExplictNewLineCharacter(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "TestGetMap")
	if err != nil {
		require.NoError(t, err)
		return
	}
	defer os.Remove(file.Name())

	// Write some content to the file
	content := []byte("key=val1\\nval2")
	_, err = file.Write(content)
	if err != nil {
		require.NoError(t, err)
		return
	}
	m, err := GetMap([]string{file.Name()})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "val1\nval2"}, m)
}
