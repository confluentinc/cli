package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"
)

// Handler for: "/fcpm/v2/regions"
func handleFcpmRegions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		aws := flinkv2.FcpmV2Region{
			DisplayName: flinkv2.PtrString("Europe (eu-west-1)"),
			Cloud:       flinkv2.PtrString("AWS"),
			RegionName:  flinkv2.PtrString("eu-west-1"),
		}
		gcp := flinkv2.FcpmV2Region{
			DisplayName: flinkv2.PtrString("Frankfurt (europe-west3-a)"),
			Cloud:       flinkv2.PtrString("GCP"),
			RegionName:  flinkv2.PtrString("europe-west3-a"),
		}

		regions := []flinkv2.FcpmV2Region{aws, gcp}
		if r.URL.Query().Get("cloud") == "AWS" {
			regions = []flinkv2.FcpmV2Region{aws}
		}

		err := json.NewEncoder(w).Encode(flinkv2.FcpmV2RegionList{Data: regions})
		require.NoError(t, err)
	}
}
