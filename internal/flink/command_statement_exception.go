package flink

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type exceptionOut struct {
	Timestamp time.Time `human:"Timestamp" serialized:"timestamp"`
	Name      string    `human:"Name" serialized:"name"`
	Message   string    `human:"Message" serialized:"message"`
}

type exceptionOutOnPrem struct {
	Timestamp string `human:"Timestamp" serialized:"timestamp"`
	Name      string `human:"Name" serialized:"name"`
	Message   string `human:"Message" serialized:"message"`
}

func (c *command) newStatementExceptionCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exception",
		Short: "Manage Flink SQL statement exceptions.",
	}

	if cfg.IsCloudLogin() {
		pcmd.AddCloudFlag(cmd)
		pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
		cmd.AddCommand(c.newStatementExceptionListCommand())
	} else {
		cmd.AddCommand(c.newStatementExceptionListCommandOnPrem())
	}

	return cmd
}
