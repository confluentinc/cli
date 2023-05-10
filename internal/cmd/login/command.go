package login

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keychain"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
	testserver "github.com/confluentinc/cli/test/test-server"
)

type Command struct {
	*pcmd.CLICommand
	cfg                      *v1.Config
	ccloudClientFactory      pauth.CCloudClientFactory
	mdsClientManager         pauth.MDSClientManager
	netrcHandler             netrc.NetrcHandler
	loginCredentialsManager  pauth.LoginCredentialsManager
	loginOrganizationManager pauth.LoginOrganizationManager
	authTokenHandler         pauth.AuthTokenHandler
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner, ccloudClientFactory pauth.CCloudClientFactory, mdsClientManager pauth.MDSClientManager, netrcHandler netrc.NetrcHandler, loginCredentialsManager pauth.LoginCredentialsManager, authTokenHandler pauth.AuthTokenHandler) *Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Confluent Cloud or Confluent Platform.",
		Long: fmt.Sprintf("Log in to Confluent Cloud using your email and password, or non-interactively using the `%s` and `%s` environment variables.\n\n", pauth.ConfluentCloudEmail, pauth.ConfluentCloudPassword) +
			fmt.Sprintf("You can log in to a specific Confluent Cloud organization using the `--organization-id` flag, or by setting the environment variable `%s`.\n\n", pauth.ConfluentCloudOrganizationId) +
			fmt.Sprintf("You can log in to Confluent Platform with your username and password, or non-interactively using `%s`, `%s`, `%s`, and `%s`.", pauth.ConfluentPlatformUsername, pauth.ConfluentPlatformPassword, pauth.ConfluentPlatformMDSURL, pauth.ConfluentPlatformCACertPath) +
			fmt.Sprintf("In a non-interactive login, `%s` replaces the `--url` flag, and `%s` replaces the `--ca-cert-path` flag.\n\n", pauth.ConfluentPlatformMDSURL, pauth.ConfluentPlatformCACertPath) +
			"Even with the environment variables set, you can force an interactive login using the `--prompt` flag.",
		Args: cobra.NoArgs,
	}

	cmd.Flags().String("url", "", "Metadata Service (MDS) URL for on-prem deployments.")
	cmd.Flags().String("ca-cert-path", "", "Self-signed certificate chain in PEM format.")
	cmd.Flags().Bool("no-browser", false, "Do not open a browser window when authenticating via Single Sign-On (SSO).")
	cmd.Flags().String("organization-id", "", "The Confluent Cloud organization to log in to. If empty, log in to the default organization.")
	cmd.Flags().Bool("prompt", false, "Bypass non-interactive login and prompt for login credentials.")
	cmd.Flags().Bool("save", false, "Save username and encrypted password (non-SSO credentials) to the configuration file in your $HOME directory, and to macOS keychain if applicable. You will be automatically logged back in when your token expires, after one hour for Confluent Cloud or after six hours for Confluent Platform.")

	c := &Command{
		CLICommand:               pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:                      cfg,
		mdsClientManager:         mdsClientManager,
		ccloudClientFactory:      ccloudClientFactory,
		netrcHandler:             netrcHandler,
		loginCredentialsManager:  loginCredentialsManager,
		loginOrganizationManager: pauth.NewLoginOrganizationManagerImpl(),
		authTokenHandler:         authTokenHandler,
	}

	cmd.RunE = pcmd.NewCLIRunE(c.login)
	return c
}

func (c *Command) login(cmd *cobra.Command, _ []string) error {
	url, err := c.getURL(cmd)
	if err != nil {
		return err
	}

	isCCloud := c.isCCloudURL(url)

	url, warningMsg, err := validateURL(url, isCCloud)
	if err != nil {
		return err
	}
	if warningMsg != "" {
		utils.ErrPrintf(cmd, errors.UsingLoginURLDefaults, warningMsg)
	}

	if isCCloud {
		return c.loginCCloud(cmd, url)
	} else {
		return c.loginMDS(cmd, url)
	}
}

