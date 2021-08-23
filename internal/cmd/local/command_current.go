package local

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func NewCurrentCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "current",
			Short: "Get the path of the current Confluent run.",
			Long:  fmt.Sprintf(`Print the filesystem path of the data and logs of the services managed by the current "%s local" command. If such a path does not exist, it will be created.`, pversion.CLIName),
			Args:  cobra.NoArgs,
			Example: examples.BuildExampleString(
				examples.Example{
					Text: fmt.Sprintf("In Linux, running `%s local current` should resemble the following:", pversion.CLIName),
					Code: "/tmp/confluent.SpBP4fQi",
				},
				examples.Example{
					Text: fmt.Sprintf("In macOS, running `%s local current` should resemble the following:", pversion.CLIName),
					Code: "/var/folders/cs/1rndf6593qb3kb6r89h50vgr0000gp/T/confluent.000000",
				},
			),
		}, prerunner)

	c.Command.RunE = cmd.NewCLIRunE(c.runCurrentCommand)
	return c.Command
}

func (c *Command) runCurrentCommand(command *cobra.Command, _ []string) error {
	dir, err := c.cc.GetCurrentDir()
	if err != nil {
		return err
	}

	utils.Println(command, dir)
	return nil
}
