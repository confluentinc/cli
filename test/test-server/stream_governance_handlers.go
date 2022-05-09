package test_server

import (
	"encoding/json"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

const (
	id                = "lsrc-1234"
	httpEndpoint      = "https://sr-endpoint"
	status            = "PROVISIONED"
	packageType       = "advanced"
	regionId          = "sgreg=1"
	cloud             = "aws"
	regionName        = "us-east-2"
	regionDisplayName = "Ohio (us-east-2)"
)

func handleStreamGovernanceClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			req := new(sgsdk.StreamGovernanceV2Cluster)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)

			sgCluster := getStreamGovernanceCluster(id, *req.Spec.Package, httpEndpoint, SRApiEnvId, regionId, status)
			err = json.NewEncoder(w).Encode(sgCluster)
			require.NoError(t, err)
		} else if r.Method == http.MethodGet {
			sgClusterList := getStreamGovernanceClusterList()
			err := json.NewEncoder(w).Encode(sgClusterList)
			require.NoError(t, err)
		}
	}
}

func handleStreamGovernanceCluster(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]

		sgCluster := getStreamGovernanceCluster(id, packageType, httpEndpoint, SRApiEnvId, regionId, status)
		err := json.NewEncoder(w).Encode(sgCluster)
		require.NoError(t, err)
	}
}

func handleStreamGovernanceRegions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := r.URL.Query()
		cloud := q.Get("spec.cloud")
		region := q.Get("spec.region_name")

		sgRegionList := getStreamGovernanceRegionList(cloud, region)
		err := json.NewEncoder(w).Encode(sgRegionList)
		require.NoError(t, err)
	}
}

func handleStreamGovernanceRegion(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]

		switch id {
		case regionId:
			sgRegion := getStreamGovernanceRegion(
				id, regionName, cloud, packageType, regionDisplayName)
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

func getStreamGovernanceClusterList() sgsdk.StreamGovernanceV2ClusterList {
	sgClusterList := sgsdk.StreamGovernanceV2ClusterList{
		Data: []sgsdk.StreamGovernanceV2Cluster{
			getStreamGovernanceCluster(id, packageType, httpEndpoint, SRApiEnvId, regionId, status)},
	}

	return sgClusterList
}

func getStreamGovernanceRegionList(filterCloud, filterRegion string) sgsdk.StreamGovernanceV2RegionList {
	sgRegionList := sgsdk.StreamGovernanceV2RegionList{
		Data: []sgsdk.StreamGovernanceV2Region{
			getStreamGovernanceRegion(regionId, regionName, cloud, packageType, regionDisplayName)},
	}
	sgRegionList.Data = filterRegionList(sgRegionList.Data, filterCloud, filterRegion)

	return sgRegionList
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

func filterRegionList(regionList []sgsdk.StreamGovernanceV2Region, cloud, region string) (filteredRegionList []sgsdk.StreamGovernanceV2Region) {
	for _, regionSpec := range regionList {
		if (regionSpec.Spec.GetCloud() == cloud || cloud == "") && (regionSpec.Spec.GetRegionName() == region || region == "") {
			filteredRegionList = append(filteredRegionList, regionSpec)
		}
	}
	return
}
