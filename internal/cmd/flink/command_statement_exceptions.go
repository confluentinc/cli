package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"time"

	"github.com/spf13/cobra"
)

type exceptionOut struct {
	Timestamp  time.Time `human:"Timestamp" serialized:"timestamp"`
	Name       string    `human:"Name" serialized:"name"`
	Stacktrace string    `human:"Stacktrace" serialized:"stacktrace"`
}

func (c *command) newExceptionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exceptions",
		Short: "Manage Flink SQL statements.",
	}
	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	cmd.AddCommand(c.newStatementExceptionsListCommand())

	return cmd
}
