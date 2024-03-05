package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	srcmv3 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v3"
)

const (
	cloudSpec   = "aws"
	regionSpec  = "us-west-2"
	packageType = "essentials"

	srClusterId     = "lsrc-1234"
	srClusterStatus = "PROVISIONED"
)

func handleSchemaRegistryClustersV3(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			sgClusterList := getSchemaRegistryClusterListV3(TestSchemaRegistryUrl.String())
			err := json.NewEncoder(w).Encode(sgClusterList)
			require.NoError(t, err)
		}
	}
}

func handleSchemaRegistryClusterV3(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != srClusterId {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodGet:
			sgCluster := getSchemaRegistryClusterV3(packageType, TestSchemaRegistryUrl.String())
			err := json.NewEncoder(w).Encode(sgCluster)
			require.NoError(t, err)
		}
	}
}

func getSchemaRegistryClusterV3(packageType, endpoint string) srcmv3.SrcmV3Cluster {
	return srcmv3.SrcmV3Cluster{
		Id: srcmv3.PtrString(srClusterId),
		Spec: &srcmv3.SrcmV3ClusterSpec{
			DisplayName:  srcmv3.PtrString("account schema-registry"),
			Package:      &packageType,
			HttpEndpoint: &endpoint,
			Environment:  &srcmv3.GlobalObjectReference{Id: SRApiEnvId},
			Region:       srcmv3.PtrString(regionSpec),
			Cloud:        srcmv3.PtrString(cloudSpec),
		},
		Status: &srcmv3.SrcmV3ClusterStatus{Phase: srClusterStatus},
	}
}

func getSchemaRegistryClusterListV3(httpEndpoint string) srcmv3.SrcmV3ClusterList {
	srcmClusterList := srcmv3.SrcmV3ClusterList{
		Data: []srcmv3.SrcmV3Cluster{
			getSchemaRegistryClusterV3(packageType, httpEndpoint)},
	}

	return srcmClusterList
}
