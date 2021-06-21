package price

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	billingv1 "github.com/confluentinc/cc-structs/kafka/billing/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccloudmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
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

	_, err = cmd.ExecuteCommand(mockPriceCommand(nil, exampleMetric, exampleUnit), "list", "--cloud", exampleCloud)
	require.Error(t, err)

	_, err = cmd.ExecuteCommand(mockPriceCommand(nil, exampleMetric, exampleUnit), "list", "--region", exampleRegion)
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
	}, exampleMetric, exampleUnit)

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
	}, exampleMetric, exampleUnit)

	_, err := cmd.ExecuteCommand(command, "list", "--cloud", exampleCloud, "--region", exampleRegion)
	require.Error(t, err)
}

func mockSingleRowCommand() *cobra.Command {
	return mockPriceCommand(map[string]float64{
		strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
	}, exampleMetric, exampleUnit)
}

func mockPriceCommand(prices map[string]float64, metricsName string, metricsUnit string) *cobra.Command {
	client := &ccloud.Client{
		Billing: &ccloudmock.Billing{
			GetPriceTableFunc: func(_ context.Context, organization *orgv1.Organization, _ string) (*billingv1.PriceTable, error) {
				table := &billingv1.PriceTable{
					PriceTable: map[string]*billingv1.UnitPrices{
						metricsName: {Unit: metricsUnit, Prices: prices},
					},
				}
				return table, nil
			},
		},
	}

	cfg := v3.AuthenticatedCloudConfigMock()

	return New(mock.NewPreRunnerMock(client, nil, nil, cfg))
}

func TestFormatPrice(t *testing.T) {
	require.Equal(t, "$1.00 USD/GB", formatPrice(1, "GB"))
	require.Equal(t, "$1.20 USD/GB", formatPrice(1.2, "GB"))
	require.Equal(t, "$1.23 USD/GB", formatPrice(1.23, "GB"))
	require.Equal(t, "$1.234 USD/GB", formatPrice(1.234, "GB"))
}

func TestClusterLinkPrice(t *testing.T) {
	command := mockPriceCommand(map[string]float64{
		strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
	}, "ClusterLinkingBase", "Hour")

	want := strings.Join([]string{
		"         Metric        | Cluster Type | Availability | Network Type |     Price       ",
		"+----------------------+--------------+--------------+--------------+----------------+",
		"  Cluster linking base | Basic        | Single zone  | Internet     | $1.00 USD/Hour  ",
	}, "\n")

	got, err := cmd.ExecuteCommand(command, "list", "--cloud", exampleCloud, "--region", exampleRegion)
	require.NoError(t, err)
	require.Equal(t, want+"\n", got)
}
