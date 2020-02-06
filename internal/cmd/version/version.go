package version

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// NewVersionCmd returns the Cobra command for the version.
func NewVersionCmd(prerunner pcmd.PreRunner, version *version.Version) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the " + version.Binary + " CLI version.",
			Run: func(cmd *cobra.Command, args []string) {
				pcmd.Println(cmd, version)
			},
			Args: cobra.NoArgs,
		},
		&v2.Config{}, prerunner)
	return cliCmd.Command
}
