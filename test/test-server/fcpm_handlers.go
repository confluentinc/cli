package testserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"
)

func handleFcpmComputePools(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			euWest1 := flinkv2.FcpmV2ComputePool{
				Id: flinkv2.PtrString("lfcp-123456"),
				Spec: &flinkv2.FcpmV2ComputePoolSpec{
					DisplayName: flinkv2.PtrString("my-compute-pool-1"),
					MaxCfu:      flinkv2.PtrInt32(1),
					Region:      flinkv2.PtrString("eu-west-1"),
					Cloud:       flinkv2.PtrString("AWS"),
					Environment: &flinkv2.GlobalObjectReference{
						Id: "env-123",
					},
				},
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONED"},
			}
			euWest2 := flinkv2.FcpmV2ComputePool{
				Id: flinkv2.PtrString("lfcp-222222"),
				Spec: &flinkv2.FcpmV2ComputePoolSpec{
					DisplayName: flinkv2.PtrString("my-compute-pool-2"),
					MaxCfu:      flinkv2.PtrInt32(2),
					Region:      flinkv2.PtrString("eu-west-2"),
					Cloud:       flinkv2.PtrString("AWS"),
					Environment: &flinkv2.GlobalObjectReference{
						Id: "env-456",
					},
				},
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONED"},
			}

			computePools := []flinkv2.FcpmV2ComputePool{euWest1, euWest2}
			if r.URL.Query().Get("spec.region") == "eu-west-2" {
				computePools = []flinkv2.FcpmV2ComputePool{euWest2}
			}
			v := flinkv2.FcpmV2ComputePoolList{Data: computePools}
			setPageToken(&v, &v.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(v)
			require.NoError(t, err)
		case http.MethodPost:
			create := new(flinkv2.FcpmV2ComputePool)
			err := json.NewDecoder(r.Body).Decode(create)
			require.NoError(t, err)
			create.Spec.Cloud = flinkv2.PtrString(strings.ToUpper(create.Spec.GetCloud()))

			v := flinkv2.FcpmV2ComputePool{
				Id:     flinkv2.PtrString("lfcp-123456"),
				Spec:   create.Spec,
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONING"},
			}
			err = json.NewEncoder(w).Encode(v)
			require.NoError(t, err)
		}
	}
}

func handleFcpmComputePoolsId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var computePool flinkv2.FcpmV2ComputePool
		id := mux.Vars(r)["id"]
		if id != "lfcp-123456" && id != "lfcp-222222" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodGet:
			computePool = flinkv2.FcpmV2ComputePool{
				Id: flinkv2.PtrString(id),
				Spec: &flinkv2.FcpmV2ComputePoolSpec{
					DisplayName: flinkv2.PtrString("my-compute-pool-1"),
					MaxCfu:      flinkv2.PtrInt32(1),
					Cloud:       flinkv2.PtrString("AWS"),
					Region:      flinkv2.PtrString("eu-west-1"),
					Environment: &flinkv2.GlobalObjectReference{Id: "env-123"},
				},
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONED"},
			}
			if id == "lfcp-222222" {
				computePool.Spec.DisplayName = flinkv2.PtrString("my-compute-pool-2")
				computePool.Spec.Region = flinkv2.PtrString("eu-west-2")
				computePool.Spec.Environment = &flinkv2.GlobalObjectReference{Id: "env-456"}
			}
		case http.MethodPatch:
			update := new(flinkv2.FcpmV2ComputePool)
			err := json.NewDecoder(r.Body).Decode(update)
			require.NoError(t, err)

			computePool = flinkv2.FcpmV2ComputePool{
				Id: flinkv2.PtrString(id),
				Spec: &flinkv2.FcpmV2ComputePoolSpec{
					DisplayName: flinkv2.PtrString("my-compute-pool-1"),
					MaxCfu:      flinkv2.PtrInt32(update.Spec.GetMaxCfu()),
					Cloud:       flinkv2.PtrString("AWS"),
					Region:      flinkv2.PtrString("eu-west-1"),
					Environment: &flinkv2.GlobalObjectReference{Id: "env-123"},
				},
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONED"},
			}
		}

		err := json.NewEncoder(w).Encode(computePool)
		require.NoError(t, err)
	}
}

