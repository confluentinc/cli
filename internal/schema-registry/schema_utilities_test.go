package schemaregistry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMetaInfoFromSchemaId(t *testing.T) {
	metaInfo := GetMetaInfoFromSchemaId(100004)
	require.Equal(t, []byte{0x0, 0x0, 0x1, 0x86, 0xa4}, metaInfo)
}
