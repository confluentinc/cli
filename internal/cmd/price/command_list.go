package price

import (
	"context"
	"fmt"
	"sort"
	"strings"

	billingv1 "github.com/confluentinc/cc-structs/kafka/billing/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
		"KafkaStorage":          "Kafka storage",
		"ClusterLinkingBase":    "Cluster linking base",
		"ClusterLinkingPerLink": "Cluster linking per-link",
		"ClusterLinkingRead":    "Cluster linking read",
		"ClusterLinkingWrite":   "Cluster linking write",
	}

	formatClusterType = map[string]string{
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
	metrics        = mapToSlice(formatMetric)
	clusterTypes   = mapToSlice(formatClusterType)
	availabilities = mapToSlice(formatAvailability)
	networkTypes   = mapToSlice(formatNetworkType)
)

type out struct {
	Metric       string `human:"Metric" serialized:"metric"`
	ClusterType  string `human:"Cluster Type" serialized:"cluster_type"`
	Availability string `human:"Availability" serialized:"availability"`
	NetworkType  string `human:"Network Type" serialized:"network_type"`
	Price        string `human:"Price" serialized:"price"`
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			filterFlags := []string{"cloud", "region", "availability", "cluster-type", "network-type"}

			filters := make([]string, len(filterFlags))
			for i, flag := range filterFlags {
				filters[i], _ = cmd.Flags().GetString(flag)
			}

			metric, _ := cmd.Flags().GetString("metric")
			legacy, _ := cmd.Flags().GetBool("legacy")

			rows, err := c.list(filters, metric, legacy)
			if err != nil {
				return err
			}

			return printTable(cmd, rows)
		},
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

func (c *command) list(filters []string, metric string, legacy bool) ([]row, error) {
	org := &orgv1.Organization{Id: c.State.Auth.Organization.Id}

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

func filterTable(table map[string]*billingv1.UnitPrices, filters []string, metric string, legacy bool) (map[string]*billingv1.UnitPrices, error) {
	filteredTable := make(map[string]*billingv1.UnitPrices)

	for service, val := range table {
		if metric != "" && service != metric {
			continue
		}

		for key, price := range val.Prices {
			args := strings.Split(key, ":")

			shouldContinue := false
			for i, val := range filters {
				if val != "" && args[i] != val {
					shouldContinue = true
				}
			}
			if shouldContinue {
				continue
			}

			// Hide legacy cluster types unless --legacy flag is enabled
			if utils.Contains([]string{"standard", "custom"}, args[3]) && !legacy {
				continue
			}

			if price == 0 {
				continue
			}

			if _, ok := filteredTable[service]; !ok {
				filteredTable[service] = &billingv1.UnitPrices{
					Prices: make(map[string]float64),
					Unit:   val.Unit,
				}
			}

			filteredTable[service].Prices[key] = price
		}
	}

	return filteredTable, nil
}

func mapToSlice(m map[string]string) []string {
	var slice []string
	for key := range m {
		slice = append(slice, key)
	}
	sort.Strings(slice)
	return slice
}

func printTable(cmd *cobra.Command, rows []row) error {
	list := output.NewList(cmd)
	for _, row := range rows {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&out{
				Metric:       formatMetric[row.metric],
				ClusterType:  formatClusterType[row.clusterType],
				Availability: formatAvailability[row.availability],
				NetworkType:  formatNetworkType[row.networkType],
				Price:        formatPrice(row.price, row.unit),
			})
		} else {
			list.Add(&out{
				Metric:       row.metric,
				ClusterType:  row.clusterType,
				Availability: row.availability,
				NetworkType:  row.networkType,
				Price:        fmt.Sprint(row.price),
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
