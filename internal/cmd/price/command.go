package price

import (
	"fmt"
	"sort"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	listFields       = []string{"metric", "clusterType", "availability", "networkType", "price"}
	humanLabels      = []string{"Metric", "Cluster Type", "Availability", "Network Type", "Price"}
	structuredLabels = []string{"metric", "cluster_type", "availability", "network_type", "price"}

	formatCloud = map[string]string{
		"aws":   "AWS",
		"azure": "Azure",
		"gcp":   "GCP",
	}

	formatMetric = map[string]string{
		"ConnectNumRecords": "Connect record",
		"ConnectNumTasks":   "Connect task",
		"ConnectThroughput": "Connect throughput",
		"KSQLNumCSUs":       "ksqlDB capacity",
		"KafkaBase":         "Kafka base price",
		"KafkaCKUUnit":      "CKU unit price",
		"KafkaNetworkRead":  "Kafka read",
		"KafkaNetworkWrite": "Kafka write",
		"KafkaNumCKUs":      "CKU price",
		"KafkaPartition":    "Kafka partition",
		"KafkaStorage":      "Kafka storage",
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

type command struct {
	*cmd.AuthenticatedCLICommand
}

func New(prerunner cmd.PreRunner) *cobra.Command {
	c := &command{
		cmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "price",
				Short: "See Confluent Cloud pricing information.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
	}

	c.AddCommand(c.newListCommand())

	return c.Command
}

func (c *command) newListCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "Print an organization's price list.",
		Args:  cobra.NoArgs,
		RunE:  cmd.NewCLIRunE(c.list),
	}

	// Required flags
	command.Flags().String("cloud", "", fmt.Sprintf("Cloud provider (%s).", listFromMap(formatCloud)))
	_ = command.MarkFlagRequired("cloud")
	command.Flags().String("region", "", "Cloud region ID for cluster (use `ccloud kafka region list` to see all).")
	_ = command.MarkFlagRequired("region")

	// Extra filtering flags
	command.Flags().String("availability", "", fmt.Sprintf("Filter by availability (%s).", listFromMap(formatAvailability)))
	command.Flags().String("cluster-type", "", fmt.Sprintf("Filter by cluster type (%s).", listFromMap(formatClusterType)))
	command.Flags().String("metric", "", fmt.Sprintf("Filter by metric (%s).", listFromMap(formatMetric)))
	command.Flags().String("network-type", "", fmt.Sprintf("Filter by network type (%s).", listFromMap(formatNetworkType)))

	command.Flags().Bool("legacy", false, "Show legacy cluster types.")
	command.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

	command.Flags().SortFlags = false

	return command
}

type humanRow struct {
	metric       string
	clusterType  string
	availability string
	networkType  string
	price        string
}

type structuredRow struct {
	metric       string
	clusterType  string
	availability string
	networkType  string
	price        float64
}

func (c *command) list(command *cobra.Command, _ []string) error {
	o, err := command.Flags().GetString("output")
	if err != nil {
		return err
	}

	org := &orgv1.Organization{Id: c.State.Auth.User.OrganizationId}
	res, err := c.Client.Organization.GetPriceTable(nil, org)
	if err != nil {
		return err
	}

	w, err := output.NewListOutputCustomizableWriter(command, listFields, humanLabels, structuredLabels, command.OutOrStdout())
	if err != nil {
		return err
	}

	table, err := filterTable(command, res.PriceTable)
	if err != nil {
		return err
	}

	if len(table) == 0 {
		return fmt.Errorf("no results found")
	}

	for metric, val := range table {
		for key, price := range val.Prices {
			args := strings.Split(key, ":")
			availability := args[2]
			clusterType := args[3]
			networkType := args[4]

			switch o {
			case "human":
				w.AddElement(&humanRow{
					metric:       formatMetric[metric],
					clusterType:  formatClusterType[clusterType],
					availability: formatAvailability[availability],
					networkType:  formatNetworkType[networkType],
					price:        formatPrice(price, val.Unit),
				})
			case "json", "yaml":
				w.AddElement(&structuredRow{
					metric:       metric,
					clusterType:  clusterType,
					availability: availability,
					networkType:  networkType,
					price:        price,
				})
			}
		}
	}
	w.StableSort()

	return w.Out()
}

func filterTable(command *cobra.Command, table map[string]*orgv1.UnitPrices) (map[string]*orgv1.UnitPrices, error) {
	metric, err := command.Flags().GetString("metric")
	if err != nil {
		return nil, err
	}

	filters := []string{"cloud", "region", "availability", "cluster-type", "network-type"}

	filterValues := make([]string, len(filters))
	for i, filter := range filters {
		var err error
		filterValues[i], err = command.Flags().GetString(filter)
		if err != nil {
			return nil, err
		}
	}

	legacy, err := command.Flags().GetBool("legacy")
	if err != nil {
		return nil, err
	}

	filteredTable := make(map[string]*orgv1.UnitPrices)

	for service, val := range table {
		if metric != "" && service != metric {
			continue
		}

		for key, price := range val.Prices {
			args := strings.Split(key, ":")

			shouldContinue := false
			for i, val := range filterValues {
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
				filteredTable[service] = &orgv1.UnitPrices{
					Prices: make(map[string]float64),
					Unit:   val.Unit,
				}
			}

			filteredTable[service].Prices[key] = price
		}
	}

	return filteredTable, nil
}

func formatPrice(price float64, unit string) string {
	if price < 0.005 {
		// Needs to be represented by a fractional dollar amount.
		return fmt.Sprintf("$%.1g USD/%s", price, unit)
	}
	return fmt.Sprintf("$%.2f USD/%s", price, unit)
}

func listFromMap(m map[string]string) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
