package version

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// New returns the Cobra command for the version.
func New(prerunner pcmd.PreRunner, version *version.Version) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the " + version.Binary + " CLI version.",
			Run: func(cmd *cobra.Command, args []string) {
				pcmd.Println(cmd, version)
			},
			Args: cobra.NoArgs,
		}, prerunner)
	return cliCmd.Command
}
