package cloudsignup

import (
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud-signup",
		Short: "Sign up for Confluent Cloud.",
		Args:  cobra.NoArgs,
		RunE:  cloudSignup,
	}

	return cmd
}

func cloudSignup(cmd *cobra.Command, _ []string) error {
	signupUrl := "https://www.confluent.io/get-started/"

	output.Printf("You will now be redirected to the Confluent Cloud sign up page in your browser. If you are not redirected, please use the following link: %s\n", signupUrl)
	if err := form.ConfirmEnter(); err != nil {
		return err
	}

	if err := browser.OpenURL(signupUrl); err != nil {
		return errors.Wrap(err, "unable to open web browser for cloud signup")
	}

	return nil
}
