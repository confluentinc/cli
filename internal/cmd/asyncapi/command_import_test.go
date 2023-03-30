package asyncapi

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func TestResolveSchemaType(t *testing.T) {
	assert.Equal(t, resolveSchemaType("avro"), "AVRO")
	assert.Equal(t, resolveSchemaType("json"), "JSON")
	assert.Equal(t, resolveSchemaType("proto"), "PROTOBUF")
}

func TestRetry(t *testing.T) {
	require.Error(t, retry(context.Background(), time.Nanosecond, 2*time.Nanosecond, func() error {
		return errors.New("error")
	}))
	require.NoError(t, retry(context.Background(), time.Nanosecond, 2*time.Nanosecond, func() error {
		return nil
	}))
}
