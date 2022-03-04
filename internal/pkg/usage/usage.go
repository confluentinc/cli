package usage

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Usage struct {
	Version string   `json:"version"`
	Command string   `json:"command"`
	Flags   []string `json:"flags"`
	Error   bool     `json:"error"`
}

// Collect collects usage data, such as the command name and flags.
func Collect(cmd *cobra.Command) *Usage {
	usage := &Usage{
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

// CollectForHelpFunc collects usage data for commands passed with the --help flag, a special case not covered by Collect.
func CollectForHelpFunc(cmd *cobra.Command) *Usage {
	return &Usage{
		Version: cmd.Version,
		Command: cmd.CommandPath(),
		Flags:   []string{"help"},
	}
}

// Report sends usage data to the backend service
func Report(usage *Usage) {
	fmt.Println(usage)
	// TODO: Log failure
}
