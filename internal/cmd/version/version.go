package version

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// NewVersionCmd returns the Cobra command for the version.
func NewVersionCmd(prerunner pcmd.PreRunner, version *version.Version) *cobra.Command {
	return &cobra.Command{
		Use:               "version",
		Short:             "Print the " + version.Binary + " CLI version.",
		PersistentPreRunE: prerunner.Anonymous(),
		Run: func(cmd *cobra.Command, args []string) {
			pcmd.Println(cmd, version)
		},
		Args: cobra.NoArgs,
	}
}
