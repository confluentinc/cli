package price

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccloudmock "github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

const (
	exampleAvailability      = "low"
	exampleCloud             = "aws"
	exampleClusterType       = "basic"
	exampleLegacyClusterType = "standard"
	exampleMetric            = "ConnectNumRecords"
	exampleNetworkType       = "internet"
	examplePrice             = 1
	exampleRegion            = "us-east-1"
	exampleUnit              = "GB"
)

func TestRequireFlags(t *testing.T) {
	var err error

	_, err = cmd.ExecuteCommand(mockPriceCommand(nil), "list", "--cloud", exampleCloud)
	require.Error(t, err)

	_, err = cmd.ExecuteCommand(mockPriceCommand(nil), "list", "--region", exampleRegion)
	require.Error(t, err)
}

func TestList(t *testing.T) {
	command := mockSingleRowCommand()

	want := strings.Join([]string{
		"      Metric     | Cluster Type | Availability | Network Type |    Price      ",
		"+----------------+--------------+--------------+--------------+--------------+",
		"  Connect record | Basic        | Single zone  | Internet     | $1.00 USD/GB  ",
	}, "\n")

	got, err := cmd.ExecuteCommand(command, "list", "--cloud", exampleCloud, "--region", exampleRegion)
	require.NoError(t, err)
	require.Equal(t, want+"\n", got)
}

func TestListJSON(t *testing.T) {
	command := mockSingleRowCommand()

	res := []map[string]string{
		{
			"availability": exampleAvailability,
			"cluster_type": exampleClusterType,
			"metric":       exampleMetric,
			"network_type": exampleNetworkType,
			"price":        strconv.Itoa(examplePrice),
		},
	}
	want, err := json.MarshalIndent(res, "", "  ")
	require.NoError(t, err)

	got, err := cmd.ExecuteCommand(command, "list", "--cloud", exampleCloud, "--region", exampleRegion, "-o", "json")
	require.NoError(t, err)
	require.Equal(t, string(want)+"\n", got)
}

func TestListLegacyClusterTypes(t *testing.T) {
	command := mockPriceCommand(map[string]float64{
		strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleLegacyClusterType, exampleNetworkType}, ":"): examplePrice,
	})

	want := strings.Join([]string{
		"      Metric     |   Cluster Type    | Availability | Network Type |    Price      ",
		"+----------------+-------------------+--------------+--------------+--------------+",
		"  Connect record | Legacy - Standard | Single zone  | Internet     | $1.00 USD/GB  ",
	}, "\n")

	got, err := cmd.ExecuteCommand(command, "list", "--cloud", exampleCloud, "--region", exampleRegion, "--legacy")
	require.NoError(t, err)
	require.Equal(t, want+"\n", got)
}

func TestOmitLegacyClusterTypes(t *testing.T) {
	command := mockPriceCommand(map[string]float64{
		strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleLegacyClusterType, exampleNetworkType}, ":"): examplePrice,
	})

	_, err := cmd.ExecuteCommand(command, "list", "--cloud", exampleCloud, "--region", exampleRegion)
	require.Error(t, err)
}

func mockSingleRowCommand() *cobra.Command {
	return mockPriceCommand(map[string]float64{
		strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
	})
}

func mockPriceCommand(prices map[string]float64) *cobra.Command {
	client := &ccloud.Client{
		Organization: &ccloudmock.Organization{
			GetPriceTableFunc: func(_ context.Context, organization *orgv1.Organization) (*orgv1.PriceTable, error) {
				table := &orgv1.PriceTable{
					PriceTable: map[string]*orgv1.UnitPrices{
						exampleMetric: {Unit: exampleUnit, Prices: prices},
					},
				}
				return table, nil
			},
		},
	}

	cfg := v3.AuthenticatedCloudConfigMock()

	return New(mock.NewPreRunnerMock(client, nil, cfg))
}

func TestFormatPrice(t *testing.T) {
	require.Equal(t, "$0.12 USD/GB", formatPrice(0.123, "GB"))
	require.Equal(t, "$0.001 USD/GB", formatPrice(0.0012, "GB"))
}
