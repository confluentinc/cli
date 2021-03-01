package test_server

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"
	"github.com/stretchr/testify/require"
)

const (
	SRApiEnvId = "env-srUpdate"
)

// Handler for: "/api/schema_registries"
func (c *CloudRouter) HandleSchemaRegistries(t *testing.T) func(http.ResponseWriter, *http.Request) {
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
		srCluster := &schedv1.SchemaRegistryCluster{
			Id:        id,
			AccountId: accountId,
			Name:      "account schema-registry",
			Endpoint:  endpoint,
		}
		switch r.Method {
		case "POST":
			createReply := &schedv1.CreateSchemaRegistryClusterReply{Cluster: srCluster}
			b, err := utilv1.MarshalJSONToBytes(createReply)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		case "GET":
			b, err := utilv1.MarshalJSONToBytes(&schedv1.GetSchemaRegistryClustersReply{Clusters: []*schedv1.SchemaRegistryCluster{srCluster}})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/schema_registries/{id}"
func (c *CloudRouter) HandleSchemaRegistry(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("sr handler")
		q := r.URL.Query()
		id := q.Get("id")
		accountId := q.Get("account_id")
		srCluster := &schedv1.SchemaRegistryCluster{
			Id:        id,
			AccountId: accountId,
			Name:      "account schema-registry",
			Endpoint:  "SASL_SSL://sr-endpoint",
		}
		fmt.Println(srCluster)
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetSchemaRegistryClusterReply{Cluster: srCluster})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
