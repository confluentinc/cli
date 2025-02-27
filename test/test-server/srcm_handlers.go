package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	srcmv3 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v3"
	srcmv3Access "github.com/confluentinc/ccloud-sdk-go-v2/srcmv3access/v1"
)

const (
	cloudSpec   = "aws"
	regionSpec  = "us-west-2"
	packageType = "essentials"

	srClusterId     = "lsrc-1234"
	ksqlClusterId   = "lksqlc-ksql5"
	ksqlClusterId2  = "lksqlc-woooo"
	ksqlClusterId3  = "lksqlc-ksql5-fail"
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
			DisplayName:         srcmv3.PtrString("account schema-registry"),
			Package:             srcmv3.PtrString(packageType),
			HttpEndpoint:        srcmv3.PtrString(endpoint),
			PrivateHttpEndpoint: srcmv3.PtrString("http://127.0.0.1:1029"),
			PrivateNetworkingConfig: &srcmv3.SrcmV3ClusterSpecPrivateNetworkingConfig{
				RegionalEndpoints: &map[string]string{"us-east-1": "https://lsrc-stk1d.us-east-1.aws.private.stag.cpdev.cloud", "us-west-2": "https://lsrc-stgvk1d.us-west-2.aws.private.stag.cpdev.cloud"},
			},
			CatalogHttpEndpoint: srcmv3.PtrString("http://127.0.0.1:1030"),
			Environment:         &srcmv3.GlobalObjectReference{Id: SRApiEnvId},
			Region:              srcmv3.PtrString(regionSpec),
			Cloud:               srcmv3.PtrString(cloudSpec),
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

func handleSchemaRegistryClusterV3Access(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != ksqlClusterId && id != ksqlClusterId2 && id != ksqlClusterId3 {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodGet:
			sgCluster := getSchemaRegistryClusterV3Access(id)
			err := json.NewEncoder(w).Encode(sgCluster)
			require.NoError(t, err)
		}
	}
}

func getSchemaRegistryClusterV3Access(id string) srcmv3Access.SrcmV3Access {
	if id == ksqlClusterId3 {
		return srcmv3Access.SrcmV3Access{
			Id: srcmv3Access.PtrString(ksqlClusterId3),
			Spec: &srcmv3Access.SrcmV3AccessSpec{
				Allowed: srcmv3Access.PtrBool(false),
			},
		}
	} else {
		return srcmv3Access.SrcmV3Access{
			Id: srcmv3Access.PtrString(ksqlClusterId),
			Spec: &srcmv3Access.SrcmV3AccessSpec{
				Allowed: srcmv3Access.PtrBool(true),
			},
		}
	}
}
