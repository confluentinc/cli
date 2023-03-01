package connect

import (
	"fmt"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

const clusterType = "connect-cluster"

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage Connect clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := new(clusterCommand)

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
		c.AddCommand(c.newCreateCommand())
		c.AddCommand(c.newDeleteCommand())
		c.AddCommand(c.newDescribeCommand())
		c.AddCommand(c.newListCommand())
		c.AddCommand(c.newPauseCommand())
		c.AddCommand(c.newResumeCommand())
		c.AddCommand(c.newUpdateCommand())
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
		c.AddCommand(c.newListCommandOnPrem())
	}

	return c.Command
}

func (c *clusterCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConnectors(cmd)
}

func (c *clusterCommand) autocompleteConnectors(cmd *cobra.Command) []string {
	connectors, err := c.fetchConnectors(cmd)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(connectors))
	i := 0
	for _, connector := range connectors {
		suggestions[i] = fmt.Sprintf("%s\t%s", connector.Id.GetId(), connector.Info.GetName())
		i++
	}
	return suggestions
}

func (c *clusterCommand) fetchConnectors(cmd *cobra.Command) (map[string]connectv1.ConnectV1ConnectorExpansion, error) {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListConnectorsWithExpansions(c.EnvironmentId(cmd), kafkaCluster.ID, "id,info")
}
