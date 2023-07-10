package feedback

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "feedback",
		Short:       fmt.Sprintf("Submit feedback for the %s.", pversion.FullCLIName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.feedback

	return cmd
}

func (c *command) feedback(_ *cobra.Command, _ []string) error {
	feedback, err := getFeedback(form.NewPrompt(os.Stdin))
	if err != nil {
		return err
	}
	if feedback != "" {
		feedbackReq := cliv1.CliV1Feedback{Content: cliv1.PtrString(feedback)}
		if err := c.V2Client.CreateCliFeedback(feedbackReq); err != nil {
			return err
		}
		output.Println("Thanks for your feedback.")
	}
	return nil
}

func getFeedback(prompt form.Prompt) (string, error) {
	output.Print("Enter feedback: ")
	feedback, err := prompt.ReadLine()
	if err != nil {
		return "", err
	}
	feedback = strings.TrimSpace(feedback)
	f := form.New(form.Field{
		ID:        "proceed",
		Prompt:    "Please confirm that your feedback does not contain any sensitive information",
		IsYesOrNo: true,
	})
	if err := f.Prompt(prompt); err != nil {
		return "", err
	}
	if !f.Responses["proceed"].(bool) {
		output.Println("Your feedback was not submitted.")
		return "", err
	}
	return feedback, err
}
