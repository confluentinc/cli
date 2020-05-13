
package quit

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

// NewQuitCmd returns the Cobra command for quitting the shell.
func NewQuitCmd(config *v3.Config) *cobra.Command {
	return &cobra.Command{
		Use:               "quit",
		Short:             "Exit the " + config.CLIName + " shell",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Exiting %s shell.\n", config.CLIName)
			os.Exit(0)
		},
		Args: cobra.NoArgs,
	}
}
