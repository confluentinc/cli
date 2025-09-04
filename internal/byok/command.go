package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

const byokUnknownKeyTypeErrorMsg = "unknown byok key type"

type command struct {
	*pcmd.AuthenticatedCLICommand
}

type out struct {
	Id                string   `human:"ID" serialized:"id"`
	DisplayName       string   `human:"Display Name,omitempty" serialized:"display_name,omitempty"`
	Key               string   `human:"Key" serialized:"key"`
	Roles             []string `human:"Roles" serialized:"roles"`
	Cloud             string   `human:"Cloud" serialized:"cloud"`
	State             string   `human:"State" serialized:"state"`
	CreatedAt         string   `human:"Created At" serialized:"created_at"`
	ValidationPhase   string   `human:"Validation Phase" serialized:"validation_phase"`
	ValidationSince   string   `human:"Validation Since" serialized:"validation_since"`
	ValidationRegion  string   `human:"Validation Region,omitempty" serialized:"validation_region,omitempty"`
	ValidationMessage string   `human:"Validation Message,omitempty" serialized:"validation_message,omitempty"`
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "byok",
		Short:       "Manage your keys in Confluent Cloud.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *command) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteByokKeyIds(c.V2Client)
}
