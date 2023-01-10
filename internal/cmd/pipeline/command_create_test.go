package pipeline

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateSecretMappings(t *testing.T) {
	secretMappings, _ := createSecretMappings([]string{"name1=value1", "_name2=value2"}, secretMappingWithoutEmptyValue)
	require.Equal(t, "value1", secretMappings["name1"])
	require.Equal(t, "value2", secretMappings["_name2"])

	secretMappings, err := createSecretMappings([]string{"123invalidName=value"}, secretMappingWithoutEmptyValue)
	require.Error(t, err)

	secretMappings, err = createSecretMappings([]string{"invalidName!@#$=value"}, secretMappingWithoutEmptyValue)
	require.Error(t, err)

	secretMappings, _ = createSecretMappings([]string{"name=value-with-,and="}, secretMappingWithoutEmptyValue)
	require.Equal(t, "value-with-,and=", secretMappings["name"])

	secretMappings, _ = createSecretMappings([]string{"name=value-with-\"and'='"}, secretMappingWithoutEmptyValue)
	require.Equal(t, "value-with-\"and'='", secretMappings["name"])

	secretMappings, _ = createSecretMappings([]string{"name=value with space"}, secretMappingWithoutEmptyValue)
	require.Equal(t, "value with space", secretMappings["name"])

	secretMappings, _ = createSecretMappings([]string{"a_really_really_really_really_really_really_really_really_really_really_really_really_long_secret_name_but_not_exceeding_128_yet=value"}, secretMappingWithoutEmptyValue)
	require.Equal(t, "value", secretMappings["a_really_really_really_really_really_really_really_really_really_really_really_really_long_secret_name_but_not_exceeding_128_yet"])

	secretMappings, err = createSecretMappings([]string{"a_really_really_really_really_really_really_really_really_really_really_really_really_long_secret_name_exceeded_128_characters_limit=value"}, secretMappingWithoutEmptyValue)
	require.Error(t, err)

	// empty secret value is NOT allowed with this regex
	secretMappings, err = createSecretMappings([]string{"name1=value1", "name2="}, secretMappingWithoutEmptyValue)
	require.Error(t, err)

	// empty secret value is allowed with this regex
	secretMappings, _ = createSecretMappings([]string{"name1=value1", "name2="}, secretMappingWithEmptyValue)
	require.Equal(t, "value1", secretMappings["name1"])
	require.Equal(t, "", secretMappings["name2"])
}

func TestSecretNamesList(t *testing.T) {
	names := getOrderedSecretNames(nil)
	require.Equal(t, []string(nil), names)

	secretMappings := make(map[string]string)
	names = getOrderedSecretNames(&secretMappings)
	require.Equal(t, []string(nil), names)

	secretMappings["name1"] = "value1"
	names = getOrderedSecretNames(&secretMappings)
	require.Equal(t, []string{"name1"}, names)

	secretMappings["name3"] = "value3"
	names = getOrderedSecretNames(&secretMappings)
	require.Equal(t, []string{"name1", "name3"}, names)

	secretMappings["name2"] = "value2"
	names = getOrderedSecretNames(&secretMappings)
	require.Equal(t, []string{"name1", "name2", "name3"}, names)
}
