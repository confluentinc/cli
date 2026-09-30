package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/types"
)

type offsetCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newOffsetCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "offset",
		Short:       "Manage offsets for managed connectors.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &offsetCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newStatusCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func (c *offsetCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConnectors()
}

func (c *offsetCommand) autocompleteConnectors() []string {
	connectors, err := c.fetchConnectors()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(connectors))
	for i, name := range types.GetSortedKeys(connectors) {
		suggestions[i] = fmt.Sprintf("%s\t%s", connectors[name].Id.GetId(), connectors[name].Info.GetName())
	}

	return suggestions
}

func (c *offsetCommand) fetchConnectors() (map[string]connectv1.ConnectV1ConnectorExpansion, error) {
	kafkaCluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return nil, err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListConnectorsWithExpansions(environmentId, kafkaCluster.ID, "id,info")
}
