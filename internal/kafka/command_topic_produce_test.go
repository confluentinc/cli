package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
