package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateValidSecretMappings(t *testing.T) {
	t.Parallel()

	secretMappings, err := createSecretMappings([]string{"name1=value1", "_name2=value2"}, secretMappingWithoutEmptyValue)
	assert.NoError(t, err)
	assert.Equal(t, "value1", secretMappings["name1"])
	assert.Equal(t, "value2", secretMappings["_name2"])

	secretMappings, _ = createSecretMappings([]string{"name=value-with-,and="}, secretMappingWithoutEmptyValue)
	assert.Equal(t, "value-with-,and=", secretMappings["name"])

	secretMappings, _ = createSecretMappings([]string{"name=value-with-\"and'='"}, secretMappingWithoutEmptyValue)
	assert.Equal(t, "value-with-\"and'='", secretMappings["name"])

	secretMappings, _ = createSecretMappings([]string{"name=value with space"}, secretMappingWithoutEmptyValue)
	assert.Equal(t, "value with space", secretMappings["name"])
}

func TestCreateSecretMappingsWithInvalidName(t *testing.T) {
	t.Parallel()

	_, err := createSecretMappings([]string{"123invalidName=value"}, secretMappingWithoutEmptyValue)
	assert.Error(t, err)

	_, err = createSecretMappings([]string{"invalidName!@#$=value"}, secretMappingWithoutEmptyValue)
	assert.Error(t, err)
}

func TestCreateSecretMappingsWithLongName(t *testing.T) {
	t.Parallel()

	secretMappings, err := createSecretMappings([]string{"a_really_really_really_really_really_really_really_really_really_really_really_really_long_secret_name_but_not_exceeding_128_yet=value"}, secretMappingWithoutEmptyValue)
	assert.NoError(t, err)
	assert.Equal(t, "value", secretMappings["a_really_really_really_really_really_really_really_really_really_really_really_really_long_secret_name_but_not_exceeding_128_yet"])
}

func TestCreateSecretMappingsWithEmptyValue(t *testing.T) {
	t.Parallel()

	// empty secret value is NOT allowed with this regex
	_, err := createSecretMappings([]string{"name1=value1", "name2="}, secretMappingWithoutEmptyValue)
	assert.Error(t, err)

	// empty secret value is allowed with this regex
	secretMappings, _ := createSecretMappings([]string{"name1=value1", "name2="}, secretMappingWithEmptyValue)
	assert.Equal(t, "value1", secretMappings["name1"])
	assert.Equal(t, "", secretMappings["name2"])
}

func TestSecretNamesListWithEmptyInput(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{}, getOrderedSecretNames(nil))

	secretMappings := make(map[string]string)
	assert.Equal(t, []string{}, getOrderedSecretNames(&secretMappings))
}

func TestSecretNamesListOrder(t *testing.T) {
	t.Parallel()

	secretMappings := make(map[string]string)

	secretMappings["name1"] = "value1"
	assert.Equal(t, []string{"name1"}, getOrderedSecretNames(&secretMappings))

	secretMappings["name3"] = "value3"
	assert.Equal(t, []string{"name1", "name3"}, getOrderedSecretNames(&secretMappings))

	secretMappings["name2"] = "value2"
	assert.Equal(t, []string{"name1", "name2", "name3"}, getOrderedSecretNames(&secretMappings))
}
