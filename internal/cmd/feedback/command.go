package feedback

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(cliName string, prerunner pcmd.PreRunner, analytics analytics.Client) *cobra.Command {
	prompt := pcmd.NewPrompt(os.Stdin)
	return NewFeedbackCmdWithPrompt(cliName, prerunner, analytics, prompt)
}

func NewFeedbackCmdWithPrompt(cliName string, prerunner pcmd.PreRunner, analyticsClient analytics.Client, prompt pcmd.Prompt) *cobra.Command {
	cmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "feedback",
			Short: "Submit feedback about the " + cliName + " CLI.",
			RunE: func(cmd *cobra.Command, _ []string) error {
				pcmd.Print(cmd, "Enter feedback: ")
				msg, err := prompt.ReadString('\n')
				if err != nil {
					return err
				}
				msg = strings.TrimSuffix(msg, "\n")

				if len(msg) > 0 {
					analyticsClient.SetSpecialProperty(analytics.FeedbackPropertiesKey, msg)
					pcmd.Println(cmd, "Thanks for your feedback.")
				}
				return nil
			},
			Args: cobra.NoArgs,
		}, prerunner)

	return cmd.Command
}
