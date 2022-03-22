package usage

import (
	"context"
	"runtime"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cli/v1"
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

func New() *Usage {
	return &Usage{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// Collect collects usage data, such as the command name and flag names. The error boolean is collected later.
func (u *Usage) Collect(cmd *cobra.Command, _ []string) {
	u.Version = cmd.Version
	u.Command = cmd.CommandPath()

	u.Flags = []string{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Changed {
			u.Flags = append(u.Flags, flag.Name)
		}
	})
}

// Report sends usage data to cc-cli-usage-service.
func (u *Usage) Report() {
	if u.Command == "" {
		return
	}

	log.CliLogger.Tracef("Reporting CLI usage: %+v", u)

	cfg := cliv1.NewConfiguration()
	cfg.Servers = cliv1.ServerConfigurations{
		{
			URL:         "https://api.devel.cpdev.cloud",
			Description: "Confluent Cloud development",
		},
	}

	api := cliv1.NewAPIClient(cfg).UsagesCliV1Api
	ctx := context.Background()
	req := api.CreateCliV1Usage(ctx)

	if _, err := api.CreateCliV1UsageExecute(req); err != nil {
		log.CliLogger.Tracef("Failed to report CLI usage: %v", err)
	}
}
