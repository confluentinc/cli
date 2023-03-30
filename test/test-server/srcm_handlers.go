package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	srcm "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

const (
	id           = "lsrc-1234"
	httpEndpoint = "https://sr-endpoint"
	status       = "PROVISIONED"
	packageType  = "essentials"
	regionId     = "sgreg=1"
)

func handleSchemaRegistryClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			req := new(srcm.SrcmV2Cluster)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)

			sgCluster := getSchemaRegistryCluster(id, *req.Spec.Package, httpEndpoint, SRApiEnvId, regionId, status)
			err = json.NewEncoder(w).Encode(sgCluster)
			require.NoError(t, err)
		} else if r.Method == http.MethodGet {
			sgClusterList := getSchemaRegistryClusterList()
			err := json.NewEncoder(w).Encode(sgClusterList)
			require.NoError(t, err)
		}
	}
}

func handleSchemaRegistryCluster(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]

		sgCluster := getSchemaRegistryCluster(id, packageType, httpEndpoint, SRApiEnvId, regionId, status)
		err := json.NewEncoder(w).Encode(sgCluster)
		require.NoError(t, err)
	}
}

func getSchemaRegistryCluster(id, packageType, endpoint, envId, regionId, status string) srcm.SrcmV2Cluster {
	return srcm.SrcmV2Cluster{
		Id: &id,
		Spec: &srcm.SrcmV2ClusterSpec{
			DisplayName:  &id,
			Package:      &packageType,
			HttpEndpoint: &endpoint,
			Environment: &srcm.GlobalObjectReference{
				Id: envId,
			},
			Region: &srcm.GlobalObjectReference{
				Id: regionId,
			},
		},
		Status: &srcm.SrcmV2ClusterStatus{
			Phase: status,
		},
	}
}

func getSchemaRegistryClusterList() srcm.SrcmV2ClusterList {
	srcmClusterList := srcm.SrcmV2ClusterList{
		Data: []srcm.SrcmV2Cluster{
			getSchemaRegistryCluster(id, packageType, httpEndpoint, SRApiEnvId, regionId, status)},
	}

	return srcmClusterList
}
