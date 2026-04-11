package kafka

import (
	"github.com/spf13/cobra"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *clusterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka clusters.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().Bool("all", false, "List clusters across all environments.")
	cmd.Flags().Bool("deletion-protection", false, `Filter by deletion protection status. Use "--deletion-protection=false" to list unprotected clusters.`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	var deletionProtection *bool
	if cmd.Flags().Changed("deletion-protection") {
		dp, err := cmd.Flags().GetBool("deletion-protection")
		if err != nil {
			return err
		}
		deletionProtection = &dp
	}

	var clusters []cmkv2.CmkV2Cluster
	if all {
		environments, err := c.V2Client.ListOrgEnvironments()
		if err != nil {
			return err
		}

		for _, env := range environments {
			clustersOfEnvironment, err := c.V2Client.ListKafkaClustersWithDeletionProtection(*env.Id, deletionProtection)
			if err != nil {
				return err
			}
			clusters = append(clusters, clustersOfEnvironment...)
		}
	} else {
		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return err
		}

		clusters, err = c.V2Client.ListKafkaClustersWithDeletionProtection(environmentId, deletionProtection)
		if err != nil {
			return err
		}
	}

	if activeEndpoint := c.Context.KafkaClusterContext.GetActiveKafkaClusterEndpoint(); activeEndpoint != "" {
		if output.GetFormat(cmd) == output.Human {
			output.Printf(c.Config.EnableColor, "The current endpoint is set to %q, "+
				"use `kafka cluster endpoint list` to view the available endpoints\n", activeEndpoint)
		}
	}

	list := output.NewList(cmd)
	for _, cluster := range clusters {
		list.Add(convertClusterToDescribeStruct(&cluster, nil, c.Context))
	}
	list.Filter([]string{"IsCurrent", "Id", "Name", "Type", "Cloud", "Region", "Availability", "Network", "Status", "DeletionProtection", "Endpoint"})
	return list.Print()
}
