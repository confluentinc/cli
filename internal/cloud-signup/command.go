package cloudsignup

import (
	"fmt"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type command struct {
	*pcmd.CLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud-signup",
		Short: "Sign up for Confluent Cloud.",
		Args:  cobra.NoArgs,
	}

	c := &command{pcmd.NewAnonymousCLICommand(cmd, prerunner)}
	cmd.RunE = c.cloudSignup

	return cmd
}

func (c *command) cloudSignup(cmd *cobra.Command, _ []string) error {
	signupUrl := "https://www.confluent.io/get-started/"

	output.Printf(c.Config.EnableColor, "You will now be redirected to the Confluent Cloud sign up page in your browser. If you are not redirected, please use the following link: %s\n", signupUrl)
	if err := form.ConfirmEnter(); err != nil {
		return err
	}

	if err := browser.OpenURL(signupUrl); err != nil {
		return fmt.Errorf("unable to open web browser for cloud signup: %w", err)
	}

	return nil
}
