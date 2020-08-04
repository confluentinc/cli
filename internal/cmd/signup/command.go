package signup

import (
	"context"
	"os"

	v1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type command struct {
	*pcmd.CLICommand
	logger    *log.Logger
	userAgent string
}

func New(prerunner pcmd.PreRunner, logger *log.Logger, userAgent string) *cobra.Command {
	signup := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "signup",
			Short: "Sign up for Confluent Cloud.",
			Args:  cobra.NoArgs,
		},
		prerunner,
	)

	c := &command{signup, logger, userAgent}
	c.RunE = pcmd.NewCLIRunE(c.signupRunE)
	c.Flags().String("url", "https://confluent.cloud", "Confluent Cloud service URL.")
	c.Flags().SortFlags = false

	return c.Command
}

func (c *command) signupRunE(cmd *cobra.Command, _ []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	client := ccloud.NewClient(&ccloud.Params{
		BaseURL:    url,
		UserAgent:  c.userAgent,
		HttpClient: ccloud.BaseClient,
		Logger:     c.logger,
	})

	return c.signup(cmd, pcmd.NewPrompt(os.Stdin), client)
}

func (c *command) signup(cmd *cobra.Command, prompt pcmd.Prompt, client *ccloud.Client) error {
	f := form.New(
		form.Field{ID: "email", Prompt: "Email"},
		form.Field{ID: "first", Prompt: "First Name"},
		form.Field{ID: "last", Prompt: "Last Name"},
		form.Field{ID: "organization", Prompt: "Organization"},
		form.Field{ID: "password", Prompt: "Password", IsHidden: true},
		form.Field{ID: "tos", Prompt: "I have read and agree to the Terms of Service (https://www.confluent.io/confluent-cloud-tos/)", IsYesOrNo: true},
		// Marketing
		form.Field{ID: "privacy", Prompt: `By typing "y", you agree that your personal data will be processed in accordance with our Privacy Policy (https://www.confluent.io/confluent-privacy-statement/)`, IsYesOrNo: true},
	)

	pcmd.Println(cmd, "Sign up for Confluent Cloud. Use Ctrl+C to quit at any time.")
	if err := f.Prompt(cmd, prompt); err != nil {
		return err
	}

	if !f.Responses["tos"].(bool) {
		pcmd.Println(cmd, "You must accept the Terms of Service.")
		return nil
	}

	if !f.Responses["privacy"].(bool) {
		pcmd.Println(cmd, "You must accept the Privacy Policy.")
		return nil
	}

	req := &v1.SignupRequest{
		Organization: &v1.Organization{
			Name: f.Responses["organization"].(string),
		},
		User: &v1.User{
			Email:     f.Responses["email"].(string),
			FirstName: f.Responses["first"].(string),
			LastName:  f.Responses["last"].(string),
		},
		Credentials: &v1.Credentials{
			Password: f.Responses["password"].(string),
		},
	}

	if _, err := client.Signup.Create(context.Background(), req); err != nil {
		return err
	}

	pcmd.Println(cmd, "Success! Welcome to Confluent Cloud.")
	return nil
}
