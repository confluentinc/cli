package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
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

func NewConnectorsCommand(prerunner cmd.PreRunner) *cobra.Command {
	connectorsCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "connectors [command]",
			Short: "Manage connectors.",
			Args:  cobra.ExactArgs(1),
		},
		prerunner)

	connectorsCommand.AddCommand(NewListConnectorsCommand(prerunner))

	return connectorsCommand.Command
}

func NewListConnectorsCommand(prerunner cmd.PreRunner) *cobra.Command {
	connectorsCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all available connectors.",
			Args:  cobra.NoArgs,
			RunE:  runListConnectorsCommand,
		},
		prerunner)

	return connectorsCommand.Command
}

func runListConnectorsCommand(command *cobra.Command, _ []string) error {
	command.Println("Bundled Predefined Connectors:")
	command.Println(buildTabbedList(connectors))

	return nil
}

