package panic

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/usage"
)

type Panic struct {
	ErrorMsg string
}

func (p *Panic) Error() string {
	return p.ErrorMsg
}

// CollectPanic collects relevant usage data for when panics occur and command execution is not completed.
func CollectPanic(cmd *cobra.Command, args []string, cfg *v1.Config) *usage.Usage {
	u := usage.New(cfg.Version.Version)
	fullCommand, flags, _ := cmd.Find(args)
	u.Command = cliv1.PtrString(fullCommand.CommandPath())
	u.Flags = parseFlags(fullCommand, flags)
	u.Error = cliv1.PtrBool(true)
	u.StackFrames = parseStack(string(debug.Stack()))
	return u
}

// FormatPanicMsg formats the returned value of the recover function when panics occur
func FormatPanicMsg(panicMsg any) string {
	var formattedMsg string
	switch panicMsg.(type) {
	default:
		formattedMsg = fmt.Sprintf("Error: %v", panicMsg)
	case error:
		formattedMsg = strings.ReplaceAll(panicMsg.(error).Error(), "runtime error", "Error")
	}
	return formattedMsg
}

// parseFlags collects the flags alongside the panicking command
func parseFlags(cmd *cobra.Command, flags []string) *[]string {
	var formattedFlags []string
	for i := range flags {
		flags[i] = strings.TrimPrefix(flags[i], "--")
		flags[i] = strings.TrimPrefix(flags[i], "-")
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if slices.Contains(flags, flag.Name) {
			formattedFlags = append(formattedFlags, flag.Name)
		} else if slices.Contains(flags, flag.Shorthand) {
			formattedFlags = append(formattedFlags, flag.Name)
		}
	})
	return &formattedFlags
}

// parseStack formats the stack trace resulting from a panic to only include line numbers up until panic
func parseStack(stack string) *[]string {
	trace := strings.Split(stack, "\n")
	if trace[len(trace)-1] == "" {
		trace = trace[:len(trace)-1]
	}
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
		if strings.Contains(s, ".go") {
			result = append(result, s)
		}
	}
	return &result
}
