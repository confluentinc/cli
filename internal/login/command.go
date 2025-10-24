package login

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/billing"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/keychain"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type command struct {
	*pcmd.CLICommand
	cfg                      *config.Config
	ccloudClientFactory      pauth.CCloudClientFactory
	mdsClientManager         pauth.MDSClientManager
	loginCredentialsManager  pauth.LoginCredentialsManager
	loginOrganizationManager pauth.LoginOrganizationManager
	authTokenHandler         pauth.AuthTokenHandler
}

func New(cfg *config.Config, prerunner pcmd.PreRunner, ccloudClientFactory pauth.CCloudClientFactory, mdsClientManager pauth.MDSClientManager, loginCredentialsManager pauth.LoginCredentialsManager, loginOrganizationManager pauth.LoginOrganizationManager, authTokenHandler pauth.AuthTokenHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Confluent Cloud or Confluent Platform.",
		Long: fmt.Sprintf("Confluent Cloud:\n\nLog in to Confluent Cloud using your email and password, or using single sign-on (SSO) credentials.\n\nEmail and password login can be accomplished non-interactively using the `%s` and `%s` environment variables.\n\nEmail and password can also be stored locally for non-interactive re-authentication with the `--save` flag.\n\nSSO login can be accomplished headlessly using the `--no-browser` flag, but non-interactive login is not natively supported. Authentication tokens last 8 hours and are automatically refreshed with CLI client usage. If the client is not used for more than 8 hours, you have to log in again.\n\nLog in to a specific Confluent Cloud organization using the `--organization` flag, or by setting the environment variable `%s`.\n\n", pauth.ConfluentCloudEmail, pauth.ConfluentCloudPassword, pauth.ConfluentCloudOrganizationId) +
			fmt.Sprintf("Confluent Platform:\n\nLog in to Confluent Platform with your username and password, the `--url` flag to identify the location of your Metadata Service (MDS), and the `--certificate-authority-path` flag to identify your self-signed certificate chain.\n\nLogin can be accomplished non-interactively using the `%s`, `%s`, `%s`, and `%s` environment variables.\n\nIn a non-interactive login, `%s` replaces the `--url` flag, and `%s` replaces the `--certificate-authority-path` flag.\n\nEven with the environment variables set, you can force an interactive login using the `--prompt` flag.", pauth.ConfluentPlatformUsername, pauth.ConfluentPlatformPassword, pauth.ConfluentPlatformMDSURL, pauth.ConfluentPlatformCertificateAuthorityPath, pauth.ConfluentPlatformMDSURL, pauth.ConfluentPlatformCertificateAuthorityPath),
		Args: cobra.NoArgs,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Log in to Confluent Cloud.",
				Code: "confluent login",
			},
			examples.Example{
				Text: "Log in to a specific organization in Confluent Cloud.",
				Code: "confluent login --organization 00000000-0000-0000-0000-000000000000",
			},
			examples.Example{
				Text: "Log in to Confluent Platform with a MDS URL.",
				Code: "confluent login --url http://localhost:8090",
			},
			examples.Example{
				Text: "Log in to Confluent Platform with a MDS URL and Certification Authority certificate.",
				Code: "confluent login --url https://localhost:8090 --certificate-authority-path certs/my-cert.crt",
			},
			examples.Example{
				Text: "Log in to Confluent Platform with SSO even if `CONFLUENT_PLATFORM_USERNAME` and `CONFLUENT_PLATFORM_PASSWORD` are set.",
				Code: "CONFLUENT_PLATFORM_SSO=true confluent login --url https://localhost:8090 --certificate-authority-path certs/my-cert.crt",
			},
		),
	}

	c := &command{
		CLICommand:               pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:                      cfg,
		mdsClientManager:         mdsClientManager,
		ccloudClientFactory:      ccloudClientFactory,
		loginCredentialsManager:  loginCredentialsManager,
		loginOrganizationManager: loginOrganizationManager,
		authTokenHandler:         authTokenHandler,
	}
	cmd.RunE = c.login

	cmd.Flags().String("url", "", "Metadata Service (MDS) URL, for on-premises deployments.")
	cmd.Flags().Bool("us-gov", false, "Log in to the Confluent Cloud US Gov environment.")
	cmd.Flags().String("certificate-authority-path", "", "Self-signed certificate chain in PEM format, for on-premises deployments.")
	pcmd.AddMDSOnPremMTLSFlags(cmd)
	cmd.Flags().Bool("certificate-only", false, "Authenticate using mTLS certificate and key without SSO or username/password.")
	cmd.Flags().Bool("no-browser", false, "Do not open a browser window when authenticating using Single Sign-On (SSO).")
	cmd.Flags().String("organization", "", "The Confluent Cloud organization to log in to. If empty, log in to the default organization.")
	cmd.Flags().Bool("prompt", false, "Bypass non-interactive login and prompt for login credentials.")
	cmd.Flags().Bool("save", false, "Save username and encrypted password (non-SSO credentials) to the configuration file in your $HOME directory, and to macOS keychain if applicable. You will be logged back in when your token expires, after one hour for Confluent Cloud, or after six hours for Confluent Platform.")

	cobra.CheckErr(cmd.Flags().MarkHidden("us-gov"))

	cmd.MarkFlagsMutuallyExclusive("url", "us-gov")

	return cmd
}

