package common

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/confluentinc/cli/version"
	"github.com/spf13/cobra"
)

// NewVersionCmd returns the Cobra command for the version.
func NewVersionCmd(version *version.Version) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the ccloud version",
		Long:  "Print the ccloud version",
		Run:   func(cmd *cobra.Command, args []string) {
			fmt.Printf(`ccloud - Confluent Cloud CLI

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
		Args:  cobra.NoArgs,
	}
}
