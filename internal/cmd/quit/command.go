package quit

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

// NewQuitCmd returns the Cobra command for quitting the shell.
func NewQuitCmd(prerunner pcmd.PreRunner, config *v3.Config) *cobra.Command {
	quitCmd := pcmd.NewAnonymousCLICommand(&cobra.Command{
		Use:   "quit",
		Short: fmt.Sprintf("Exit the %s shell\n", config.CLIName),
		Args:  cobra.NoArgs,
	}, prerunner)
	quitCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Exiting %s shell.\n", config.CLIName)
		os.Exit(0)
	}
	return quitCmd.Command
}
