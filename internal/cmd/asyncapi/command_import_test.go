package asyncapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveSchemaType(t *testing.T) {
	require.Equal(t, resolveSchemaType("avro"), "AVRO")
	require.Equal(t, resolveSchemaType("json"), "JSON")
	require.Equal(t, resolveSchemaType("proto"), "PROTOBUF")
}
