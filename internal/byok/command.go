package byok

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

const byokUnknownKeyTypeErrorMsg = "unknown byok key type"

type command struct {
	*pcmd.AuthenticatedCLICommand
}

type humanOut struct {
	Id        string `human:"ID"`
	Key       string `human:"Key"`
	Roles     string `human:"Roles"`
	Provider  string `human:"Provider"`
	State     string `human:"State"`
	CreatedAt string `human:"Created At"`
}

type serializedOut struct {
	Id        string   `serialized:"id"`
	Key       string   `serialized:"key"`
	Roles     []string `serialized:"roles"`
	Provider  string   `serialized:"provider"`
	State     string   `serialized:"state"`
	CreatedAt string   `serialized:"created_at"`
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
