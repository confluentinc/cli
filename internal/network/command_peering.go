package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type peeringCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type peeringHumanOut struct {
	Id        string `human:"ID"`
	Name      string `human:"Name"`
	NetworkId string `human:"Network ID"`
	Cloud     string `human:"Cloud"`
	Phase     string `human:"Phase"`
}

type peeringSerializedOut struct {
	Id        string `serialized:"id"`
	Name      string `serialized:"name"`
	NetworkId string `serialized:"network_id"`
	Cloud     string `serialized:"cloud"`
	Phase     string `serialized:"phase"`
}

func newPeeringCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "peering",
		Short:       "Manage peering connections.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &peeringCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newPeeringListCommand())

	return cmd
}

func (c *peeringCommand) getPeerings() ([]networkingv1.NetworkingV1Peering, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListPeerings(environmentId)
}
