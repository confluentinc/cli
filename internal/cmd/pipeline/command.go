package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Manage stream designer pipelines.",
	}
	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newListCommand(prerunner))
	c.AddCommand(c.newActivateCommand(prerunner))
	c.AddCommand(c.newDeactivateCommand(prerunner))
	c.AddCommand(c.newCreateCommand(prerunner))
	c.AddCommand(c.newDeleteCommand(prerunner))
	c.AddCommand(c.newUpdateCommand(prerunner))
	c.AddCommand(c.newDescribeCommand(prerunner))

	return c.Command
}
