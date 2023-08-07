package testserver

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
)

const SRApiEnvId = "env-srUpdate"

// Handler for: "/api/schema_registries"
func handleSchemaRegistries(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			q := r.URL.Query()

			id := q.Get("id")
			if id == "" {
				id = "lsrc-1234"
			}

			createReply := &ccloudv1.CreateSchemaRegistryClusterReply{Cluster: &ccloudv1.SchemaRegistryCluster{
				Id:                    id,
				AccountId:             q.Get("account_id"),
				Name:                  "account schema-registry",
				Endpoint:              TestSchemaRegistryUrl.String(),
				ServiceProvider:       "aws",
				ServiceProviderRegion: "us-west-2",
				Package:               "free",
			}}

			b, err := ccloudv1.MarshalJSONToBytes(createReply)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}
