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
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
	}
	cmd.RunE = c.feedback

	return cmd
}

func (c *command) feedback(_ *cobra.Command, _ []string) error {
	feedback, err := getFeedback(form.NewPrompt(os.Stdin))
	if err != nil {
		return err
	}
	if len(feedback) > 0 {
		if err := c.sendFeedback(feedback); err != nil {
			return err
		}
		fmt.Println("Thanks for your feedback.")
	}
	return nil
}

func getFeedback(prompt form.Prompt) (string, error) {
	f := form.New(form.Field{
		ID:        "proceed",
		Prompt:    "Please confirm that your feedback does not contain any sensitive information",
		IsYesOrNo: true,
	})
	if err := f.Prompt(prompt); err != nil || !f.Responses["proceed"].(bool) {
		return "", err
	}
	output.Print("Enter feedback: ")
	feedback, err := prompt.ReadLine()
	if err != nil {
		return "", err
	}
	feedback = strings.TrimLeft(feedback, " ")
	feedback = strings.TrimRight(feedback, " \n")
	return feedback, err
}

func (c *command) sendFeedback(content string) error {
	feedback := cliv1.CliV1Feedback{Content: cliv1.PtrString(content)}
	return c.V2Client.CreateCliFeedback(feedback)
}
