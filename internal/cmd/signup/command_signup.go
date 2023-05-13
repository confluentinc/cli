package signup

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gogo/protobuf/types"

	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"

	"github.com/confluentinc/countrycode"
)

type command struct {
	*pcmd.CLICommand
	logger        *log.Logger
	userAgent     string
	clientFactory pauth.CCloudClientFactory
}

func New(prerunner pcmd.PreRunner, logger *log.Logger, userAgent string, ccloudClientFactory pauth.CCloudClientFactory) *command {
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
		ccloudClientFactory,
	}

	c.Flags().String("url", "https://confluent.cloud", "Confluent Cloud service URL.")
	c.Flags().SortFlags = false

	c.RunE = pcmd.NewCLIRunE(c.signupRunE)

	return c
}

func (c *command) Cmd() *cobra.Command {
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

	return c.Signup(cmd, form.NewPrompt(os.Stdin), client)
}

func (c *command) Signup(cmd *cobra.Command, prompt form.Prompt, client *ccloud.Client) error {
	utils.Println(cmd, "Sign up for Confluent Cloud. Use Ctrl+C to quit at any time.")
	fEmailName := form.New(
		form.Field{ID: "email", Prompt: "Email", Regex: "^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"},
		form.Field{ID: "first", Prompt: "First Name"},
		form.Field{ID: "last", Prompt: "Last Name"},
	)

	fCountrycode := form.New(
		form.Field{ID: "country", Prompt: "Two-letter country code (https://github.com/confluentinc/countrycode/blob/master/codes.go)", Regex: "^[a-zA-Z]{2}$"},
	)

	fOrgPswdTosPri := form.New(
		form.Field{ID: "organization", Prompt: "Organization"},
		form.Field{ID: "password", Prompt: "Password (must contain at least 8 characters, including 1 lowercase character, 1 uppercase character, and 1 number)", IsHidden: true},
		form.Field{ID: "tos", Prompt: "I have read and agree to the Terms of Service (https://www.confluent.io/confluent-cloud-tos/)", IsYesOrNo: true, RequireYes: true},
		form.Field{ID: "privacy", Prompt: `By entering "y" to submit this form, you agree that your personal data will be processed in accordance with our Privacy Policy (https://www.confluent.io/confluent-privacy-statement/)`, IsYesOrNo: true, RequireYes: true},
	)

	if err := fEmailName.Prompt(cmd, prompt); err != nil {
		return err
	}

	var countryCode string

	for {
		if err := fCountrycode.Prompt(cmd, prompt); err != nil {
			return err
		}
		countryCode = strings.ToUpper(fCountrycode.Responses["country"].(string))
		if country, ok := countrycode.Codes[countryCode]; ok {
			f := form.New(
				form.Field{ID: "confirmation", Prompt: fmt.Sprintf("You entered %s for %s. Is that correct?", countryCode, country), IsYesOrNo: true},
			)
			if err := f.Prompt(cmd, prompt); err != nil {
				return err
			}
			if f.Responses["confirmation"].(bool) {
				break
			}
		} else {
			utils.Println(cmd, "Country code not found.")
		}
	}

	if err := fOrgPswdTosPri.Prompt(cmd, prompt); err != nil {
		return err
	}

	req := &v1.SignupRequest{
		Organization: &v1.Organization{
			Name: fOrgPswdTosPri.Responses["organization"].(string),
			Plan: &v1.Plan{AcceptTos: &types.BoolValue{Value: fOrgPswdTosPri.Responses["tos"].(bool)}},
		},
		User: &v1.User{
			Email:     fEmailName.Responses["email"].(string),
			FirstName: fEmailName.Responses["first"].(string),
			LastName:  fEmailName.Responses["last"].(string),
		},
		Credentials: &v1.Credentials{
			Password: fOrgPswdTosPri.Responses["password"].(string),
		},
		CountryCode: countryCode,
	}

	if _, err := client.Signup.Create(context.Background(), req); err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			return errors.NewErrorWithSuggestions("failed to signup", "Please check if a verification link has been sent to your inbox, otherwise contact support at support@confluent.io")
		}
		return err

	}

	utils.Printf(cmd, "A verification email has been sent to %s.\n", fEmailName.Responses["email"].(string))
	v := form.New(form.Field{ID: "verified", Prompt: `Type "y" once verified, or type "n" to resend.`, IsYesOrNo: true})

	for {
		if err := v.Prompt(cmd, prompt); err != nil {
			return err
		}

		if !v.Responses["verified"].(bool) {
			if err := client.Signup.SendVerificationEmail(context.Background(), &v1.User{Email: fEmailName.Responses["email"].(string)}); err != nil {
				return err
			}

			utils.Printf(cmd, "A new verification email has been sent to %s. If this email is not received, please contact support@confluent.io.\n", fEmailName.Responses["email"].(string))
			continue
		}
		var token string
		var err error
		if token, err = client.Auth.Login(context.Background(), "", fEmailName.Responses["email"].(string), fOrgPswdTosPri.Responses["password"].(string), ""); err != nil {
			if err.Error() == "username or password is invalid" {
				utils.ErrPrintln(cmd, "Sorry, your email is not verified. Another verification email was sent to your address. Please click the verification link in that message to verify your email.")
				continue
			}
			return err
		}

		utils.Println(cmd, "Success! Welcome to Confluent Cloud.")
		authorizedClient := c.clientFactory.JwtHTTPClientFactory(context.Background(), token, client.BaseURL)
		credentials := &pauth.Credentials{
			Username:  fEmailName.Responses["email"].(string),
			AuthToken: token,
		}
		_, err = pauth.PersistCCloudLoginToConfig(c.Config.Config, credentials, client.BaseURL, token, authorizedClient, false)
		if err != nil {
			utils.Println(cmd, "Failed to persist login to local config. Run `ccloud login` to log in using the new credentials.")
			return nil
		}
		c.logger.Debugf(errors.LoggedInAsMsg, fEmailName.Responses["email"])
		return nil
	}
}
