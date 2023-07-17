package kafka

import (
	"github.com/spf13/cobra"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	nameconversions "github.com/confluentinc/cli/internal/pkg/name-conversions"
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
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	var clusters []cmkv2.CmkV2Cluster
	if all {
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
		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return err
		}

		clusters, err = c.V2Client.ListKafkaClusters(environmentId)
		if err != nil {
			if environmentId, err = nameconversions.ConvertEnvironmentNameToId(environmentId, c.V2Client); err != nil {
				return err
			}
			if clusters, err = c.V2Client.ListKafkaClusters(environmentId); err != nil {
				return err
			}
		}
	}

	list := output.NewList(cmd)
	for _, cluster := range clusters {
		list.Add(convertClusterToDescribeStruct(&cluster, c.Context.Context))
	}
	list.Filter([]string{"IsCurrent", "Id", "Name", "Type", "ServiceProvider", "Region", "Availability", "Status"})
	return list.Print()
}
