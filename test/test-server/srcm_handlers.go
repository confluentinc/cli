package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	srcm "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
)

const (
	// test region 1
	idRegion1          = "sgreg-1"
	packageTypeRegion1 = "advanced"
	cloudRegion1       = "aws"
	nameRegion1        = "us-east-2"
	displayNameRegion1 = "Ohio (us-east-2)"

	// test region 2
	idRegion2          = "sgreg-2"
	packageTypeRegion2 = "essentials"
	cloudRegion2       = "gcp"
	nameRegion2        = "us-central1"
	displayNameRegion2 = "Iowa (us-central1)"
)

func handleSchemaRegistryRegions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]

		switch id {
		case idRegion1:
			sgRegion := getSchemaRegistryRegion(
				id, nameRegion1, cloudRegion1, packageTypeRegion1, displayNameRegion1)
			err := json.NewEncoder(w).Encode(sgRegion)
			require.NoError(t, err)
		case idRegion2:
			sgRegion := getSchemaRegistryRegion(
				id, nameRegion2, cloudRegion2, packageTypeRegion2, displayNameRegion2)
			err := json.NewEncoder(w).Encode(sgRegion)
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func getSchemaRegistryRegionList(filterCloud, filterRegion, filterPackage string) srcm.SrcmV2RegionList {
	sgRegionList := srcm.SrcmV2RegionList{
		Data: []srcm.SrcmV2Region{
			getSchemaRegistryRegion(idRegion1, nameRegion1, cloudRegion1, packageTypeRegion1, displayNameRegion1),
			getSchemaRegistryRegion(idRegion2, nameRegion2, cloudRegion2, packageTypeRegion2, displayNameRegion2),
		},
	}
	sgRegionList.Data = filterRegionList(sgRegionList.Data, filterCloud, filterRegion, filterPackage)

	return sgRegionList
}

func getSchemaRegistryRegion(id, region, cloud, packageType, displayName string) srcm.SrcmV2Region {
	return srcm.SrcmV2Region{
		Id: &id,
		Spec: &srcm.SrcmV2RegionSpec{
			RegionName:  &region,
			Cloud:       &cloud,
			Packages:    &[]string{packageType},
			DisplayName: &displayName,
		},
	}
}

func filterRegionList(regionList []srcm.SrcmV2Region, cloud, region, packageType string) []srcm.SrcmV2Region {
	var filteredRegionList []srcm.SrcmV2Region
	for _, regionSpec := range regionList {
		if (regionSpec.Spec.GetCloud() == cloud || cloud == "") &&
			(regionSpec.Spec.GetRegionName() == region || region == "") &&
			(contains(regionSpec.Spec.GetPackages(), packageType) || packageType == "") {
			filteredRegionList = append(filteredRegionList, regionSpec)
		}
	}
	return filteredRegionList
}

func contains(arr []string, elementToCheck string) bool {
	for _, ele := range arr {
		if ele == elementToCheck {
			return true
		}
	}
	return false
}
