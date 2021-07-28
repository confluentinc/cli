package quit

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// NewQuitCmd returns the Cobra command for quitting the shell.
func NewQuitCmd(prerunner pcmd.PreRunner) *cobra.Command {
	quitCmd := pcmd.NewAnonymousCLICommand(&cobra.Command{
		Use:   "quit",
		Short: fmt.Sprintf("Exit the %s shell.", version.CLIName),
		Args:  cobra.NoArgs,
	}, prerunner)
	quitCmd.Run = func(cmd *cobra.Command, args []string) {
		os.Exit(0)
	}
	return quitCmd.Command
}
