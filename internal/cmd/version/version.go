package version

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/terminal"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// NewVersionCmd returns the Cobra command for the version.
func NewVersionCmd(version *version.Version, prompt terminal.Prompt) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the ccloud version",
		Long:  "Print the ccloud version",
		Run: func(cmd *cobra.Command, args []string) {
			version.Print(prompt)
		},
		Args: cobra.NoArgs,
	}
}
