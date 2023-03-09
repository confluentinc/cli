package schemaregistry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertMapToString(t *testing.T) {
	t.Parallel()

	m := map[string]string{"name": "alice", "phone": "xxx-xxx-xxxx", "age": "20"}
	require.Equal(t, "age=\"20\"\nname=\"alice\"\nphone=\"xxx-xxx-xxxx\"", convertMapToString(m))
}
