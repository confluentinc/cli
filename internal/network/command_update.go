package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing network.",
		Args:  cobra.ExactArgs(1),
		// TODO: Implement autocompletion after List Network is implemented.
		// ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE: c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of the network "n-abcde1" with a new name.`,
				Code: "confluent network update n-abcde1 --name new-test-network",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the network.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	flags := []string{
		"name",
	}
	if err := errors.CheckNoUpdate(cmd.Flags(), flags...); err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updateNetwork := networkingv1.NetworkingV1NetworkUpdate{Spec: &networkingv1.NetworkingV1NetworkSpecUpdate{}}

	if name != "" {
		updateNetwork.Spec.SetDisplayName(name)
	}

	network, err := c.V2Client.UpdateNetwork(environmentId, args[0], updateNetwork)
	if err != nil {
		return err
	}

	return printTable(cmd, network)
}
