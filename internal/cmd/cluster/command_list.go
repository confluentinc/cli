package cluster

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/antihax/optional"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

const (
	typeFlag                      = "type"
	connectClusterTypeName        = "connect-cluster"
	kafkaClusterTypeName          = "kafka-cluster"
	ksqlClusterTypeName           = "ksql-cluster"
	schemaRegistryClusterTypeName = "schema-registry-cluster"
)

var (
	clusterFields    = []string{"Name", "Type", "ID", "Hosts"}
	clusterLabels    = []string{"Name", "Type", "ID", "Hosts"}
	clusterTypeNames = []string{connectClusterTypeName, kafkaClusterTypeName, ksqlClusterTypeName, schemaRegistryClusterTypeName}
)

type listCommand struct {
	*cmd.AuthenticatedCLICommand
}

type prettyCluster struct {
	Name  string
	Type  string
	ID    string
	Hosts string
}

// NewListCommand returns the sub-command object for listing clusters
func NewListCommand(cfg *v3.Config, prerunner cmd.PreRunner) *cobra.Command {
	listCmd := &listCommand{
		AuthenticatedCLICommand: cmd.NewAuthenticatedWithMDSCLICommand(
			&cobra.Command{
				Use:   "list",
				Short: "List registered clusters.",
				Long:  "List clusters that are registered with the MDS cluster registry.",
			},
			cfg, prerunner),
	}
	listCmd.Flags().String(typeFlag, "", fmt.Sprintf("Filter list to this cluster type (%s).", strings.Join(clusterTypeNames, ", ")))
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	listCmd.RunE = listCmd.list
	return listCmd.Command
}

func (c *listCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}

func (c *listCommand) list(cmd *cobra.Command, args []string) error {
	var clusterType optional.String
	t, err := cmd.Flags().GetString(typeFlag)
	if err == nil && t != "" {
		known := false
		for _, ctn := range clusterTypeNames {
			if ctn == t {
				known = true
				break
			}
		}
		if !known {
			return fmt.Errorf("%s should be one of %s", typeFlag, strings.Join(clusterTypeNames, ", "))
		}
		clusterType = optional.NewString(t)
	}
	ct := &mds.ClusterRegistryListOpts{
		ClusterType: clusterType,
	}
	clusterInfos, response, err := c.MDSClient.ClusterRegistryApi.ClusterRegistryList(c.createContext(), ct)
	if err != nil {
		return c.handleClusterError(cmd, err, response)
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if format == output.Human.String() {
		var data [][]string
		for _, clusterInfo := range clusterInfos {
			clusterDisplay, err := createPrettyCluster(clusterInfo)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
			data = append(data, printer.ToRow(clusterDisplay, clusterFields))
		}
		outputTable(data)
	} else {
		return output.StructuredOutput(format, clusterInfos)
	}
	return nil
}

func (c *listCommand) handleClusterError(cmd *cobra.Command, err error, response *http.Response) error {
	if response != nil && response.StatusCode == http.StatusNotFound {
		cmd.SilenceUsage = true
		return fmt.Errorf("Unable to access Cluster Registry (%s). Ensure that you're running against MDS with CP 6.0+.", err.Error())
	}
	return errors.HandleCommon(err, cmd)
}

func createPrettyHost(hostInfo mds.HostInfo) (string, error) {
	if hostInfo.Port > 0 {
		return fmt.Sprintf("%s:%d", hostInfo.Host, hostInfo.Port), nil
	}
	return hostInfo.Host, nil
}

func createPrettyCluster(clusterInfo mds.ClusterInfo) (*prettyCluster, error) {
	var t, id string
	switch {
	case clusterInfo.Scope.Clusters.ConnectCluster != "":
		t = connectClusterTypeName
		id = clusterInfo.Scope.Clusters.ConnectCluster
	case clusterInfo.Scope.Clusters.KsqlCluster != "":
		t = ksqlClusterTypeName
		id = clusterInfo.Scope.Clusters.KsqlCluster
	case clusterInfo.Scope.Clusters.SchemaRegistryCluster != "":
		t = schemaRegistryClusterTypeName
		id = clusterInfo.Scope.Clusters.SchemaRegistryCluster
	default:
		t = kafkaClusterTypeName
		id = clusterInfo.Scope.Clusters.KafkaCluster
	}
	hosts := make([]string, len(clusterInfo.Hosts))
	for i, hostInfo := range clusterInfo.Hosts {
		hosts[i], _ = createPrettyHost(hostInfo)
	}
	return &prettyCluster{
		clusterInfo.Name,
		t,
		id,
		strings.Join(hosts, ", "),
	}, nil
}

func outputTable(data [][]string) {
	tablePrinter := tablewriter.NewWriter(os.Stdout)
	tablePrinter.SetAutoWrapText(false)
	tablePrinter.SetAutoFormatHeaders(false)
	tablePrinter.SetHeader(clusterLabels)
	tablePrinter.AppendBulk(data)
	tablePrinter.SetBorder(false)
	tablePrinter.Render()
}
