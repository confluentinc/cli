package feedback

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type command struct {
	analyticsClient analytics.Client
	prompt          form.Prompt
}

func New(cliName string, prerunner pcmd.PreRunner, analytics analytics.Client) *cobra.Command {
	prompt := form.NewPrompt(os.Stdin)
	return NewFeedbackCmdWithPrompt(cliName, prerunner, analytics, prompt)
}

func NewFeedbackCmdWithPrompt(cliName string, prerunner pcmd.PreRunner, analyticsClient analytics.Client, prompt form.Prompt) *cobra.Command {
	c := command{
		analyticsClient: analyticsClient,
		prompt:          prompt,
	}
	cmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "feedback",
			Short: fmt.Sprintf("Submit feedback about the %s.", version.GetFullCLIName(cliName)),
			Args:  cobra.NoArgs,
			RunE:  pcmd.NewCLIRunE(c.feedbackRunE),
		}, prerunner)

	return cmd.Command
}

func (c *command) feedbackRunE(cmd *cobra.Command, _ []string) error {
	f := form.New(form.Field{ID: "feedback", Prompt: "Enter feedback"})
	if err := f.Prompt(cmd, c.prompt); err != nil {
		return err
	}
	msg := f.Responses["feedback"].(string)

	if len(msg) > 0 {
		c.analyticsClient.SetSpecialProperty(analytics.FeedbackPropertiesKey, msg)
		utils.Println(cmd, errors.ThanksForFeedbackMsg)
	}
	return nil
}