func (c *command) login(cmd *cobra.Command, _ []string) error {
	url, err := c.getURL(cmd)
	if err != nil {
		return err
	}

	isCCloud := ccloudv2.IsCCloudURL(url, c.cfg.IsTest)

	url, warningMsg, err := validateURL(url, isCCloud)
	if err != nil {
		return err
	}
	if warningMsg != "" {
		output.ErrPrintf(c.Config.EnableColor, "Assuming %s.\n", warningMsg)
	}

	if isCCloud {
		return c.loginCCloud(cmd, url)
	} else {
		return c.loginMDS(cmd, url)
	}
}

func (c *command) loginCCloud(cmd *cobra.Command, url string) error {
	organization := c.getOrganizationId(cmd)

	noBrowser, err := cmd.Flags().GetBool("no-browser")
	if err != nil {
		return err
	}

	credentials, err := c.getCCloudCredentials(cmd, url, organization)
	if err != nil {
		return err
	}

	token, refreshToken, err := c.authTokenHandler.GetCCloudTokens(c.ccloudClientFactory, url, credentials, noBrowser, organization)

	endOfFreeTrialErr, isEndOfFreeTrialErr := err.(*errors.EndOfFreeTrialError)

	// for orgs that are suspended because it reached end of free trial due to paywall removal, they should still be
	// able to log in.
	if err != nil && !isEndOfFreeTrialErr {
		// for orgs that are suspended due to other reason, they shouldn't be able to log in and should return error immediately.
		if err, ok := err.(*ccloudv1.SuspendedOrganizationError); ok {
			return errors.NewErrorWithSuggestions(err.Error(), errors.SuspendedOrganizationSuggestions)
		}
		return err
	}

	client := c.ccloudClientFactory.JwtHTTPClientFactory(context.Background(), token, url)
	credentials.AuthToken = token
	credentials.AuthRefreshToken = refreshToken

	save, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}

	currentEnvironment, currentOrg, err := pauth.PersistCCloudCredentialsToConfig(c.Config, client, url, credentials, save)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.LoggedInAsMsgWithOrg, credentials.Username, currentOrg.GetResourceId(), currentOrg.GetName())
	if currentEnvironment != "" {
		log.CliLogger.Debugf(errors.LoggedInUsingEnvMsg, currentEnvironment)
	}

	// if org is at the end of free trial, print instruction about how to add payment method to unsuspend the org.
	// otherwise, print remaining free credit upon each login.
	if isEndOfFreeTrialErr {
		// only print error and do not return it, since end-of-free-trial users should still be able to log in.
		output.ErrPrintf(c.Config.EnableColor, "Error: %s", endOfFreeTrialErr.Error())
		output.ErrPrint(c.Config.EnableColor, errors.DisplaySuggestionsMessage(endOfFreeTrialErr.UserFacingError()))
	} else if !c.cfg.HasGovHostname() {
		c.printRemainingFreeCredit(client, currentOrg)
	}

	if save && runtime.GOOS == "darwin" && !c.cfg.IsTest {
		return c.saveLoginToKeychain(true, url, credentials)
	}
	return nil
}

