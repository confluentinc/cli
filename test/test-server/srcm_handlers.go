package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	"github.com/confluentinc/cli/internal/pkg/types"
)

const (
	// test region 1
	idRegion1          = "sgreg-1"
	packageTypeRegion1 = "advanced"
	cloudRegion1       = "aws"
	nameRegion1        = "us-west-2"
	displayNameRegion1 = "Ohio (us-east-2)"

	// test region 2
	idRegion2          = "sgreg-2"
	packageTypeRegion2 = "essentials"
	cloudRegion2       = "gcp"
	nameRegion2        = "us-central1"
	displayNameRegion2 = "Iowa (us-central1)"

	srClusterId = "lsrc-1234"
	status      = "PROVISIONED"
)

func (c *CloudV2Router) HandleSchemaRegistryClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		accountId := q.Get("environment")
		var endpoint string
		// for sr commands that use the sr api (use accountId to differentiate) we want to use the SR server URL so that we can make subsequent requests there
		// for describe commands we want to use a standard endpoint so that it will always match the test fixture
		if accountId == SRApiEnvId {
			endpoint = c.srApiUrl
		} else {
			endpoint = "SASL_SSL://sr-endpoint"
		}
		if r.Method == http.MethodPost {
			req := new(srcmv2.SrcmV2Cluster)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			sgCluster := getSchemaRegistryCluster(srClusterId, req.Spec.GetPackage(), endpoint, SRApiEnvId, idRegion1, status)
			err = json.NewEncoder(w).Encode(sgCluster)
			require.NoError(t, err)
		} else if r.Method == http.MethodGet {
			sgClusterList := getSchemaRegistryClusterList(endpoint)
			err := json.NewEncoder(w).Encode(sgClusterList)
			require.NoError(t, err)
		}
	}
}

func handleSchemaRegistryCluster(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			req := new(srcmv2.SrcmV2ClusterUpdate)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			packageType := req.Spec.GetPackage()
			sgCluster := getSchemaRegistryCluster(srClusterId, packageType, "https://sr-endpoint", SRApiEnvId, idRegion1, status)
			err = json.NewEncoder(w).Encode(sgCluster)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			sgCluster := getSchemaRegistryCluster(srClusterId, packageTypeRegion2, "https://sr-endpoint", SRApiEnvId, idRegion1, status)
			err := json.NewEncoder(w).Encode(sgCluster)
			require.NoError(t, err)
		}
	}
}

func getSchemaRegistryCluster(id, packageType, endpoint, envId, idRegion1, status string) srcmv2.SrcmV2Cluster {
	return srcmv2.SrcmV2Cluster{
		Id: &id,
		Spec: &srcmv2.SrcmV2ClusterSpec{
			DisplayName:  srcmv2.PtrString("account schema-registry"),
			Package:      &packageType,
			HttpEndpoint: &endpoint,
			Environment: &srcmv2.GlobalObjectReference{
				Id: envId,
			},
			Region: &srcmv2.GlobalObjectReference{
				Id: idRegion1,
			},
		},
		Status: &srcmv2.SrcmV2ClusterStatus{
			Phase: status,
		},
	}
}

func getSchemaRegistryClusterList(httpEndpoint string) srcmv2.SrcmV2ClusterList {
	srcmClusterList := srcmv2.SrcmV2ClusterList{
		Data: []srcmv2.SrcmV2Cluster{
			getSchemaRegistryCluster(srClusterId, packageTypeRegion2, httpEndpoint, SRApiEnvId, idRegion1, status)},
	}

	return srcmClusterList
}

func handleSchemaRegistryRegions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		cloud := q.Get("spec.cloud")
		region := q.Get("spec.region_name")
		packageType := q.Get("spec.packages")

		sgRegionList := getSchemaRegistryRegionList(cloud, region, packageType)
		err := json.NewEncoder(w).Encode(sgRegionList)
		require.NoError(t, err)
	}
}

func handleSchemaRegistryRegion(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		switch id {
		case idRegion1:
			sgRegion := getSchemaRegistryRegion(id, nameRegion1, cloudRegion1, packageTypeRegion1, displayNameRegion1)
			err := json.NewEncoder(w).Encode(sgRegion)
			require.NoError(t, err)
		case idRegion2:
			sgRegion := getSchemaRegistryRegion(id, nameRegion2, cloudRegion2, packageTypeRegion2, displayNameRegion2)
			err := json.NewEncoder(w).Encode(sgRegion)
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func getSchemaRegistryRegionList(filterCloud, filterRegion, filterPackage string) srcmv2.SrcmV2RegionList {
	sgRegionList := srcmv2.SrcmV2RegionList{
		Data: []srcmv2.SrcmV2Region{
			getSchemaRegistryRegion(idRegion1, nameRegion1, cloudRegion1, packageTypeRegion1, displayNameRegion1),
			getSchemaRegistryRegion(idRegion2, nameRegion2, cloudRegion2, packageTypeRegion2, displayNameRegion2),
		},
	}
	sgRegionList.Data = filterRegionList(sgRegionList.Data, filterCloud, filterRegion, filterPackage)

	return sgRegionList
}

func getSchemaRegistryRegion(id, region, cloud, packageType, displayName string) srcmv2.SrcmV2Region {
	return srcmv2.SrcmV2Region{
		Id: srcmv2.PtrString(id),
		Spec: &srcmv2.SrcmV2RegionSpec{
			RegionName:  srcmv2.PtrString(region),
			Cloud:       srcmv2.PtrString(cloud),
			Packages:    &[]string{packageType},
			DisplayName: srcmv2.PtrString(displayName),
		},
	}
}

func filterRegionList(regionList []srcmv2.SrcmV2Region, cloud, region, packageType string) []srcmv2.SrcmV2Region {
	var filteredRegionList []srcmv2.SrcmV2Region
	for _, regionSpec := range regionList {
		if (regionSpec.Spec.GetCloud() == cloud || cloud == "") &&
			(regionSpec.Spec.GetRegionName() == region || region == "") &&
			(types.Contains(regionSpec.Spec.GetPackages(), packageType) || packageType == "") {
			filteredRegionList = append(filteredRegionList, regionSpec)
		}
	}
	return filteredRegionList
}
