package signup

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"

	"github.com/gogo/protobuf/types"

	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.CLICommand
	logger        *log.Logger
	userAgent     string
	clientFactory pauth.CCloudClientFactory
}

type countries struct {
	countries []countries `json:"country_codes"`
}

type country struct {
	Name string `json:"`
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
	f_email_name := form.New(
		form.Field{ID: "email", Prompt: "Email", Regex: "^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"},
		form.Field{ID: "first", Prompt: "First Name"},
		form.Field{ID: "last", Prompt: "Last Name"},
	)

	f_countrycode := form.New(
		form.Field{ID: "country", Prompt: "Two-letter country code (https://www.nationsonline.org/oneworld/country_code_list.htm)"},
	)

	f_org_pswd_toc_pri := form.New(
		form.Field{ID: "organization", Prompt: "Organization"},
		form.Field{ID: "password", Prompt: "Password (8 characters, 1 lowercase, 1 uppercase, 1 number)", IsHidden: true},
		form.Field{ID: "tos", Prompt: "I have read and agree to the Terms of Service (https://www.confluent.io/confluent-cloud-tos/)", IsYesOrNo: true, RequireYes: true},
		form.Field{ID: "privacy", Prompt: `By entering "y" to submit this form, you agree that your personal data will be processed in accordance with our Privacy Policy (https://www.confluent.io/confluent-privacy-statement/)`, IsYesOrNo: true, RequireYes: true},
	)

	if err := f_email_name.Prompt(cmd, prompt); err != nil {
		return err
	}

	countries, err := getCountriesMap()
	if err != nil {
		return err
	}

	var confirmedCountryCode string

	for {
		if err := f_countrycode.Prompt(cmd, prompt); err != nil {
			return err
		}
		code := strings.ToUpper(f_countrycode.Responses["country"].(string))
		if country, ok := countries[code]; ok {
			f := form.New(
				form.Field{ID: "confirmation", Prompt: fmt.Sprintf("You entered %s for %s. Is that correct?", code, country), IsYesOrNo: true},
			)
			if err := f.Prompt(cmd, prompt); err != nil {
				return err
			}
			if f.Responses["confirmation"].(bool) {
				confirmedCountryCode = code
				break
			} else {
				continue // reprompt for country code
			}
		} else {
			utils.Println(cmd, "Country code not found.")
			continue // reprompt for country code
		}
	}

	if err := f_org_pswd_toc_pri.Prompt(cmd, prompt); err != nil {
		return err
	}

	req := &v1.SignupRequest{
		Organization: &v1.Organization{
			Name: f_org_pswd_toc_pri.Responses["organization"].(string),
			Plan: &v1.Plan{AcceptTos: &types.BoolValue{Value: f_org_pswd_toc_pri.Responses["tos"].(bool)}},
		},
		User: &v1.User{
			Email:     f_email_name.Responses["email"].(string),
			FirstName: f_email_name.Responses["first"].(string),
			LastName:  f_email_name.Responses["last"].(string),
		},
		Credentials: &v1.Credentials{
			Password: f_org_pswd_toc_pri.Responses["password"].(string),
		},
		CountryCode: confirmedCountryCode,
	}

	if _, err := client.Signup.Create(context.Background(), req); err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			return errors.NewErrorWithSuggestions("failed to signup", "Please check if a verification link has been sent to your inbox, otherwise contact support at support@confluent.io")
		}
		return err

	}

	utils.Printf(cmd, "A verification email has been sent to %s.\n", f_email_name.Responses["email"].(string))
	v := form.New(form.Field{ID: "verified", Prompt: `Type "y" once verified, or type "n" to resend.`, IsYesOrNo: true})

	for {
		if err := v.Prompt(cmd, prompt); err != nil {
			return err
		}

		if !v.Responses["verified"].(bool) {
			if err := client.Signup.SendVerificationEmail(context.Background(), &v1.User{Email: f_email_name.Responses["email"].(string)}); err != nil {
				return err
			}

			utils.Printf(cmd, "A new verification email has been sent to %s. If this email is not received, please contact support@confluent.io.\n", f_email_name.Responses["email"].(string))
			continue
		}
		var token string
		var err error
		if token, err = client.Auth.Login(context.Background(), "", f_email_name.Responses["email"].(string), f_org_pswd_toc_pri.Responses["password"].(string)); err != nil {
			if err.Error() == "username or password is invalid" {
				utils.ErrPrintln(cmd, "Sorry, your email is not verified. Another verification email was sent to your address. Please click the verification link in that message to verify your email.")
				continue
			}
			return err
		}

		utils.Println(cmd, "Success! Welcome to Confluent Cloud.")
		authorizedClient := c.clientFactory.JwtHTTPClientFactory(context.Background(), token, client.BaseURL)
		_, err = pauth.PersistCCloudLoginToConfig(c.Config.Config, f_email_name.Responses["email"].(string), client.BaseURL, token, authorizedClient)
		if err != nil {
			utils.Println(cmd, "Failed to persist login to local config. Run `ccloud login` to log in using the new credentials.")
			return nil
		}
		c.logger.Debugf(errors.LoggedInAsMsg, f_email_name.Responses["email"])
		return nil
	}
}

func getCountriesMap() (map[string]interface{}, error) {
	_, callerFileName, _, _ := runtime.Caller(0)
	countryCodeFilepath := filepath.Join(filepath.Dir(callerFileName), "country_code.json")
	countryCode, err := os.Open(countryCodeFilepath)
	if err != nil {
		return nil, err
	}
	defer countryCode.Close()

	byteValue, _ := ioutil.ReadAll(countryCode)
	var countries map[string]interface{}
	err = json.Unmarshal(byteValue, &countries)
	if err != nil {
		return nil, err
	}
	return countries, nil
}
