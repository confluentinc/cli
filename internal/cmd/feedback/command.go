package feedback

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type command struct {
	prompt form.Prompt
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	prompt := form.NewPrompt(os.Stdin)
	return NewFeedbackCmdWithPrompt(prerunner, prompt)
}

func NewFeedbackCmdWithPrompt(prerunner pcmd.PreRunner, prompt form.Prompt) *cobra.Command {
	c := command{prompt: prompt}
	cmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "feedback",
			Short: fmt.Sprintf("Submit feedback about the %s.", version.FullCLIName),
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
		utils.Println(cmd, errors.ThanksForFeedbackMsg)
	}
	return nil
}
