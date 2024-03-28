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
		var v any

		switch r.Method {
		case http.MethodGet:
			usWest1 := flinkv2.FcpmV2ComputePool{
				Id: flinkv2.PtrString("lfcp-123456"),
				Spec: &flinkv2.FcpmV2ComputePoolSpec{
					DisplayName: flinkv2.PtrString("my-compute-pool-1"),
					MaxCfu:      flinkv2.PtrInt32(1),
					Region:      flinkv2.PtrString("us-west-1"),
					Cloud:       flinkv2.PtrString("AWS"),
					Environment: &flinkv2.GlobalObjectReference{
						Id: "env-123",
					},
				},
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONED"},
			}
			usWest2 := flinkv2.FcpmV2ComputePool{
				Id: flinkv2.PtrString("lfcp-222222"),
				Spec: &flinkv2.FcpmV2ComputePoolSpec{
					DisplayName: flinkv2.PtrString("my-compute-pool-2"),
					MaxCfu:      flinkv2.PtrInt32(2),
					Region:      flinkv2.PtrString("us-west-2"),
					Cloud:       flinkv2.PtrString("AWS"),
					Environment: &flinkv2.GlobalObjectReference{
						Id: "env-456",
					},
				},
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONED"},
			}

			computePools := []flinkv2.FcpmV2ComputePool{usWest1, usWest2}
			if r.URL.Query().Get("spec.region") == "us-west-2" {
				computePools = []flinkv2.FcpmV2ComputePool{usWest2}
			}
			v = flinkv2.FcpmV2ComputePoolList{Data: computePools}
		case http.MethodPost:
			create := new(flinkv2.FcpmV2ComputePool)
			err := json.NewDecoder(r.Body).Decode(create)
			require.NoError(t, err)
			create.Spec.Cloud = flinkv2.PtrString(strings.ToUpper(create.Spec.GetCloud()))

			v = flinkv2.FcpmV2ComputePool{
				Id:     flinkv2.PtrString("lfcp-123456"),
				Spec:   create.Spec,
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONING"},
			}
		}

		err := json.NewEncoder(w).Encode(v)
		require.NoError(t, err)
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
					Region:      flinkv2.PtrString("us-west-2"),
					Environment: &flinkv2.GlobalObjectReference{Id: "env-123"},
				},
				Status: &flinkv2.FcpmV2ComputePoolStatus{Phase: "PROVISIONED"},
			}
			if id == "lfcp-222222" {
				computePool.Spec.DisplayName = flinkv2.PtrString("my-compute-pool-2")
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
					Region:      flinkv2.PtrString("us-west-2"),
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
		aws := flinkv2.FcpmV2Region{
			Id:           flinkv2.PtrString("aws.eu-west-1"),
			DisplayName:  flinkv2.PtrString("Europe (eu-west-1)"),
			Cloud:        flinkv2.PtrString("AWS"),
			RegionName:   flinkv2.PtrString("eu-west-1"),
			HttpEndpoint: flinkv2.PtrString(TestFlinkGatewayUrl.String()),
		}
		gcp := flinkv2.FcpmV2Region{
			Id:           flinkv2.PtrString("gcp.europe-west3-a"),
			DisplayName:  flinkv2.PtrString("Frankfurt (europe-west3-a)"),
			Cloud:        flinkv2.PtrString("GCP"),
			RegionName:   flinkv2.PtrString("europe-west3-a"),
			HttpEndpoint: flinkv2.PtrString(TestFlinkGatewayUrl.String()),
		}

		regions := []flinkv2.FcpmV2Region{aws, gcp}
		if r.URL.Query().Get("cloud") == "AWS" {
			regions = []flinkv2.FcpmV2Region{aws}
		}

		err := json.NewEncoder(w).Encode(flinkv2.FcpmV2RegionList{Data: regions})
		require.NoError(t, err)
	}
}
