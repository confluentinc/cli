package asyncapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveSchemaType(t *testing.T) {
	assert.Equal(t, resolveSchemaType("avro"), "AVRO")
	assert.Equal(t, resolveSchemaType("json"), "JSON")
	assert.Equal(t, resolveSchemaType("proto"), "PROTOBUF")
}
