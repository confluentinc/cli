package kafka

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *clusterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka clusters.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().Bool("all", false, "List clusters across all environments.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	listAllClusters, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	var clusters []cmkv2.CmkV2Cluster
	if listAllClusters {
		environments, err := c.V2Client.ListOrgEnvironments()
		if err != nil {
			return err
		}

		for _, env := range environments {
			clustersOfEnvironment, err := c.V2Client.ListKafkaClusters(*env.Id)
			if err != nil {
				return err
			}
			clusters = append(clusters, clustersOfEnvironment...)
		}
	} else {
		clusters, err = c.V2Client.ListKafkaClusters(c.EnvironmentId())
		if err != nil {
			return err
		}
	}

	list := output.NewList(cmd)
	for _, cluster := range clusters {
		row := convertClusterToDescribeStruct(&cluster)
		row.IsCurrent = *cluster.Id == c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
		list.Add(row)
	}
	list.Filter([]string{"IsCurrent", "Id", "Name", "Type", "ServiceProvider", "Region", "Availability", "Status"})
	return list.Print()
}