func (c *command) printRemainingFreeCredit(client *ccloudv1.Client, currentOrg *ccloudv1.Organization) {
	promoCodeClaims, err := client.Growth.GetFreeTrialInfo(currentOrg.GetId())
	if err != nil {
		log.CliLogger.Warnf("Failed to get free trial info: %v", err)
		return
	}

	// the org is not on free trial or there is no promo code claims
	if len(promoCodeClaims) == 0 {
		log.CliLogger.Debugf("Skip printing remaining free credit")
		return
	}

	// aggregate remaining free credit
	remainingFreeCredit := int64(0)
	for _, promoCodeClaim := range promoCodeClaims {
		remainingFreeCredit += promoCodeClaim.GetBalance()
	}

	// only print remaining free credit if there is any unexpired promo code and there is no payment method yet
	if remainingFreeCredit > 0 {
		output.ErrPrintf(c.Config.EnableColor, "Free credits: $%.2f USD remaining\n", billing.ConvertToUSD(remainingFreeCredit))
		output.ErrPrintln(c.Config.EnableColor, "You are currently using a free trial version of Confluent Cloud. Add a payment method with `confluent billing payment update` to avoid an interruption in service once your trial ends.")
	}
}

// Order of precedence: environment variables > configuration file / macOS keychain > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (c *command) getCCloudCredentials(cmd *cobra.Command, url, organization string) (*pauth.Credentials, error) {
	client := c.ccloudClientFactory.AnonHTTPClientFactory(url)
	c.loginCredentialsManager.SetCloudClient(client)

	prompt, err := cmd.Flags().GetBool("prompt")
	if err != nil {
		return nil, err
	}
	if prompt {
		return pauth.GetLoginCredentials(c.loginCredentialsManager.GetCloudCredentialsFromPrompt(organization))
	}

	return pauth.GetLoginCredentials(
		c.loginCredentialsManager.GetCloudCredentialsFromEnvVar(organization),
		c.loginCredentialsManager.GetCloudCredentialsFromPrompt(organization),
	)
}

func (c *command) loginMDS(cmd *cobra.Command, url string) error {
	caCertPath, err := c.getCaCertPath(cmd)
	if err != nil {
		return err
	}

	clientCertPath, clientKeyPath, err := pcmd.GetClientCertAndKeyPaths(cmd)
	if err != nil {
		return err
	}

	credentials, err := c.getConfluentCredentials(cmd, url, caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		return err
	}

	unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
	if err != nil {
		return err
	}

	client, err := c.mdsClientManager.GetMDSClient(url, caCertPath, clientCertPath, clientKeyPath, unsafeTrace)
	if err != nil {
		return err
	}

	noBrowser, err := cmd.Flags().GetBool("no-browser")
	if err != nil {
		return err
	}

	token, refreshToken, err := c.authTokenHandler.GetConfluentToken(client, credentials, noBrowser)
	if err != nil {
		return err
	}

	save, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}

	if err := pauth.PersistConfluentLoginToConfig(c.Config, credentials, url, token, refreshToken, caCertPath, save); err != nil {
		return err
	}

	if save && runtime.GOOS == "darwin" && !c.cfg.IsTest {
		if err := c.saveLoginToKeychain(false, url, credentials); err != nil {
			return err
		}
	}

	log.CliLogger.Debugf(errors.LoggedInAsMsg, credentials.Username)
	return nil
}

func (c *command) getCaCertPath(cmd *cobra.Command) (string, error) {
	certificateAuthorityPath, err := cmd.Flags().GetString("certificate-authority-path")
	if err != nil {
		return "", err
	}

	if certificateAuthorityPath == "" {
		certificateAuthorityPath = os.Getenv(pauth.ConfluentPlatformCertificateAuthorityPath)
	}

	if certificateAuthorityPath != "" {
		certificateAuthorityPath, err = filepath.Abs(certificateAuthorityPath)
		if err != nil {
			return "", err
		}
	}

	return certificateAuthorityPath, nil
}

