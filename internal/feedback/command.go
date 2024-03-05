package feedback

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	"github.com/confluentinc/go-editor"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
	pversion "github.com/confluentinc/cli/v3/pkg/version"
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

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.feedback

	return cmd
}

func (c *command) feedback(_ *cobra.Command, _ []string) error {
	file, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if err := editor.NewEditor().Launch(file.Name()); err != nil {
		return err
	}

	feedback, err := os.ReadFile(file.Name())
	if err != nil {
		return err
	}

	if len(bytes.Trim(feedback, " \t\n")) == 0 {
		output.ErrPrintln(c.Config.EnableColor, "Empty feedback not submitted.")
		return nil
	}

	ok, err := shouldProceed()
	if err != nil {
		return err
	}
	if !ok {
		output.Println(false, "Your feedback was not submitted.")
		return nil
	}

	feedbackReq := cliv1.CliV1Feedback{Content: cliv1.PtrString(string(feedback))}
	if err := c.V2Client.CreateCliFeedback(feedbackReq); err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, "Thanks for your feedback.")
	return nil
}

func shouldProceed() (bool, error) {
	f := form.New(form.Field{
		ID:        "proceed",
		Prompt:    "Please confirm that your feedback does not contain any sensitive information",
		IsYesOrNo: true,
	})

	if err := f.Prompt(form.NewPrompt()); err != nil {
		return false, err
	}

	return f.Responses["proceed"].(bool), nil
}
