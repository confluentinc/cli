package testserver

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	rtcev1 "github.com/confluentinc/ccloud-sdk-go-v2/rtce/v1"
)

// Handler for "/rtce/v1/regions"
func handleRtceV1Regions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			region := readRtceV1RegionFile(t, "read_created_region.json")

			regionList := &rtcev1.RtceV1RegionList{
				Data: []rtcev1.RtceV1Region{region},
			}

			err := json.NewEncoder(w).Encode(regionList)
			require.NoError(t, err)
		}
	}
}

func readRtceV1RegionFile(t *testing.T, filename string) rtcev1.RtceV1Region {
	jsonPath := filepath.Join("..", "fixtures", "input", "rtce", "region", filename)
	jsonData, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	region := rtcev1.RtceV1Region{}
	err = json.Unmarshal(jsonData, &region)
	require.NoError(t, err)

	return region
}
