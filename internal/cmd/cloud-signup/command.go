package cloudsignup

import (
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.CLICommand
	userAgent     string
	clientFactory pauth.CCloudClientFactory
	isTest        bool
}

func New(prerunner pcmd.PreRunner, userAgent string, ccloudClientFactory pauth.CCloudClientFactory, isTest bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud-signup",
		Short: "Sign up for Confluent Cloud.",
		Args:  cobra.NoArgs,
	}

	c := &command{
		CLICommand:    pcmd.NewAnonymousCLICommand(cmd, prerunner),
		userAgent:     userAgent,
		clientFactory: ccloudClientFactory,
		isTest:        isTest,
	}
	cmd.RunE = c.signup

	return cmd
}

func (c *command) signup(cmd *cobra.Command, _ []string) error {
	utils.Println(cmd, "You will now be redirected to the Confluent Cloud sign up page in your browser.")
	err := form.ConfirmEnter(cmd)
	if err != nil {
		return err
	}

	err = browser.OpenURL("https://www.confluent.io/get-started/")
	if err != nil {
		return errors.Wrap(err, "unable to open web browser for cloud signup")
	}

	return nil
}
