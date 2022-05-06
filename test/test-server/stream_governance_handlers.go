package test_server

import (
	"encoding/json"
	"fmt"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func handleStreamGovernanceClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		req := new(sgsdk.StreamGovernanceV2Cluster)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		id := "lsrc-1234"
		regionId := "sgreg-1"
		httpEndpoint := "https://sr-endpoint"
		status := "PROVISIONED"
		sgCluster := getStreamGovernanceCluster(id, *req.Spec.Package, httpEndpoint, SRApiEnvId, regionId, status)

		err = json.NewEncoder(w).Encode(sgCluster)
		require.NoError(t, err)
	}
}

func handleStreamGovernanceCluster(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]

		regionId := "sgreg-1"
		packageType := "advanced"
		httpEndpoint := "https://sr-endpoint"
		status := "PROVISIONED"

		sgCluster := getStreamGovernanceCluster(id, packageType, httpEndpoint, SRApiEnvId, regionId, status)
		err := json.NewEncoder(w).Encode(sgCluster)
		require.NoError(t, err)
	}
}

func handleStreamGovernanceRegions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Println("Inside region method")

		sgRegionList := &sgsdk.StreamGovernanceV2RegionList{
			Data: []sgsdk.StreamGovernanceV2Region{
				getStreamGovernanceRegion("sgreg-1", "us-east-2", "aws", "advanced", "Ohio (us-east-2)")},
		}
		err := json.NewEncoder(w).Encode(sgRegionList)
		require.NoError(t, err)
	}
}

func handleStreamGovernanceRegion(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]

		switch id {
		case "sgreg-1":
			sgRegion := getStreamGovernanceRegion(
				id, "us-east-2", "aws", "advanced", "Ohio (us-east-2)")
			err := json.NewEncoder(w).Encode(sgRegion)
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func getStreamGovernanceCluster(id, packageType, endpoint, envId, regionId, status string) sgsdk.StreamGovernanceV2Cluster {
	return sgsdk.StreamGovernanceV2Cluster{
		Id: &id,
		Spec: &sgsdk.StreamGovernanceV2ClusterSpec{
			DisplayName:  &id,
			Package:      &packageType,
			HttpEndpoint: &endpoint,
			Environment: &sgsdk.ObjectReference{
				Id: envId,
			},
			Region: &sgsdk.ObjectReference{
				Id: regionId,
			},
		},
		Status: &sgsdk.StreamGovernanceV2ClusterStatus{
			Phase: status,
		},
	}
}

func getStreamGovernanceRegion(id, region, cloud, packageType, displayName string) sgsdk.StreamGovernanceV2Region {
	return sgsdk.StreamGovernanceV2Region{
		Id: &id,
		Spec: &sgsdk.StreamGovernanceV2RegionSpec{
			RegionName:  &region,
			Cloud:       &cloud,
			Packages:    &[]string{packageType},
			DisplayName: &displayName,
		},
	}
}
