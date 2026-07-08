package endpoint

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

type out struct {
	Id             string `human:"ID" serialized:"id"`
	DisplayName    string `human:"Display Name" serialized:"display_name"`
	SwitchoverPair string `human:"Switchover Pair" serialized:"switchover_pair"`
	Environment    string `human:"Environment" serialized:"environment"`
	Phase          string `human:"Phase" serialized:"phase"`
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage switchover endpoints.",
		Long:  "Manage switchover endpoints. This API is not yet implemented on the backend; commands will fail against a live Confluent Cloud environment.",
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newActivateCommand())

	return cmd
}
