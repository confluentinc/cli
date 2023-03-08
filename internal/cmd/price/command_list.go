package price

import (
	"context"
	"fmt"
	"strings"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/types"
)

var (
	formatMetric = map[string]string{
		"ConnectCapacity":       "Connect capacity",
		"ConnectNumRecords":     "Connect record",
		"ConnectNumTasks":       "Connect task",
		"ConnectThroughput":     "Connect throughput",
		"KSQLNumCSUs":           "ksqlDB capacity",
		"KafkaBase":             "Kafka base price",
		"KafkaCKUUnit":          "CKU unit price",
		"KafkaNetworkRead":      "Kafka read",
		"KafkaNetworkWrite":     "Kafka write",
		"KafkaNumCKUs":          "CKU price",
		"KafkaPartition":        "Kafka partition",
		"KafkaRestProduce":      "Kafka REST produce",
		"KafkaStorage":          "Kafka storage",
		"ClusterLinkingBase":    "Cluster linking base",
		"ClusterLinkingPerLink": "Cluster linking per-link",
		"ClusterLinkingRead":    "Cluster linking read",
		"ClusterLinkingWrite":   "Cluster linking write",
	}

	formatClusterTypeSerialized = map[string]string{
		"basic":       "basic",
		"custom":      "custom-legacy",
		"dedicated":   "dedicated",
		"standard":    "basic-legacy",
		"standard_v2": "standard",
	}

	formatClusterTypeHuman = map[string]string{
		"basic":       "Basic",
		"custom":      "Legacy - Custom",
		"dedicated":   "Dedicated",
		"standard":    "Legacy - Standard",
		"standard_v2": "Standard",
	}

	formatAvailability = map[string]string{
		"high": "Multi zone",
		"low":  "Single zone",
	}

	formatNetworkType = map[string]string{
		"internet":        "Internet",
		"peered-vpc":      "Peered VPC",
		"private-link":    "Private Link",
		"transit-gateway": "Transit Gateway",
	}
)

var (
	metrics        = types.GetSortedKeys(formatMetric)
	clusterTypes   = types.GetSortedValues(formatClusterTypeSerialized)
	availabilities = types.GetSortedKeys(formatAvailability)
	networkTypes   = types.GetSortedKeys(formatNetworkType)
)

type humanOut struct {
	Metric       string `human:"Metric"`
	ClusterType  string `human:"Cluster Type"`
	Availability string `human:"Availability"`
	NetworkType  string `human:"Network Type"`
	Price        string `human:"Price"`
}

type serializedOut struct {
	Metric       string  `serialized:"metric"`
	ClusterType  string  `serialized:"cluster_type"`
	Availability string  `serialized:"availability"`
	NetworkType  string  `serialized:"network_type"`
	Price        float64 `serialized:"price"`
}

