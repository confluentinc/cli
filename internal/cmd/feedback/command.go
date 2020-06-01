package feedback

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewFeedbackCmd(prerunner pcmd.PreRunner, cfg *v3.Config, analytics analytics.Client) *cobra.Command {
	prompt := pcmd.NewPrompt(os.Stdin)
	return NewFeedbackCmdWithPrompt(prerunner, cfg, analytics, prompt)
}

func NewFeedbackCmdWithPrompt(prerunner pcmd.PreRunner, cfg *v3.Config, analyticsClient analytics.Client, prompt pcmd.Prompt) *cobra.Command {
	cmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "feedback",
			Short: "Submit feedback about the " + cfg.CLIName + " CLI.",
			RunE: func(cmd *cobra.Command, _ []string) error {
				pcmd.Print(cmd, "Enter feedback: ")
				msg, err := prompt.ReadString('\n')
				if err != nil {
					return err
				}
				msg = strings.TrimRight(msg, "\n")

				if len(msg) > 0 {
					analyticsClient.SetSpecialProperty(analytics.FeedbackPropertiesKey, msg)
					pcmd.Println(cmd, "Thanks for your feedback.")
				}
				return nil
			},
			Args: cobra.NoArgs,
		},
		cfg, prerunner)

	return cmd.Command
}
