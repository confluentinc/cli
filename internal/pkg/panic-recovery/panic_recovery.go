package panic_recovery

import (
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/usage"
)

// CollectPanic collects relevant usage data for when panics occur and command execution is not completed.
func CollectPanic(cmd *cobra.Command, args []string, cfg *v1.Config) *usage.Usage {
	fullCommand, flags, _ := cmd.Find(args)
	trimmedFlags := ParseFlags(fullCommand, flags)
	parsedStack := parseStack(string(debug.Stack()))
	return &usage.Usage{
		Os:          cliv1.PtrString(runtime.GOOS),
		Arch:        cliv1.PtrString(runtime.GOARCH),
		Version:     cliv1.PtrString(cfg.Version.Version),
		Command:     cliv1.PtrString(fullCommand.CommandPath()),
		Flags:       &trimmedFlags,
		Error:       cliv1.PtrBool(true),
		StackFrames: &parsedStack,
	}
}

// ParseFlags collects the flags of a command after being found with Find()
func ParseFlags(cmd *cobra.Command, flags []string) []string {
	var formattedFlags []string
	trimmedFlags := make([]string, len(flags))
	for i := range flags {
		trimmedFlags[i] = strings.TrimLeft(flags[i], "-")
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if slices.Contains(trimmedFlags, flag.Name) || slices.Contains(trimmedFlags, flag.Shorthand) {
			formattedFlags = append(formattedFlags, flag.Name)
		}
	})
	return formattedFlags
}

// parseStack formats the stack trace resulting from a panic-recovery to only include line numbers up until panic-recovery
func parseStack(stack string) []string {
	stack = strings.TrimRight(stack, "\n")
	trace := strings.Split(stack, "\n")
	panicIndex := 0
	for idx := range trace {
		trace[idx] = strings.TrimSpace(trace[idx])
		if strings.Contains(trace[idx], "panic-recovery.go") {
			panicIndex = idx
		}
	}
	trace = trace[panicIndex:]

	var result []string
	for _, s := range trace {
		if strings.Contains(s, ".go:") {
			result = append(result, s)
		}
	}
	return result
}
