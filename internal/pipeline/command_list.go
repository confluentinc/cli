package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display pipelines in the current environment and cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	pipelines, err := c.getPipelines()
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
