package testserver

import (
	"io"
	"net/http"
	"testing"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/stretchr/testify/require"
)

const (
	SRApiEnvId = "env-srUpdate"
)

// Handler for: "/api/schema_registries"
func (c *CloudRouter) HandleSchemaRegistries(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			id = "lsrc-1234"
		}
		accountId := q.Get("account_id")
		var endpoint string
		// for sr commands that use the sr api (use accountId to differentiate) we want to use the SR server URL so that we can make subsequent requests there
		// for describe commands we want to use a standard endpoint so that it will always match the test fixture
		if accountId == SRApiEnvId {
			endpoint = c.srApiUrl
		} else {
			endpoint = "SASL_SSL://sr-endpoint"
		}
		srCluster := &ccloudv1.SchemaRegistryCluster{
			Id:                    id,
			AccountId:             accountId,
			Name:                  "account schema-registry",
			Endpoint:              endpoint,
			ServiceProvider:       "aws",
			ServiceProviderRegion: "us-west-2",
			Package:               "free",
		}
		switch r.Method {
		case http.MethodPost:
			createReply := &ccloudv1.CreateSchemaRegistryClusterReply{Cluster: srCluster}
			b, err := ccloudv1.MarshalJSONToBytes(createReply)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		case http.MethodGet:
			b, err := ccloudv1.MarshalJSONToBytes(&ccloudv1.GetSchemaRegistryClustersReply{Clusters: []*ccloudv1.SchemaRegistryCluster{srCluster}})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/schema_registries/{id}"
func (c *CloudRouter) HandleSchemaRegistry(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		accountId := q.Get("account_id")
		srCluster := &ccloudv1.SchemaRegistryCluster{
			Id:        id,
			AccountId: accountId,
			Name:      "account schema-registry",
			Endpoint:  "SASL_SSL://sr-endpoint",
		}
		b, err := ccloudv1.MarshalJSONToBytes(&ccloudv1.GetSchemaRegistryClusterReply{Cluster: srCluster})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
