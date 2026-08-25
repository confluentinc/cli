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

// CollectAgentDetect runs agent detection and records the results for usage
// reporting. Detect degrades to empty signals on its own (walk timeout, depth
// cap, lookup failure) rather than returning an error, so this can never fail
// the invocation; the recover is only a backstop against a panic escaping Detect.
//
// TODO(APIE-1608): assign attrs onto the new CliV1Usage agent-detect fields once
// APIE-1607 lands the schema update. Until then the computed Attributes are only
// trace-logged.
func (u *Usage) CollectAgentDetect() {
	defer func() {
		if r := recover(); r != nil {
			log.CliLogger.Tracef("agent detection panicked: %v", r)
		}
	}()

	attrs := agentdetect.Detect(agentdetect.Options{}).Attributes()
	log.CliLogger.Tracef("agent detection attributes: %+v", attrs)
}

// Report sends usage data to cc-cli-usage-service.
func (u *Usage) Report(client *ccloudv2.Client) {
	if err := client.CreateCliUsage(cliv1.CliV1Usage(*u)); err != nil {
		log.CliLogger.Warnf("Failed to report CLI usage: %v", err)
	}
}
