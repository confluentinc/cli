package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "flink",
		Short:       "Manage Apache Flink.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newComputePoolCommand())
	cmd.AddCommand(c.newRegionCommand())
	cmd.AddCommand(c.newStatementCommand())

	return cmd
}

func (c *command) autocompleteRegions(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return nil
	}

	regions, err := c.V2Client.ListFlinkRegions(cloud)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(regions))
	for i, region := range regions {
		suggestions[i] = region.GetId()
	}
	return suggestions
}
