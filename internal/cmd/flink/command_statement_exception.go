package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"time"

	"github.com/spf13/cobra"
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
	c.addRegionFlag(cmd)
	cmd.AddCommand(c.newStatementExceptionListCommand())

	return cmd
}
