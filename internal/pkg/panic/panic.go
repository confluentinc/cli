package panic

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
	return &usage.Usage{
		Os:          cliv1.PtrString(runtime.GOOS),
		Arch:        cliv1.PtrString(runtime.GOARCH),
		Version:     cliv1.PtrString(cfg.Version.Version),
		Command:     cliv1.PtrString(fullCommand.CommandPath()),
		Flags:       parseFlags(fullCommand, flags),
		Error:       cliv1.PtrBool(true),
		StackFrames: parseStack(string(debug.Stack())),
	}
}

// parseFlags collects the flags alongside the panicking command
func parseFlags(cmd *cobra.Command, flags []string) *[]string {
	var formattedFlags []string
	for i := range flags {
		flags[i] = strings.TrimLeft(flags[i], "-")
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if slices.Contains(flags, flag.Name) || slices.Contains(flags, flag.Shorthand) {
			formattedFlags = append(formattedFlags, flag.Name)
		}
	})
	return &formattedFlags
}

// parseStack formats the stack trace resulting from a panic to only include line numbers up until panic
func parseStack(stack string) *[]string {
	stack = strings.TrimRight(stack, "\n")
	trace := strings.Split(stack, "\n")
	panicIndex := 0
	for idx := range trace {
		trace[idx] = strings.TrimSpace(trace[idx])
		if strings.Contains(trace[idx], "panic.go") {
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
	return &result
}
