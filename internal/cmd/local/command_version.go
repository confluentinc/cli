package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/local"
)

func NewVersionCommand(prerunner cmd.PreRunner) *cobra.Command {
	versionCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the Confluent Platform version.",
			Args:  cobra.NoArgs,
			RunE:  runVersionCommand,
		}, prerunner)

	return versionCommand.Command
}

func runVersionCommand(command *cobra.Command, _ []string) error {
	ch := local.NewConfluentHomeManager()

	isCP, err := ch.IsConfluentPlatform()
	if err != nil {
		return err
	}

	flavor := "Confluent Community Software"
	if isCP {
		flavor = "Confluent Platform"
	}

	version, err := ch.GetVersion(flavor)
	if err != nil {
		return err
	}

	cmd.Printf(command, "%s: %s\n", flavor, version)
	return nil
}
