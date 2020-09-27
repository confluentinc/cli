package schema_registry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/shell/completer"

	"github.com/confluentinc/cli/internal/pkg/log"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
	logger          *log.Logger
	srClient        *srsdk.APIClient
	prerunner       pcmd.PreRunner
	serverCompleter completer.ServerSideCompleter
}

func New(cliName string, prerunner pcmd.PreRunner, srClient *srsdk.APIClient, logger *log.Logger, serverCompleter completer.ServerSideCompleter) *cobra.Command {
	cliCmd := pcmd.NewCLICommand(
		&cobra.Command{
			Use:   "schema-registry",
			Short: `Manage Schema Registry.`,
		}, prerunner)
	cmd := &command{
		CLICommand:      cliCmd,
		srClient:        srClient,
		logger:          logger,
		prerunner:       prerunner,
		serverCompleter: serverCompleter,
	}
	cmd.init(cliName)
	return cmd.Command
}

func (c *command) init(cliName string) {
	if cliName == "ccloud" {
		clusterCmd := NewClusterCommand(cliName, c.prerunner, c.srClient, c.logger)
		subjectCmd := NewSubjectCommand(cliName, c.prerunner, c.srClient)
		schemaCmd := NewSchemaCommand(cliName, c.prerunner, c.srClient)

		c.AddCommand(clusterCmd)
		c.AddCommand(subjectCmd.Command)
		c.AddCommand(schemaCmd)

		c.serverCompleter.AddCommand(subjectCmd)
	} else {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
	}
}
