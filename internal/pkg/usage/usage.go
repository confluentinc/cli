package usage

import (
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type Usage cliv1.CliV1Usage

func New(version string) *Usage {
	return &Usage{
		Os:      cliv1.PtrString(runtime.GOOS),
		Arch:    cliv1.PtrString(runtime.GOARCH),
		Version: cliv1.PtrString(version),
	}
}

// Collect is a post-run function that collects the command name and flag names. The error boolean is collected later.
func (u *Usage) Collect(cmd *cobra.Command, _ []string) {
	u.Command = cliv1.PtrString(cmd.CommandPath())

	var flags []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Changed {
			flags = append(flags, flag.Name)
		}
	})
	u.Flags = &flags
}

// PanicCollect provides the same functionality above but for when panics occur and command execution is not completed.
func (u *Usage) PanicCollect(cmd *cobra.Command, args []string) {
	fullCommand, flags, _ := cmd.Find(args)
	for i := range flags {
		flags[i] = strings.TrimPrefix(flags[i], "--")
		flags[i] = strings.TrimPrefix(flags[i], "-")
	}
	var formattedFlags []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if slices.Contains(flags, flag.Name) {
			formattedFlags = append(formattedFlags, flag.Name)
		} else if slices.Contains(flags, flag.Shorthand) {
			formattedFlags = append(formattedFlags, flag.Name)
		}
	})
	u.Command = cliv1.PtrString(fullCommand.CommandPath())
	u.Flags = &formattedFlags
}

// Report sends usage data to cc-cli-usage-service.
func (u *Usage) Report(client *ccloudv2.Client) {
	if err := client.CreateCliUsage(cliv1.CliV1Usage(*u)); err != nil {
		log.CliLogger.Warnf("Failed to report CLI usage: %v", err)
	}
}
