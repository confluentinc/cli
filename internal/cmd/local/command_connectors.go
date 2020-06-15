package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

var connectors = []string{
	"elasticsearch-sink",
	"file-source",
	"file-sink",
	"jdbc-source",
	"jdbc-sink",
	"hdfs-sink",
	"s3-sink",
}

func NewConnectorsCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectorsCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "connectors [command]",
			Short: "Manage connectors.",
			Args:  cobra.ExactArgs(1),
		},
		cfg, prerunner)

	connectorsCommand.AddCommand(NewListConnectorsCommand(prerunner, cfg))

	return connectorsCommand.Command
}

func NewListConnectorsCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectorsCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all available connectors.",
			Args:  cobra.NoArgs,
			RunE:  runListConnectorsCommand,
		},
		cfg, prerunner)

	return connectorsCommand.Command
}

func runListConnectorsCommand(command *cobra.Command, _ []string) error {
	command.Println("Bundled Predefined Connectors:")
	command.Println(buildTabbedList(connectors))

	return nil
}
