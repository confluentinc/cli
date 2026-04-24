package kafka

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type splitTest struct {
	Data      string
	Delimiter string
	Key       string
	Value     string
}

func TestGetKeyAndValue(t *testing.T) {
	testCases := []splitTest{
		// Different delimiters
		{Data: `{"CustomerId": 1, "Name": "My Name"}:message`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name"};message`, Delimiter: ";", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name"}"message`, Delimiter: `"`, Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name"},message`, Delimiter: ",", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name"}{message`, Delimiter: "{", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name"}}message`, Delimiter: "}", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name"}1message`, Delimiter: "1", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: "message"},

		// Extra spaces
		{Data: `  {"CustomerId": 1, "Name": "My Name"} : message`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: "message"},
		{Data: `{ "CustomerId": 1, "Name": "My Name" }:message`, Delimiter: ":", Key: `{ "CustomerId": 1, "Name": "My Name" }`, Value: "message"},

		// Different values
		{Data: `{"CustomerId": 1}:{"Name": "My Name"}`, Delimiter: ":", Key: `{"CustomerId": 1}`, Value: `{"Name": "My Name"}`},
		{Data: `{"CustomerId": 1, "Name": "My Name"}::`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: ":"},
		{Data: `{"CustomerId": 1, "Name": "My Name"}`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name"}`, Value: ""},

		// JSON special characters inside JSON strings
		{Data: `{"CustomerId": 1, "Name": "My Name}"}:message`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name}"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name\"}"}:message`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name\"}"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name\\\"}"}:message`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name\\\"}"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name\\"}:message`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name\\"}`, Value: "message"},
		{Data: `{"CustomerId": 1, "Name": "My Name\\\\"}:message`, Delimiter: ":", Key: `{"CustomerId": 1, "Name": "My Name\\\\"}`, Value: "message"},

		// Case with all JSON types
		{Data: `{"Key1": "string", "Key2": 42, "Key3": true, "Key4": null, "Key5":{"CustomerId": 1, "Name": "My Name"}, "Key6":["Name1", "Name2"]}:message`, Delimiter: ":", Key: `{"Key1": "string", "Key2": 42, "Key3": true, "Key4": null, "Key5":{"CustomerId": 1, "Name": "My Name"}, "Key6":["Name1", "Name2"]}`, Value: "message"},
	}

	for _, testCase := range testCases {
		key, value, err := getKeyAndValue(true, testCase.Data, testCase.Delimiter)
		assert.NoError(t, err)
		assert.Equal(t, testCase.Key, key)
		assert.Equal(t, testCase.Value, value)
	}
}

func TestGetKeyAndValue_StringKey(t *testing.T) {
	testCases := []splitTest{
		{Data: `key:{"CustomerId": 1, "Name": "My Name"}`, Delimiter: ":", Key: "key", Value: `{"CustomerId": 1, "Name": "My Name"}`},
		{Data: `:{"CustomerId": 1, "Name": "My Name"}`, Delimiter: ":", Key: "", Value: `{"CustomerId": 1, "Name": "My Name"}`},
		{Data: `{"CustomerId": 1, "Name": "My Name"}`, Delimiter: ":", Key: `{"CustomerId"`, Value: `1, "Name": "My Name"}`},
	}

	for _, testCase := range testCases {
		key, value, err := getKeyAndValue(false, testCase.Data, testCase.Delimiter)
		assert.NoError(t, err)
		assert.Equal(t, testCase.Key, key)
		assert.Equal(t, testCase.Value, value)
	}
}

func TestGetKeyAndValue_Fail(t *testing.T) {
	// Missing or malformed key
	testCases := []splitTest{
		{Data: `{"CustomerId": 1, "Name": "My Name"}:message`, Delimiter: ","},
		{Data: `{"CustomerId": 1, "Name": "My Name\\"}"}:message`, Delimiter: ":"},
		{Data: `{"CustomerId": 1, "Name": "My Name}\"}:message`, Delimiter: ":"},
		{Data: `:message`, Delimiter: ":"},
	}

	for _, testCase := range testCases {
		_, _, err := getKeyAndValue(true, testCase.Data, testCase.Delimiter)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), missingOrMalformedKeyErrorMsg)
	}

	// Missing key (non-schema key format)
	testCases = []splitTest{
		{Data: `{"CustomerId": 1, "Name": "My Name"}:message`, Delimiter: "|"},
		{Data: `message`, Delimiter: ":"},
	}

	for _, testCase := range testCases {
		_, _, err := getKeyAndValue(false, testCase.Data, testCase.Delimiter)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), missingKeyOrValueErrorMsg)
	}
}

func TestGetMetaInfoFromSchemaId(t *testing.T) {
	metaInfo := getMetaInfoFromSchemaId(100004)
	require.Equal(t, []byte{0x0, 0x0, 0x1, 0x86, 0xa4}, metaInfo)
}

func TestProduceCommand_NormalizeFlag(t *testing.T) {
	cfg := config.AuthenticatedCloudConfigMock()
	prerunner := &pcmd.PreRun{Config: cfg}
	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(new(cobra.Command), prerunner),
		clientID:                "test-client-id",
	}

	cmd := c.newProduceCommand()

	flag := cmd.Flags().Lookup("normalize")
	require.NotNil(t, flag, "--normalize flag should be registered on `kafka topic produce`")
	assert.Equal(t, "false", flag.DefValue, "--normalize flag should default to false")
	assert.Equal(t, "Alphabetize the list of schema fields when registering a new schema.", flag.Usage)

	normalize, err := cmd.Flags().GetBool("normalize")
	require.NoError(t, err)
	assert.False(t, normalize, "reading --normalize without setting it should return false")

	require.NoError(t, cmd.Flags().Set("normalize", "true"))
	normalize, err = cmd.Flags().GetBool("normalize")
	require.NoError(t, err)
	assert.True(t, normalize, "reading --normalize after setting it to true should return true")
}

func TestHeaders(t *testing.T) {
	t.Run("It should return valid Kafka headers", func(t *testing.T) {
		headers := []string{"contenttype:application/json", "x-request-id:12345"}

		expected := []ckgo.Header{
			{Key: "contenttype", Value: []byte("application/json")},
			{Key: "x-request-id", Value: []byte("12345")},
		}

		parsedHeaders, err := parseHeaders(headers, ":")
		assert.Equal(t, parsedHeaders, expected)
		assert.NoError(t, err)
	})

	t.Run("It should return an invalid headers error when key is missing", func(t *testing.T) {
		headers := []string{":valueOnly"}

		parsedHeaders, err := parseHeaders(headers, ":")
		expectedErrorMsg := fmt.Sprintf(invalidHeadersErrorMsg, ":")

		assert.Error(t, err)
		assert.Equal(t, err.Error(), expectedErrorMsg)
		assert.Nil(t, parsedHeaders)
	})

	t.Run("It should return an invalid headers error when delimiter is incorrect", func(t *testing.T) {
		headers := []string{"asdasdas:valueOnly", "sadsad=sdsadasd"}

		parsedHeaders, err := parseHeaders(headers, ":")
		expectedErrorMsg := fmt.Sprintf(invalidHeadersErrorMsg, ":")

		assert.Error(t, err)
		assert.Equal(t, err.Error(), expectedErrorMsg)
		assert.Nil(t, parsedHeaders)
	})
}
