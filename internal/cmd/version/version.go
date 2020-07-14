package version

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// Returns the Cobra command for the version.
func New(cliName string, prerunner pcmd.PreRunner, v *version.Version) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the " + version.GetFullCLIName(cliName) + " version.",
			Run: func(cmd *cobra.Command, args []string) {
				pcmd.Println(cmd, v)
			},
			Args: cobra.NoArgs,
		}, prerunner)
	return cliCmd.Command
}
