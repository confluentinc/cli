package common

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/version"
)

// NewVersionCmd returns the Cobra command for the version.
func NewVersionCmd(version *version.Version, prompt command.Prompt) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the ccloud version",
		Long:  "Print the ccloud version",
		Run: func(cmd *cobra.Command, args []string) {
			shared.PrintVersion(version, prompt)
		},
		Args: cobra.NoArgs,
	}
}
