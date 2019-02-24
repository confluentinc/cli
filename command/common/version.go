package common

import (
	"runtime"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/version"
)

// NewVersionCmd returns the Cobra command for the version.
func NewVersionCmd(version *version.Version, prompt command.Prompt) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the ccloud version",
		Long:  "Print the ccloud version",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = prompt.Printf(`ccloud - Confluent Cloud CLI

Version:     %s
Git Ref:     %s
Build Date:  %s
Build Host:  %s
Go Version:  %s (%s/%s)
Development: %s
`, version.Version,
				version.Commit,
				version.BuildDate,
				version.BuildHost,
				runtime.Version(),
				runtime.GOOS,
				runtime.GOARCH,
				strconv.FormatBool(!version.IsReleased()))
		},
		Args: cobra.NoArgs,
	}
}
