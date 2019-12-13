package quit

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/spf13/cobra"
	"os"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

// NewQuitCmd returns the Cobra command for quitting the shell.
func NewQuitCmd(prerunner pcmd.PreRunner, cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:               "quit",
		Short:             "Exit the " + cfg.CLIName + " shell",
		PersistentPreRunE: prerunner.Anonymous(),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Exiting %s\n", cfg.CLIName)
			os.Exit(0)
		},
		Args: cobra.NoArgs,
	}
}
