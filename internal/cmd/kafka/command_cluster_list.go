package kafka

import (
	"fmt"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/cmk"
	"github.com/confluentinc/cli/internal/pkg/org"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name", "Type", "ServiceProvider", "Region", "Availability", "Status"}
	listHumanLabels      = []string{"Id", "Name", "Type", "Provider", "Region", "Availability", "Status"}
	listStructuredLabels = []string{"id", "name", "type", "provider", "region", "availability", "status"}
)

func (c *clusterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List Kafka clusters.",
		RunE:  pcmd.NewCLIRunE(c.list),
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
		environments, _, err := org.ListEnvironments(c.OrgClient, c.AuthToken())
		if err != nil {
			return err
		}

		for _, env := range environments.Data {
			clusterList, _, err := cmk.ListKafkaClusters(c.CmkClient, *env.Id, c.AuthToken())
			if err != nil {
				return err
			}
			clusters = append(clusters, clusterList.Data...)
		}
	} else {
		clusterList, _, err := cmk.ListKafkaClusters(c.CmkClient, c.EnvironmentId(), c.AuthToken())
		if err != nil {
			fmt.Println("we got an err...")
			return err
		}
		clusters = clusterList.Data
		if err != nil {
			return err
		}
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		// Add '*' only in the case where we are printing out tables
		if outputWriter.GetOutputFormat() == output.Human {
			if *cluster.Id == c.Context.KafkaClusterContext.GetActiveKafkaClusterId() {
				*cluster.Id = fmt.Sprintf("* %s", *cluster.Id)
			} else {
				*cluster.Id = fmt.Sprintf("  %s", *cluster.Id)
			}
		}
		outputWriter.AddElement(convertClusterToDescribeStruct(&cluster))
	}

	return outputWriter.Out()
}
