package signup

import (
	"context"
	v1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"
	"os"
	"strings"

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
	c := &command{
		pcmd.NewAnonymousCLICommand(
			&cobra.Command{
				Use:   "signup",
				Short: "Sign up for Confluent Cloud.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
		logger,
		userAgent,
	}

	c.Flags().String("url", "https://confluent.cloud", "Confluent Cloud service URL.")
	c.Flags().SortFlags = false

	c.RunE = pcmd.NewCLIRunE(c.signupRunE)

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

	return signup(cmd, pcmd.NewPrompt(os.Stdin), client)
}

func signup(cmd *cobra.Command, prompt pcmd.Prompt, client *ccloud.Client) error {
	pcmd.Println(cmd, "Sign up for Confluent Cloud. Use Ctrl+C to quit at any time.")

	f := form.New(
		form.Field{ID: "email", Prompt: "Email", Regex: `(?:[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`},
		form.Field{ID: "first", Prompt: "First Name"},
		form.Field{ID: "last", Prompt: "Last Name"},
		form.Field{ID: "organization", Prompt: "Organization"},
		form.Field{ID: "password", Prompt: "Password", IsHidden: true},
		form.Field{ID: "tos", Prompt: "I have read and agree to the Terms of Service (https://www.confluent.io/confluent-cloud-tos/)", IsYesOrNo: true, RequireYes: true},
		form.Field{ID: "privacy", Prompt: `By entering "y" to submit this form, you agree that your personal data will be processed in accordance with our Privacy Policy (https://www.confluent.io/confluent-privacy-statement/)`, IsYesOrNo: true, RequireYes: true},
	)

	if err := f.Prompt(cmd, prompt); err != nil {
		return err
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
		if strings.Contains(err.Error(), "email: already exists") {
			if _, err := client.Auth.Login(context.Background(), "", f.Responses["email"].(string), f.Responses["password"].(string)); err != nil {
				pcmd.Println(cmd, "There is already an account associated with this email. If you are unable to login, please ensure your email is verified and your password is correct.")
				pcmd.Printf(cmd, "A new verification email has been sent to %s. If this email is not received, please contact support@confluent.io.\n", f.Responses["email"].(string))
				res := &v1.Credentials{
					Username: f.Responses["email"].(string),
				}
				if err := client.Signup.SendVerificationEmail(context.Background(), res); err != nil {
					return err
				}
				v := form.New(form.Field{ID: "verified", Prompt: `Type "y" once verified, or type "n" to exit flow.`, IsYesOrNo: true})
				if err := v.Prompt(cmd, prompt); err != nil {
					return err
				}
				if !v.Responses["verified"].(bool) {
					pcmd.Println(cmd, "Exiting.")
					return nil
				}
				if _, err := client.Auth.Login(context.Background(), "", f.Responses["email"].(string), f.Responses["password"].(string)); err != nil {
					pcmd.ErrPrintln(cmd, "Please ensure that you have verified the email. If you have, then your password is incorrect. Please try \"ccloud login\" using your correct credentials. For assistance, contact support@confluent.io")
					return nil
				}
			}
			pcmd.Println(cmd, "Welcome, you have been logged into an existing account.")
			return nil
		} else {
			return err
		}

	}

	pcmd.Printf(cmd, "A verification email has been sent to %s.\n", f.Responses["email"].(string))
	v := form.New(form.Field{ID: "verified", Prompt: `Type "y" once verified, or type "n" to resend.`, IsYesOrNo: true})

	for {
		if err := v.Prompt(cmd, prompt); err != nil {
			return err
		}

		if !v.Responses["verified"].(bool) {
			res := &v1.Credentials{
				Username: f.Responses["email"].(string),
			}
			if err := client.Signup.SendVerificationEmail(context.Background(), res); err != nil {
				return err
			}

			pcmd.Printf(cmd, "A new verification email has been sent to %s. If this email is not received, please contact support@confluent.io.\n", f.Responses["email"].(string))
			continue
		}

		if _, err := client.Auth.Login(context.Background(), "", f.Responses["email"].(string), f.Responses["password"].(string)); err != nil {
			if err.Error() == "username or password is invalid" {
				pcmd.ErrPrintln(cmd, "Sorry, your email is not verified. Another verification email was sent to your address. Please click the verification link in that message to verify your email.")
				continue
			}
			return err
		}

		pcmd.Println(cmd, "Success! Welcome to Confluent Cloud.")
		return nil
	}
}
