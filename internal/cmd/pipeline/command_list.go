package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand(prerunner pcmd.PreRunner) *cobra.Command {
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

	pipelines, err := c.V2Client.ListPipelines(c.EnvironmentId(), cluster.ID)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, pipeline := range pipelines {
		list.Add(&out{
			Id:          pipeline.GetId(),
			Name:        pipeline.Spec.GetDisplayName(),
			Description: pipeline.Spec.GetDescription(),
			KsqlCluster: pipeline.Spec.KsqlCluster.GetId(),
			State:       pipeline.Status.GetState(),
		})
	}
	list.Filter([]string{"Id", "Name", "Description", "KsqlCluster", "State"})
	return list.Print()
}
