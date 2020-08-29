package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
)

const (
	exampleAvailability = "low"
	exampleCloud        = "aws"
	exampleClusterType  = "basic"
	exampleMetric       = "ConnectNumRecords"
	exampleNetworkType  = "internet"
	examplePrice        = 1
	exampleRegion       = "us-east-1"
	exampleUnit         = "GB"
)

func (s *CLITestSuite) TestPriceList() {
	tests := []CLITest{
		{
			args:    fmt.Sprintf("price list --cloud %s --region %s", exampleCloud, exampleRegion),
			fixture: "price/list.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		loginURL := serve(s.T(), "").URL
		s.runCcloudTest(test, loginURL)
	}
}

func handlePriceTable(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		prices := map[string]float64{
			strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
		}

		res := &orgv1.GetPriceTableReply{
			PriceTable: &orgv1.PriceTable{
				PriceTable: map[string]*orgv1.UnitPrices{
					exampleMetric: {Unit: exampleUnit, Prices: prices},
				},
			},
		}

		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}
