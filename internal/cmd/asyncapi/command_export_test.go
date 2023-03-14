package asyncapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlePrimitiveSchemas(t *testing.T) {
	unmarshalledSchema, err := handlePrimitiveSchemas(`"string"`, fmt.Errorf("unable to unmarshal schema"))
	require.NoError(t, err)
	require.Equal(t, unmarshalledSchema, map[string]any{"type": "string"})
	_, err = handlePrimitiveSchemas("Invalid_schema", fmt.Errorf("unable to marshal schema"))
	require.Error(t, err)
}

func TestMsgName(t *testing.T) {
	require.Equal(t, "TopicNameMessage", msgName("topic name"))
}
