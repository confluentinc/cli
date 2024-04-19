package apikey

import (
	"testing"

	"github.com/stretchr/testify/require"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
)

func TestGetResourceType(t *testing.T) {
	require.Equal(t, "cloud", getResourceType(apikeysv2.ObjectReference{Kind: apikeysv2.PtrString("Cloud")}))
	require.Equal(t, "flink-region", getResourceType(apikeysv2.ObjectReference{ApiVersion: apikeysv2.PtrString("fcpm/v2"), Kind: apikeysv2.PtrString("Region")}))
	require.Equal(t, "kafka", getResourceType(apikeysv2.ObjectReference{ApiVersion: apikeysv2.PtrString("cmk/v2"), Kind: apikeysv2.PtrString("Cluster")}))
	require.Equal(t, "ksql", getResourceType(apikeysv2.ObjectReference{Kind: apikeysv2.PtrString("ksqlDB")}))
	require.Equal(t, "schema-registry", getResourceType(apikeysv2.ObjectReference{Kind: apikeysv2.PtrString("SchemaRegistry")}))
}
