package flink

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type exceptionOut struct {
	Timestamp time.Time `human:"Timestamp" serialized:"timestamp"`
	Name      string    `human:"Name" serialized:"name"`
	Message   string    `human:"Message" serialized:"message"`
}

func (c *command) newStatementExceptionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exception",
		Short: "Manage Flink SQL statement exceptions.",
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	cmd.AddCommand(c.newStatementExceptionListCommand())

	return cmd
}