func handleFcpmRegions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		awsEuWest1 := flinkv2.FcpmV2Region{
			Id:                  flinkv2.PtrString("aws.eu-west-1"),
			DisplayName:         flinkv2.PtrString("Europe (eu-west-1)"),
			Cloud:               flinkv2.PtrString("AWS"),
			RegionName:          flinkv2.PtrString("eu-west-1"),
			HttpEndpoint:        flinkv2.PtrString(TestFlinkGatewayUrl.String()),
			PrivateHttpEndpoint: flinkv2.PtrString(TestFlinkGatewayUrlPrivate.String()),
		}
		awsEuWest2 := flinkv2.FcpmV2Region{
			Id:                  flinkv2.PtrString("aws.eu-west-2"),
			DisplayName:         flinkv2.PtrString("Europe (eu-west-2)"),
			Cloud:               flinkv2.PtrString("AWS"),
			RegionName:          flinkv2.PtrString("eu-west-2"),
			HttpEndpoint:        flinkv2.PtrString(TestFlinkGatewayUrl.String()),
			PrivateHttpEndpoint: flinkv2.PtrString(TestFlinkGatewayUrlPrivate.String()),
		}
		gcp := flinkv2.FcpmV2Region{
			Id:                  flinkv2.PtrString("gcp.europe-west3-a"),
			DisplayName:         flinkv2.PtrString("Frankfurt (europe-west3-a)"),
			Cloud:               flinkv2.PtrString("GCP"),
			RegionName:          flinkv2.PtrString("europe-west3-a"),
			HttpEndpoint:        flinkv2.PtrString(TestFlinkGatewayUrl.String()),
			PrivateHttpEndpoint: flinkv2.PtrString(TestFlinkGatewayUrlPrivate.String()),
		}
		azure := flinkv2.FcpmV2Region{
			Id:                  flinkv2.PtrString("azure.centralus"),
			DisplayName:         flinkv2.PtrString("Iowa (centralus)"),
			Cloud:               flinkv2.PtrString("AZURE"),
			RegionName:          flinkv2.PtrString("centralus"),
			HttpEndpoint:        flinkv2.PtrString(TestFlinkGatewayUrl.String()),
			PrivateHttpEndpoint: flinkv2.PtrString(TestFlinkGatewayUrlPrivate.String()),
		}

		// Allowing flexible mock results based on query parameters
		regions := []flinkv2.FcpmV2Region{awsEuWest1, awsEuWest2, gcp, azure}
		if r.URL.Query().Get("cloud") == "AWS" {
			regions = []flinkv2.FcpmV2Region{awsEuWest1, awsEuWest2}
			if r.URL.Query().Get("region_name") == "eu-west-1" {
				regions = []flinkv2.FcpmV2Region{awsEuWest1}
			} else if r.URL.Query().Get("region_name") == "eu-west-2" {
				regions = []flinkv2.FcpmV2Region{awsEuWest2}
			}
		} else if r.URL.Query().Get("cloud") == "GCP" {
			regions = []flinkv2.FcpmV2Region{gcp}
		} else if r.URL.Query().Get("cloud") == "AZURE" {
			regions = []flinkv2.FcpmV2Region{azure}
			if r.URL.Query().Get("region_name") == "eastus" {
				regions = []flinkv2.FcpmV2Region{}
			}
		}
		regionsList := &flinkv2.FcpmV2RegionList{Data: regions}
		setPageToken(regionsList, &regionsList.Metadata, r.URL)
		err := json.NewEncoder(w).Encode(regionsList)
		require.NoError(t, err)
	}
}
