package flink

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type exceptionOut struct {
	Timestamp  time.Time `human:"Timestamp" serialized:"timestamp"`
	Name       string    `human:"Name" serialized:"name"`
	StackTrace string    `human:"Stack Trace" serialized:"stack_trace"`
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
