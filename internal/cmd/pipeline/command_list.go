package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display pipelines in the current environment and cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	environmentId, err := c.EnvironmentId()
	if err != nil {
		return err
	}

	pipelines, err := c.V2Client.ListPipelines(environmentId, cluster.ID)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, pipeline := range pipelines {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&humanOut{
				Id:          pipeline.GetId(),
				Name:        pipeline.Spec.GetDisplayName(),
				Description: pipeline.Spec.GetDescription(),
				KsqlCluster: pipeline.Spec.KsqlCluster.GetId(),
				State:       pipeline.Status.GetState(),
			})
		} else {
			list.Add(&serializedOut{
				Id:          pipeline.GetId(),
				Name:        pipeline.Spec.GetDisplayName(),
				Description: pipeline.Spec.GetDescription(),
				KsqlCluster: pipeline.Spec.KsqlCluster.GetId(),
				State:       pipeline.Status.GetState(),
			})
		}
	}
	list.Filter([]string{"Id", "Name", "Description", "KsqlCluster", "State"})
	return list.Print()
}