type row struct {
	metric       string
	clusterType  string
	availability string
	networkType  string
	price        float64
	unit         string
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print an organization's price list.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("availability", "", fmt.Sprintf("Filter by availability (%s).", strings.Join(availabilities, ", ")))
	cmd.Flags().String("cluster-type", "", fmt.Sprintf("Filter by cluster type (%s).", strings.Join(clusterTypes, ", ")))
	cmd.Flags().String("network-type", "", fmt.Sprintf("Filter by network type (%s).", strings.Join(networkTypes, ", ")))
	cmd.Flags().String("metric", "", fmt.Sprintf("Filter by metric (%s).", strings.Join(metrics, ", ")))
	cmd.Flags().Bool("legacy", false, "Show legacy cluster types.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("cloud")
	_ = cmd.MarkFlagRequired("region")

	pcmd.RegisterFlagCompletionFunc(cmd, "availability", func(_ *cobra.Command, _ []string) []string { return availabilities })
	pcmd.RegisterFlagCompletionFunc(cmd, "cluster-type", func(_ *cobra.Command, _ []string) []string { return clusterTypes })
	pcmd.RegisterFlagCompletionFunc(cmd, "network-type", func(_ *cobra.Command, _ []string) []string { return networkTypes })
	pcmd.RegisterFlagCompletionFunc(cmd, "metric", func(_ *cobra.Command, _ []string) []string { return metrics })

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	filterFlags := []string{"cloud", "region", "availability", "cluster-type", "network-type"}

	filters := make([]string, len(filterFlags))
	for i, flag := range filterFlags {
		if cmd.Flags().Changed(flag) {
			filter, err := cmd.Flags().GetString(flag)
			if err != nil {
				return err
			}
			if filter == "" {
				return fmt.Errorf("`--%s` must not be empty", flag)
			}
			filters[i] = filter
		}
	}

	metric, err := cmd.Flags().GetString("metric")
	if err != nil {
		return err
	}

	legacy, err := cmd.Flags().GetBool("legacy")
	if err != nil {
		return err
	}

	rows, err := c.listRows(filters, metric, legacy)
	if err != nil {
		return err
	}

	return printTable(cmd, rows)
}

func (c *command) listRows(filters []string, metric string, legacy bool) ([]row, error) {
	org := &ccloudv1.Organization{Id: c.Context.GetOrganization().GetId()}

	kafkaPricesReply, err := c.Client.Billing.GetPriceTable(context.Background(), org, "kafka")
	if err != nil {
		return nil, err
	}

	clusterLinkPricesReply, err := c.Client.Billing.GetPriceTable(context.Background(), org, "cluster-link")
	if err != nil {
		return nil, err
	}

	kafkaTable, err := filterTable(kafkaPricesReply.PriceTable, filters, metric, legacy)
	if err != nil {
		return nil, err
	}

	clusterLinkTable, err := filterTable(clusterLinkPricesReply.PriceTable, filters, metric, legacy)
	if err != nil {
		return nil, err
	}

	// Merge cluster link price table into kafka table
	// Kafka metrics and cluster link metrics will not have overlap
	for metrics, val := range clusterLinkTable {
		kafkaTable[metrics] = val
	}
	if len(kafkaTable) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	var rows []row
	for metric, val := range kafkaTable {
		for key, price := range val.Prices {
			x := strings.Split(key, ":")
			availability, clusterType, networkType := x[2], x[3], x[4]

			rows = append(rows, row{
				metric:       metric,
				availability: availability,
				clusterType:  clusterType,
				networkType:  networkType,
				price:        price,
				unit:         val.Unit,
			})
		}
	}

	return rows, nil
}

func filterTable(table map[string]*ccloudv1.UnitPrices, filters []string, metric string, legacy bool) (map[string]*ccloudv1.UnitPrices, error) {
	filteredTable := make(map[string]*ccloudv1.UnitPrices)

	for service, val := range table {
		if metric != "" && service != metric {
			continue
		}

		for key, price := range val.Prices {
			fields := strings.Split(key, ":")

			shouldContinue := false
			for i, val := range filters {
				if val != "" && fields[i] != val {
					shouldContinue = true
				}
			}
			if shouldContinue {
				continue
			}

			// Hide legacy cluster types unless --legacy flag is enabled
			if types.Contains([]string{"standard", "custom"}, fields[3]) && !legacy {
				continue
			}

			if price == 0 {
				continue
			}

			if _, ok := filteredTable[service]; !ok {
				filteredTable[service] = &ccloudv1.UnitPrices{
					Prices: make(map[string]float64),
					Unit:   val.Unit,
				}
			}

			filteredTable[service].Prices[key] = price
		}
	}

	return filteredTable, nil
}

func printTable(cmd *cobra.Command, rows []row) error {
	list := output.NewList(cmd)
	for _, row := range rows {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&humanOut{
				Metric:       formatMetric[row.metric],
				ClusterType:  formatClusterTypeHuman[row.clusterType],
				Availability: formatAvailability[row.availability],
				NetworkType:  formatNetworkType[row.networkType],
				Price:        formatPrice(row.price, row.unit),
			})
		} else {
			list.Add(&serializedOut{
				Metric:       row.metric,
				ClusterType:  formatClusterTypeSerialized[row.clusterType],
				Availability: row.availability,
				NetworkType:  row.networkType,
				Price:        row.price,
			})
		}
	}
	return list.Print()
}

func formatPrice(price float64, unit string) string {
	priceStr := fmt.Sprint(price)

	// Require >= 2 digits after the decimal
	if strings.Contains(priceStr, ".") {
		// Extend the remainder if needed
		r := strings.Split(priceStr, ".")
		for len(r[1]) < 2 {
			r[1] += "0"
		}
		priceStr = strings.Join(r, ".")
	} else {
		priceStr += ".00"
	}

	return fmt.Sprintf("$%s USD/%s", priceStr, unit)
}