// Order of precedence: prompt flag > environment variables (SSO > LDAP) > LDAP (keychain > config) > SSO > LDAP (prompt)
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
// SSO and LDAP (basic auth) can be enabled simultaneously
func (c *command) getConfluentCredentials(cmd *cobra.Command, url, caCertPath, clientCertPath, clientKeyPath string) (*pauth.Credentials, error) {
	prompt, err := cmd.Flags().GetBool("prompt")
	if err != nil {
		return nil, err
	}
	if prompt {
		return pauth.GetLoginCredentials(c.loginCredentialsManager.GetOnPremCredentialsFromPrompt())
	}

	unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
	if err != nil {
		return nil, err
	}

	certificateOnly, err := cmd.Flags().GetBool("certificate-only")
	if err != nil {
		return nil, err
	}

	if certificateOnly {
		return pauth.GetLoginCredentials(
			c.loginCredentialsManager.GetOnPremCertOnlyCredentials(certificateOnly),
			c.loginCredentialsManager.GetOnPremCredentialsFromPrompt(),
		)
	}

	if pauth.IsOnPremSSOEnv() {
		return pauth.GetLoginCredentials(
			c.loginCredentialsManager.GetOnPremSsoCredentials(url, caCertPath, clientCertPath, clientKeyPath, unsafeTrace),
			c.loginCredentialsManager.GetOnPremCredentialsFromPrompt(),
		)
	}

	return pauth.GetLoginCredentials(
		c.loginCredentialsManager.GetOnPremCredentialsFromEnvVar(),
		c.loginCredentialsManager.GetOnPremSsoCredentials(url, caCertPath, clientCertPath, clientKeyPath, unsafeTrace),
		c.loginCredentialsManager.GetOnPremCredentialsFromPrompt(),
	)
}

func (c *command) getURL(cmd *cobra.Command) (string, error) {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return "", err
	}
	if url != "" {
		return url, nil
	}

	usGov, err := cmd.Flags().GetBool("us-gov")
	if err != nil {
		return "", err
	}
	if usGov {
		return "https://confluentgov.com", nil
	}

	if url := os.Getenv(pauth.ConfluentPlatformMDSURL); url != "" {
		return url, nil
	}

	return pauth.CCloudURL, nil
}

func (c *command) saveLoginToKeychain(isCloud bool, url string, credentials *pauth.Credentials) error {
	if credentials.IsSSO || credentials.IsMFA {
		output.ErrPrintln(c.cfg.EnableColor, "The `--save` flag was ignored since SSO or MFA credentials are not stored on keychain.")
		return nil
	}

	ctxName := c.Config.Context().GetMachineName()
	if err := keychain.Write(isCloud, ctxName, url, credentials.Username, credentials.Password); err != nil {
		return err
	}

	output.ErrPrintln(c.cfg.EnableColor, "Wrote login credentials to keychain.")

	return nil
}

func validateURL(url string, isCCloud bool) (string, string, error) {
	if isCCloud {
		if strings.Contains(url, ccloudv2.Hostnames[0]) {
			if !strings.HasSuffix(strings.TrimSuffix(url, "/"), ccloudv2.Hostnames[0]) {
				return url, "", errors.NewErrorWithSuggestions("there is no need to pass the `--url` flag if you are logging in to Confluent Cloud", "Log in to Confluent Cloud with `confluent login`.")
			}
		}
	}

	var msg []string
	if !regexp.MustCompile(`(\w+)://`).MatchString(url) {
		url = "https://" + url
		msg = append(msg, "https protocol")
	}
	if !isCCloud && !regexp.MustCompile(`:(\d+\/?)`).MatchString(url) {
		url += ":8090"
		msg = append(msg, "default MDS port 8090")
	}

	var pattern *regexp.Regexp
	if isCCloud {
		pattern = regexp.MustCompile(`^\w+://[^/ ]+`)
	} else {
		pattern = regexp.MustCompile(`^\w+://[^/ ]+:\d+(?:\/|$)`)
	}
	if !pattern.MatchString(url) {
		return "", "", fmt.Errorf(errors.InvalidLoginURLErrorMsg)
	}

	return url, strings.Join(msg, " and "), nil
}

func (c *command) getOrganizationId(cmd *cobra.Command) string {
	return pauth.GetLoginOrganization(
		c.loginOrganizationManager.GetLoginOrganizationFromFlag(cmd),
		c.loginOrganizationManager.GetLoginOrganizationFromEnvironmentVariable(),
	)
}
