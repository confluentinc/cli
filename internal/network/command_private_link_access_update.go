package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPrivateLinkAccessUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing private link access.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateLinkAccessArgs),
		RunE:              c.privateLinkAccessUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of private link access "pla-123456".`,
				Code: `confluent network private-link access update pla-123456 --name "new name"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the private link access.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) privateLinkAccessUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updatePrivateLinkAccess := networkingv1.NetworkingV1PrivateLinkAccessUpdate{
		Spec: &networkingv1.NetworkingV1PrivateLinkAccessSpecUpdate{
			DisplayName: networkingv1.PtrString(name),
			Environment: &networkingv1.ObjectReference{Id: environmentId},
		},
	}

	access, err := c.V2Client.UpdatePrivateLinkAccess(environmentId, args[0], updatePrivateLinkAccess)
	if err != nil {
		return err
	}

	return printPrivateLinkAccessTable(cmd, access)
}
