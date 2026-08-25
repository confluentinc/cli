package usage

import (
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	"github.com/confluentinc/cli/v4/pkg/agentdetect"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/log"
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

// CollectAgentDetect runs agent detection and assigns the results onto this
// Usage's agent-detect fields. Detect degrades to empty signals on its own
// (walk timeout, depth cap, lookup failure) rather than returning an error, so
// this can never fail the invocation; the recover is only a backstop against a
// panic escaping Detect.
//
// The CliV1Usage fields assigned below are flat additions from APIE-1607 (see
// https://confluentinc.atlassian.net/wiki/spaces/AEGI/pages/6089736699), named
// and typed to mirror agentdetect.Attributes one field at a time. This won't
// compile until that schema update lands in ccloud-sdk-go-v2.
func (u *Usage) CollectAgentDetect() {
	defer func() {
		if r := recover(); r != nil {
			log.CliLogger.Tracef("agent detection panicked: %v", r)
		}
	}()

	attrs := agentdetect.Detect(agentdetect.Options{}).Attributes()

	u.AgentEnv = optionalStrings(attrs.AgentEnv)
	u.AgentProc = attrs.AgentProc
	u.AgentArgv = attrs.AgentArgv
	u.IdeHost = attrs.IDEHost
	u.Interactive = cliv1.PtrString(attrs.Interactive)
	u.ChainShape = cliv1.PtrString(attrs.ChainShape)
	u.Wrappers = optionalStrings(attrs.Wrappers)
	u.Ci = optionalStrings(attrs.CI)
	u.AgentTables = cliv1.PtrString(attrs.Tables)
}

// optionalStrings mirrors the CliV1Usage convention (see Flags, StackFrames):
// an unset slice is a nil pointer, not a pointer to an empty slice.
func optionalStrings(s []string) *[]string {
	if len(s) == 0 {
		return nil
	}
	return &s
}

// Report sends usage data to cc-cli-usage-service.
func (u *Usage) Report(client *ccloudv2.Client) {
	if err := client.CreateCliUsage(cliv1.CliV1Usage(*u)); err != nil {
		log.CliLogger.Warnf("Failed to report CLI usage: %v", err)
	}
}
