package quit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const PanicKey = "quitCmdCalled"

// NewQuitCmd returns the Cobra command for quitting the shell.
func NewQuitCmd(prerunner cmd.PreRunner, config *v3.Config, logger *log.Logger, client analytics.Client) *cobra.Command {
	quitCmd := cmd.NewAnonymousCLICommand(&cobra.Command{
		Use:   "quit",
		Short: fmt.Sprintf("Exit the %s shell\n", config.CLIName),
		Args:  cobra.NoArgs,
	}, prerunner)
	quitCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Exiting %s shell.\n", config.CLIName)
		panic(PanicKey)
	}
	return quitCmd.Command
}
