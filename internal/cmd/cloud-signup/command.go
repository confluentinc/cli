package cloudsignup

import (
	"context"
	"fmt"
	"os"
	"strings"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/admin"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/countrycode"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
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
	cmd.RunE = c.cloudSignupRunE

	cmd.Flags().String("url", "https://confluent.cloud", "Confluent Cloud service URL.")

	return cmd
}

func (c *command) cloudSignupRunE(cmd *cobra.Command, _ []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	client := ccloudv1.NewClient(&ccloudv1.Params{
		BaseURL:    url,
		UserAgent:  c.userAgent,
		HttpClient: ccloudv1.BaseClient,
		Logger:     log.CliLogger,
	})

	return c.signup(cmd, form.NewPrompt(os.Stdin), client)
}

func (c *command) signup(cmd *cobra.Command, prompt form.Prompt, client *ccloudv1.Client) error {
	utils.Println(cmd, "Sign up for Confluent Cloud. Use Ctrl-C to quit at any time.")
	fNameCompanyEmail := form.New(
		form.Field{ID: "name", Prompt: "Full name"},
		form.Field{ID: "company", Prompt: "Company"},
		form.Field{ID: "email", Prompt: "Company email", Regex: "^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"},
	)

	fCountry := form.New(
		form.Field{ID: "country", Prompt: "Two-letter country code (https://github.com/confluentinc/cli/blob/main/internal/pkg/countrycode/codes.go)", Regex: "^[a-zA-Z]{2}$"},
	)

	fPasswordTosPrivacy := form.New(
		form.Field{ID: "password", Prompt: "Password (must contain at least 8 characters, including 1 lowercase character, 1 uppercase character, and 1 number)", IsHidden: true},
		form.Field{ID: "tos", Prompt: "I have read and agree to the Terms of Service (https://www.confluent.io/confluent-cloud-tos/)", IsYesOrNo: true, RequireYes: true},
		form.Field{ID: "privacy", Prompt: `By entering "y" to submit this form, you agree that your personal data will be processed in accordance with our Privacy Policy (https://www.confluent.io/confluent-privacy-statement/)`, IsYesOrNo: true, RequireYes: true},
	)

	if err := fNameCompanyEmail.Prompt(cmd, prompt); err != nil {
		return err
	}

	var countryCode string

	for {
		if err := fCountry.Prompt(cmd, prompt); err != nil {
			return err
		}
		countryCode = strings.ToUpper(fCountry.Responses["country"].(string))
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

	if err := fPasswordTosPrivacy.Prompt(cmd, prompt); err != nil {
		return err
	}

	req := &ccloudv1.SignupRequest{
		Organization: &ccloudv1.Organization{
			Name: fNameCompanyEmail.Responses["company"].(string),
			Plan: &ccloudv1.Plan{AcceptTos: &types.BoolValue{Value: fPasswordTosPrivacy.Responses["tos"].(bool)}},
		},
		User: &ccloudv1.User{
			Email:     fNameCompanyEmail.Responses["email"].(string),
			FirstName: fNameCompanyEmail.Responses["name"].(string),
		},
		Credentials: &ccloudv1.Credentials{
			Password: fPasswordTosPrivacy.Responses["password"].(string),
		},
		CountryCode: countryCode,
	}

	signupReply, err := client.Signup.Create(context.Background(), req)
	if err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			return errors.NewErrorWithSuggestions("failed to sign up", "Please check if a verification link has been sent to your inbox, otherwise contact support at support@confluent.io")
		}
		return err

	}
	org := signupReply.GetOrganization()

	utils.Printf(cmd, "A verification email has been sent to %s.\n", fNameCompanyEmail.Responses["email"].(string))
	v := form.New(form.Field{ID: "verified", Prompt: `Type "y" once verified, or type "n" to resend.`, IsYesOrNo: true})

	for {
		if err := v.Prompt(cmd, prompt); err != nil {
			return err
		}

		if !v.Responses["verified"].(bool) {
			if err := client.Signup.SendVerificationEmail(context.Background(), &ccloudv1.User{Email: fNameCompanyEmail.Responses["email"].(string)}); err != nil {
				return err
			}

			utils.Printf(cmd, "A new verification email has been sent to %s. If this email is not received, please contact support@confluent.io.\n", fNameCompanyEmail.Responses["email"].(string))
			continue
		}

		req := &ccloudv1.AuthenticateRequest{
			Email:         fNameCompanyEmail.Responses["email"].(string),
			Password:      fPasswordTosPrivacy.Responses["password"].(string),
			OrgResourceId: org.ResourceId,
		}

		res, err := client.Auth.Login(context.Background(), req)
		if err != nil {
			if err.Error() == "username or password is invalid" {
				utils.ErrPrintln(cmd, "Sorry, your email is not verified. Another verification email was sent to your address. Please click the verification link in that message to verify your email.")
				continue
			}
			return err
		}

		utils.Print(cmd, errors.CloudSignUpMsg)

		authorizedClient := c.clientFactory.JwtHTTPClientFactory(context.Background(), res.Token, client.BaseURL)
		credentials := &pauth.Credentials{
			Username:         fNameCompanyEmail.Responses["email"].(string),
			AuthToken:        res.GetToken(),
			AuthRefreshToken: res.GetRefreshToken(),
		}
		_, currentOrg, err := pauth.PersistCCloudCredentialsToConfig(c.Config.Config, authorizedClient, client.BaseURL, credentials, false)
		if err != nil {
			utils.Println(cmd, "Failed to persist login to local config. Run `confluent login` to log in using the new credentials.")
			return nil
		}

		c.printFreeTrialAnnouncement(cmd, authorizedClient, currentOrg)

		utils.Printf(cmd, errors.LoggedInAsMsgWithOrg, fNameCompanyEmail.Responses["email"].(string), currentOrg.ResourceId, currentOrg.Name)
		return nil
	}
}

func (c *command) printFreeTrialAnnouncement(cmd *cobra.Command, client *ccloudv1.Client, currentOrg *ccloudv1.Organization) {
	promoCodeClaims, err := client.Growth.GetFreeTrialInfo(context.Background(), currentOrg.Id)
	if err != nil {
		log.CliLogger.Warnf("Failed to get free trial info: %v", err)
		return
	}

	// try to find free trial promo code
	hasFreeTrialCode := false
	freeTrialPromoCodeAmount := int64(0)
	for _, promoCodeClaim := range promoCodeClaims {
		if promoCodeClaim.GetIsFreeTrialPromoCode() {
			hasFreeTrialCode = true
			freeTrialPromoCodeAmount = promoCodeClaim.GetAmount()
			break
		}
	}

	if hasFreeTrialCode {
		utils.ErrPrintf(cmd, errors.FreeTrialSignUpMsg, admin.ConvertToUSD(freeTrialPromoCodeAmount))
	}
}
