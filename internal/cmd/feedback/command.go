package feedback

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/spf13/cobra"
)

func NewFeedbackCmd(prerunner pcmd.PreRunner, cfg *v3.Config) *cobra.Command {
	cmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "feedback",
			Short: "Submit feedback about the " + cfg.CLIName + " CLI",
			Run: func(cmd *cobra.Command, _ []string) {
				pcmd.Println(cmd, "Thanks for your feedback.")
			},
			Args: cobra.ExactArgs(1),
		},
		cfg, prerunner)
	return cmd.Command
}