func (c *Command) loginCCloud(cmd *cobra.Command, url string) error {
	orgResourceId, err := c.getOrgResourceId(cmd)
	if err != nil {
		return err
	}

	noBrowser, err := cmd.Flags().GetBool("no-browser")
	if err != nil {
		return err
	}

	credentials, err := c.getCCloudCredentials(cmd, url, orgResourceId)
	if err != nil {
		return err
	}

	token, refreshToken, err := c.authTokenHandler.GetCCloudTokens(c.ccloudClientFactory, url, credentials, noBrowser, orgResourceId)
	if err != nil {
		if err, ok := err.(*ccloud.SuspendedOrganizationError); ok {
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

	currentEnv, currentOrg, err := pauth.PersistCCloudLoginToConfig(c.Config.Config, credentials, url, client, save)
	if err != nil {
		return err
	}

	// If refresh token is available, we want to save that in the place of password
	if refreshToken != "" {
		credentials.Password = refreshToken
	}

	log.CliLogger.Debugf(errors.LoggedInAsMsgWithOrg, credentials.Username, currentOrg.ResourceId, currentOrg.Name)
	log.CliLogger.Debugf(errors.LoggedInUsingEnvMsg, currentEnv.Id, currentEnv.Name)

	if save && runtime.GOOS == "darwin" && !c.cfg.IsTest {
		return c.saveLoginToKeychain(cmd, true, url, credentials)
	}
	return nil
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (c *Command) getCCloudCredentials(cmd *cobra.Command, url, orgResourceId string) (*pauth.Credentials, error) {
	client := c.ccloudClientFactory.AnonHTTPClientFactory(url)
	c.loginCredentialsManager.SetCloudClient(client)

	promptOnly, err := cmd.Flags().GetBool("prompt")
	if err != nil {
		return nil, err
	}

	if promptOnly {
		return pauth.GetLoginCredentials(c.loginCredentialsManager.GetCloudCredentialsFromPrompt(cmd, orgResourceId))
	}

	filterParams := netrc.NetrcMachineParams{
		IsCloud: true,
		URL:     url,
	}
	ctx := c.Config.Config.Context()
	if strings.Contains(ctx.GetNetrcMachineName(), url) {
		filterParams.Name = ctx.GetNetrcMachineName()
	}

	return pauth.GetLoginCredentials(
		c.loginCredentialsManager.GetCloudCredentialsFromEnvVar(cmd, orgResourceId),
		c.loginCredentialsManager.GetSsoCredentialsFromConfig(c.cfg, filterParams),
		c.loginCredentialsManager.GetCredentialsFromKeychain(c.cfg, true, filterParams.Name, url),
		c.loginCredentialsManager.GetCredentialsFromConfig(c.cfg, filterParams),
		c.loginCredentialsManager.GetCredentialsFromNetrc(cmd, filterParams),
		c.loginCredentialsManager.GetCloudCredentialsFromPrompt(cmd, orgResourceId),
	)
}

func (c *Command) loginMDS(cmd *cobra.Command, url string) error {
	credentials, err := c.getConfluentCredentials(cmd, url)
	if err != nil {
		return err
	}

	// Current functionality:
	// empty ca-cert-path is equivalent to not using ca-cert-path flag
	// if users want to login with ca-cert-path they must explicilty use the flag every time they login
	//
	// For legacy users:
	// if ca-cert-path flag is not used, then return caCertPath value stored in config for the login context
	// if user passes empty string for ca-cert-path flag then reset the ca-cert-path value in config for the context
	// (only for legacy contexts is it still possible for the context name without ca-cert-path to have ca-cert-path)
	var isLegacyContext bool
	caCertPath, err := getCACertPath(cmd)
	if err != nil {
		return err
	}
	if caCertPath == "" {
		contextName := pauth.GenerateContextName(credentials.Username, url, "")
		caCertPath, err = c.checkLegacyContextCACertPath(cmd, contextName)
		if err != nil {
			return err
		}
		isLegacyContext = caCertPath != ""
	}

	if caCertPath != "" {
		caCertPath, err = filepath.Abs(caCertPath)
		if err != nil {
			return err
		}
	}

	client, err := c.mdsClientManager.GetMDSClient(url, caCertPath)

	if err != nil {
		return err
	}

	token, err := c.authTokenHandler.GetConfluentToken(client, credentials)
	if err != nil {
		return err
	}

	save, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}

	err = pauth.PersistConfluentLoginToConfig(c.Config.Config, credentials, url, token, caCertPath, isLegacyContext, save)
	if err != nil {
		return err
	}

	if save && runtime.GOOS == "darwin" && !c.cfg.IsTest {
		if err := c.saveLoginToKeychain(cmd, false, url, credentials); err != nil {
			return err
		}
	}

	log.CliLogger.Debugf(errors.LoggedInAsMsg, credentials.Username)
	return nil
}

func getCACertPath(cmd *cobra.Command) (string, error) {
	if path, err := cmd.Flags().GetString("ca-cert-path"); path != "" || err != nil {
		return path, err
	}

	return pauth.GetEnvWithFallback(pauth.ConfluentPlatformCACertPath, pauth.DeprecatedConfluentPlatformCACertPath), nil
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (c *Command) getConfluentCredentials(cmd *cobra.Command, url string) (*pauth.Credentials, error) {
	promptOnly, err := cmd.Flags().GetBool("prompt")
	if err != nil {
		return nil, err
	}
	if promptOnly {
		return pauth.GetLoginCredentials(c.loginCredentialsManager.GetOnPremCredentialsFromPrompt(cmd))
	}

	netrcFilterParams := netrc.NetrcMachineParams{
		IgnoreCert: true,
		URL:        url,
	}
	ctx := c.Config.Config.Context()
	if strings.Contains(ctx.GetNetrcMachineName(), url) {
		netrcFilterParams.Name = ctx.GetNetrcMachineName()
	}

	return pauth.GetLoginCredentials(
		c.loginCredentialsManager.GetOnPremCredentialsFromEnvVar(cmd),
		c.loginCredentialsManager.GetCredentialsFromKeychain(c.cfg, false, netrcFilterParams.Name, url),
		c.loginCredentialsManager.GetCredentialsFromConfig(c.cfg, netrcFilterParams),
		c.loginCredentialsManager.GetCredentialsFromNetrc(cmd, netrcFilterParams),
		c.loginCredentialsManager.GetOnPremCredentialsFromPrompt(cmd),
	)
}

func (c *Command) checkLegacyContextCACertPath(cmd *cobra.Command, contextName string) (string, error) {
	changed := cmd.Flags().Changed("ca-cert-path")
	// if flag used but empty string is passed then user intends to reset the ca-cert-path
	if changed {
		return "", nil
	}
	ctx, ok := c.Config.Contexts[contextName]
	if !ok {
		return "", nil
	}
	return ctx.Platform.CaCertPath, nil
}

func (c *Command) getURL(cmd *cobra.Command) (string, error) {
	if url, err := cmd.Flags().GetString("url"); url != "" || err != nil {
		return url, err
	}

	if url := pauth.GetEnvWithFallback(pauth.ConfluentPlatformMDSURL, pauth.DeprecatedConfluentPlatformMDSURL); url != "" {
		return url, nil
	}

	return pauth.CCloudURL, nil
}

func (c *Command) saveLoginToKeychain(cmd *cobra.Command, isCloud bool, url string, credentials *pauth.Credentials) error {
	if credentials.IsSSO {
		utils.ErrPrintln(cmd, "The `--save` flag was ignored since SSO credentials are not stored locally.")
		return nil
	}

	ctxName := c.Config.Config.Context().GetNetrcMachineName()
	if err := keychain.Write(isCloud, ctxName, url, credentials.Username, credentials.Password); err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.WroteCredentialsToKeychainMsg)

	return nil
}

func validateURL(url string, isCCloud bool) (string, string, error) {
	if isCCloud {
		for _, hostname := range v1.CCloudHostnames {
			if strings.Contains(url, hostname) {
				if !strings.HasSuffix(strings.TrimSuffix(url, "/"), hostname) {
					return url, "", errors.NewErrorWithSuggestions(errors.UnneccessaryUrlFlagForCloudLoginErrorMsg, errors.UnneccessaryUrlFlagForCloudLoginSuggestions)
				} else {
					break
				}
			}
		}
	}
	protocolRgx, _ := regexp.Compile(`(\w+)://`)
	portRgx, _ := regexp.Compile(`:(\d+\/?)`)

	protocolMatch := protocolRgx.MatchString(url)
	portMatch := portRgx.MatchString(url)

	var msg []string
	if !protocolMatch {
		if isCCloud {
			url = "https://" + url
			msg = append(msg, "https protocol")
		} else {
			url = "http://" + url
			msg = append(msg, "http protocol")
		}
	}
	if !portMatch && !isCCloud {
		url = url + ":8090"
		msg = append(msg, "default MDS port 8090")
	}

	var pattern string
	if isCCloud {
		pattern = `^\w+://[^/ ]+`
	} else {
		pattern = `^\w+://[^/ ]+:\d+(?:\/|$)`
	}
	matched, _ := regexp.MatchString(pattern, url)
	if !matched {
		return url, "", errors.New(errors.InvalidLoginURLMsg)
	}

	return url, strings.Join(msg, " and "), nil
}

func (c *Command) isCCloudURL(url string) bool {
	for _, hostname := range v1.CCloudHostnames {
		if strings.Contains(url, hostname) {
			return true
		}
	}
	if c.cfg.IsTest {
		return strings.Contains(url, testserver.TestCloudURL.Host)
	}
	return false
}

func (c *Command) getOrgResourceId(cmd *cobra.Command) (string, error) {
	return pauth.GetLoginOrganization(
		c.loginOrganizationManager.GetLoginOrganizationFromArgs(cmd),
		c.loginOrganizationManager.GetLoginOrganizationFromEnvVar(cmd),
		c.loginOrganizationManager.GetDefaultLoginOrganization(),
	)
}
