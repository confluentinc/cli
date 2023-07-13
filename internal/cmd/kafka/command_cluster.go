package kafka

import (
	"fmt"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

const (
	singleZone       = "single-zone"
	multiZone        = "multi-zone"
	lowAvailability  = "SINGLE_ZONE"
	highAvailability = "MULTI_ZONE"
)

var availabilitiesToHuman = map[string]string{
	lowAvailability:  singleZone,
	highAvailability: multiZone,
}

var availabilitiesToModel = map[string]string{
	singleZone: lowAvailability,
	multiZone:  highAvailability,
}

type clusterCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage Kafka clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &clusterCommand{}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
	}

	cmd.AddCommand(c.newCreateCommand(cfg))
	cmd.AddCommand(c.newDeleteCommand(cfg))
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newUpdateCommand(cfg))
	cmd.AddCommand(c.newUseCommand(cfg))

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newListCommand())
	} else {
		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}

func (c *clusterCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	clusters, err := c.V2Client.ListKafkaClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}
