package version

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// Returns the Cobra command for the version.
func New(cliName string, prerunner pcmd.PreRunner, v *version.Version) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version",
			Short: fmt.Sprintf("Show version of the %s.", version.GetFullCLIName(cliName)),
			Run: func(cmd *cobra.Command, _ []string) {
				cmd.Println(v)
			},
			Args: cobra.NoArgs,
		}, prerunner)
	return cliCmd.Command
}
