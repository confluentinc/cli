package quit

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/feedback"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// NewQuitCmd returns the Cobra command for quitting the shell.
func NewQuitCmd(prerunner pcmd.PreRunner, config *v3.Config, logger *log.Logger, client analytics.Client) *cobra.Command {
	quitCmd := pcmd.NewAnonymousCLICommand(&cobra.Command{
		Use:   "quit",
		Short: fmt.Sprintf("Exit the %s shell\n", config.CLIName),
		Args:  cobra.NoArgs,
	}, prerunner)
	quitCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Exiting %s shell.\n", config.CLIName)
		// For quit pcmd.
		_ = client.SendCommandAnalytics(quitCmd.Command, args, nil)
		// For shell pcmd.
		_ = client.SendCommandAnalytics(quitCmd.Command.Parent(), args, nil)
		err := client.Close()
		if err != nil {
			logger.Debug(err)
		}
		feedback.HandleFeedbackNudge(config.CLIName, args)
		os.Exit(0)
	}
	return quitCmd.Command
}
