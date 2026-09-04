package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
)

var queryTime = time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC)

// schemaCountTwoPsrcValue simulates an LSRC whose schemas span two PSRCs (e.g.
// PSRC1=6, PSRC2=1 schemas). With the gauge's default MEAN time aggregation,
// the API would return a fractional value (~3.5). With the billing-shaped query
// (time_agg=MAX, agg=SUM), the per-PSRC counts collapse correctly.
const schemaCountTwoPsrcValue float64 = 7

func handleMetricsQuery(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		// schema_count is a GAUGE; for legacy LSRCs spanning two PSRCs, the
		// API's default MEAN under-counts. The CLI must send the undocumented
		// "time_agg":"MAX" alongside "agg":"SUM". See
		// internal/schema-registry/command_cluster_describe.go:schemaCountQueryBodyFor.
		// We parse as raw JSON because the v2 SDK type can't represent time_agg.
		var raw map[string]any
		require.NoError(t, json.Unmarshal(body, &raw))

		value := 0.0
		if aggs, ok := raw["aggregations"].([]any); ok {
			for _, a := range aggs {
				agg, _ := a.(map[string]any)
				metric, _ := agg["metric"].(string)
				if !strings.HasSuffix(metric, "/schema_count") {
					continue
				}
				require.Equal(t, "MAX", agg["time_agg"], "schema_count query must set time_agg=MAX to override the gauge MEAN default")
				require.Equal(t, "SUM", agg["agg"], "schema_count query must set agg=SUM")
				value = schemaCountTwoPsrcValue
			}
		}

		resp := &metricsv2.QueryResponse{
			FlatQueryResponse: &metricsv2.FlatQueryResponse{
				Data: []metricsv2.Point{
					{Value: float32(value), Timestamp: queryTime},
				},
			},
		}
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}
}
