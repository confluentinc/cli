package pair

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

type out struct {
	Id           string `human:"ID" serialized:"id"`
	DisplayName  string `human:"Display Name" serialized:"display_name"`
	ActiveMember string `human:"Active Member" serialized:"active_member"`
	Environment  string `human:"Environment" serialized:"environment"`
	Phase        string `human:"Phase" serialized:"phase"`
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pair",
		Short: "Manage switchover pairs.",
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newTriggerSwitchCommand())

	return cmd
}
