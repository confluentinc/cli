package feedback

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	form.Prompt
}

const (
	maxLength = 2000
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "feedback",
		Short:       fmt.Sprintf("Submit feedback about the %s.", pversion.FullCLIName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		Prompt:                  form.NewPrompt(os.Stdin),
	}
	cmd.RunE = c.feedback

	return cmd
}

func (c *command) feedback(_ *cobra.Command, _ []string) error {
	msg, err := getFeedback(c.Prompt)
	if err != nil {
		return err
	}
	if len(msg) > 0 {
		if err := c.sendFeedback(msg); err != nil {
			return err
		}
		fmt.Println("Thanks for your feedback.")
	}
	return nil
}

func getFeedback(prompt form.Prompt) (string, error) {
	f := form.New(form.Field{
		ID: "proceed",
		Prompt: "Please confirm that your feedback will not contain any " +
			"sensitive information such as API-keys, unencrypted tokens, etc.",
		IsYesOrNo: true,
	})
	if err := f.Prompt(prompt); err != nil || !f.Responses["proceed"].(bool) {
		output.Println("Not proceeding with feedback")
		return "", err
	}
	output.Println("Enter feedback (maximum length of 2,000 characters): ")
	msg, err := prompt.ReadLine()
	if err != nil {
		return "", err
	}
	msg = strings.TrimLeft(msg, " ")
	msg = strings.TrimRight(msg, " \n")
	if len(msg) > maxLength {
		return "", errors.New(errors.ExceedsMaxLengthMsg)
	}
	return msg, err
}

func (c *command) sendFeedback(msg string) error {
	feedback := cliv1.CliV1Feedback{Content: cliv1.PtrString(msg)}
	return c.V2Client.CreateCliFeedback(feedback)
}
