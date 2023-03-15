package price

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	climock "github.com/confluentinc/cli/mock"
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
		"-----------------+--------------+--------------+--------------+---------------",
		"  Connect record | Basic        | Single zone  | Internet     | $1.00 USD/GB  ",
	}, "\n")

	got, err := cmd.ExecuteCommand(command, "list", "--cloud", exampleCloud, "--region", exampleRegion)
	require.NoError(t, err)
	require.Equal(t, want+"\n", got)
}

func TestListLegacyClusterTypes(t *testing.T) {
	command := mockPriceCommand(map[string]float64{
		strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleLegacyClusterType, exampleNetworkType}, ":"): examplePrice,
	}, exampleMetric, exampleUnit)

	want := strings.Join([]string{
		"      Metric     |   Cluster Type    | Availability | Network Type |    Price      ",
		"-----------------+-------------------+--------------+--------------+---------------",
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

func mockPriceCommand(prices map[string]float64, metricsName, metricsUnit string) *cobra.Command {
	client := &ccloudv1.Client{
		Billing: &ccloudv1mock.Billing{
			GetPriceTableFunc: func(_ context.Context, organization *ccloudv1.Organization, _ string) (*ccloudv1.PriceTable, error) {
				table := &ccloudv1.PriceTable{
					PriceTable: map[string]*ccloudv1.UnitPrices{
						metricsName: {Unit: metricsUnit, Prices: prices},
					},
				}
				return table, nil
			},
		},
	}

	cfg := v1.AuthenticatedCloudConfigMock()

	return New(climock.NewPreRunnerMock(client, nil, nil, nil, cfg))
}

func TestFormatPrice(t *testing.T) {
	require.Equal(t, "$1.00 USD/GB", formatPrice(1, "GB"))
	require.Equal(t, "$1.20 USD/GB", formatPrice(1.2, "GB"))
	require.Equal(t, "$1.23 USD/GB", formatPrice(1.23, "GB"))
	require.Equal(t, "$1.234 USD/GB", formatPrice(1.234, "GB"))
}

func TestPrice_ClusterLink(t *testing.T) {
	command := mockPriceCommand(map[string]float64{
		strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
	}, "ClusterLinkingBase", "Hour")

	want := strings.Join([]string{
		"         Metric        | Cluster Type | Availability | Network Type |     Price       ",
		"-----------------------+--------------+--------------+--------------+-----------------",
		"  Cluster linking base | Basic        | Single zone  | Internet     | $1.00 USD/Hour  ",
	}, "\n")

	got, err := cmd.ExecuteCommand(command, "list", "--cloud", exampleCloud, "--region", exampleRegion)
	require.NoError(t, err)
	require.Equal(t, want+"\n", got)
}
