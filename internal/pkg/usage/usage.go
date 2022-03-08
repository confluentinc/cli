package usage

import (
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/log"
)

type Usage struct {
	OS      string   `json:"os"`
	Arch    string   `json:"arch"`
	Version string   `json:"version"`
	Command string   `json:"command"`
	Flags   []string `json:"flags"`
	Error   bool     `json:"error"`
}

// Collect collects usage data, such as the command name and flags.
func Collect(cmd *cobra.Command) *Usage {
	usage := &Usage{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: cmd.Version,
		Flags:   []string{},
	}

	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		usage.Command = cmd.CommandPath()

		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if flag.Changed {
				usage.Flags = append(usage.Flags, flag.Name)
			}
		})
	}

	return usage
}

// Report sends usage data to cc-cli-usage-service.
func Report(usage *Usage) {
	if usage.Command != "" {
		log.CliLogger.Debug(usage)
		// TODO: Send data, log failure
	}
}
