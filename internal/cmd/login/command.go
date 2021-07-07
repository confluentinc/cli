package login

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
	testserver "github.com/confluentinc/cli/test/test-server"
)

type Command struct {
	*pcmd.CLICommand
	logger                  *log.Logger
	analyticsClient         analytics.Client
	ccloudClientFactory     pauth.CCloudClientFactory
	mdsClientManager        pauth.MDSClientManager
	netrcHandler            netrc.NetrcHandler
	loginCredentialsManager pauth.LoginCredentialsManager
	authTokenHandler        pauth.AuthTokenHandler
	isTest                  bool
}

func New(prerunner pcmd.PreRunner, log *log.Logger, ccloudClientFactory pauth.CCloudClientFactory,
	mdsClientManager pauth.MDSClientManager, analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler,
	loginCredentialsManager pauth.LoginCredentialsManager, authTokenHandler pauth.AuthTokenHandler, isTest bool) *Command {
	cmd := &Command{
		logger:                  log,
		analyticsClient:         analyticsClient,
		mdsClientManager:        mdsClientManager,
		ccloudClientFactory:     ccloudClientFactory,
		netrcHandler:            netrcHandler,
		loginCredentialsManager: loginCredentialsManager,
		authTokenHandler:        authTokenHandler,
		isTest:                  isTest,
	}
	cmd.init(prerunner)
	return cmd
}

func (c *Command) init(prerunner pcmd.PreRunner) {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Confluent Cloud or Confluent Platform.",
		Long: "Log in to Confluent Cloud or using your email and password, or non-interactively using the ``CCLOUD_EMAIL`` and ``CCLOUD_PASSWORD`` environment variables.\n\n" +
			"You can log in to Confluent Platform with your username and password, or non-interactively using ``CONFLUENT_USERNAME``, ``CONFLUENT_PASSWORD``, ``CONFLUENT_MDS_URL``, and ``CONFLUENT_CA_CERT_PATH``. In a non-interactive login, ``CONFLUENT_MDS_URL`` replaces the ``--url`` flag, and ``CONFLUENT_CA_CERT_PATH`` replaces the ``--ca-cert-path`` flag.\n\n" +
			"Even with environment variables set, you can force an interactive login using the ``--prompt`` flag.",
		Args:              cobra.NoArgs,
		RunE:              pcmd.NewCLIRunE(c.login),
		PersistentPreRunE: pcmd.NewCLIPreRunnerE(c.loginPreRunE),
	}

	loginCmd.Flags().String("url", "", "Metadata Service (MDS) URL for on-prem deployments.")
	loginCmd.Flags().String("ca-cert-path", "", "Self-signed certificate chain in PEM format.")
	loginCmd.Flags().Bool("save", false, "Save login credentials or SSO refresh token to local .netrc file.")
	loginCmd.Flags().Bool("prompt", false, "Bypass non-interactive login and prompt for login credentials.")
	loginCmd.Flags().Bool("no-browser", false, "Do not open a browser window when authenticating via Single Sign-On (SSO).")
	loginCmd.Flags().SortFlags = false

	c.CLICommand = pcmd.NewAnonymousCLICommand(loginCmd, prerunner)
}

func (c *Command) loginPreRunE(cmd *cobra.Command, args []string) error {
	c.analyticsClient.SetCommandType(analytics.Login)
	return c.CLICommand.PersistentPreRunE(cmd, args)
}

func (c *Command) login(cmd *cobra.Command, _ []string) error {
	url, err := c.getURL(cmd)
	if err != nil {
		return err
	}

	isCCloud := c.isCCloudURL(url)

	url, valid, errMsg := validateURL(url, isCCloud)
	if !valid {
		return errors.New(errors.InvalidLoginURLMsg)
	}
	if errMsg != "" {
		utils.ErrPrintf(cmd, errors.UsingLoginURLDefaults, errMsg)
	}

	if isCCloud {
		return c.loginCCloud(cmd, url)
	} else {
		return c.loginMDS(cmd, url)
	}
}

func (c *Command) loginCCloud(cmd *cobra.Command, url string) error {
	credentials, err := c.getCCloudCredentials(cmd, url)
	if err != nil {
		return err
	}

	noBrowser, err := cmd.Flags().GetBool("no-browser")
	if err != nil {
		return err
	}

	client := c.ccloudClientFactory.AnonHTTPClientFactory(url)
	token, refreshToken, err := c.authTokenHandler.GetCCloudTokens(client, credentials, noBrowser)
	if err != nil {
		return err
	}

	client = c.ccloudClientFactory.JwtHTTPClientFactory(context.Background(), token, url)

	currentEnv, err := pauth.PersistCCloudLoginToConfig(c.Config.Config, credentials.Username, url, token, client)
	if err != nil {
		return err
	}

	// If refresh token is available, we want to save that in the place of password
	if refreshToken != "" {
		credentials.Password = refreshToken
	}

	if err := c.saveLoginToNetrc(cmd, credentials); err != nil {
		return err
	}

	c.logger.Debugf(errors.LoggedInAsMsg, credentials.Username)
	c.logger.Debugf(errors.LoggedInUsingEnvMsg, currentEnv.Id, currentEnv.Name)

	return err
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (c *Command) getCCloudCredentials(cmd *cobra.Command, url string) (*pauth.Credentials, error) {
	client := c.ccloudClientFactory.AnonHTTPClientFactory(url)
	promptOnly, err := cmd.Flags().GetBool("prompt")
	if err != nil {
		return nil, err
	}

	if promptOnly {
		return pauth.GetLoginCredentials(c.loginCredentialsManager.GetCCloudCredentialsFromPrompt(cmd, client))
	}
	netrcFilterParams := netrc.NetrcMachineParams{
		CLIName: "ccloud",
		URL:     url,
	}
	return pauth.GetLoginCredentials(
		c.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(cmd),
		c.loginCredentialsManager.GetCredentialsFromNetrc(cmd, netrcFilterParams),
		c.loginCredentialsManager.GetCCloudCredentialsFromPrompt(cmd, client),
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

	client, err := c.mdsClientManager.GetMDSClient(url, caCertPath, c.logger)
	if err != nil {
		return err
	}

	token, err := c.authTokenHandler.GetConfluentToken(client, credentials)
	if err != nil {
		return err
	}

	err = pauth.PersistConfluentLoginToConfig(c.Config.Config, credentials.Username, url, token, caCertPath, isLegacyContext)
	if err != nil {
		return err
	}

	err = c.saveLoginToNetrc(cmd, credentials)
	if err != nil {
		return err
	}

	c.logger.Debugf(errors.LoggedInAsMsg, credentials.Username)
	return nil
}

func getCACertPath(cmd *cobra.Command) (string, error) {
	caCertPath, err := cmd.Flags().GetString("ca-cert-path")
	if err != nil {
		return "", err
	}
	if caCertPath != "" {
		return caCertPath, nil
	}

	return os.Getenv(pauth.ConfluentCACertPathEnvVar), nil
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (c *Command) getConfluentCredentials(cmd *cobra.Command, url string) (*pauth.Credentials, error) {
	promptOnly, err := cmd.Flags().GetBool("prompt")
	if err != nil {
		return nil, err
	}
	if promptOnly {
		return pauth.GetLoginCredentials(c.loginCredentialsManager.GetConfluentCredentialsFromPrompt(cmd))
	}

	netrcFilterParams := netrc.NetrcMachineParams{
		CLIName: "confluent",
		URL:     url,
	}

	return pauth.GetLoginCredentials(
		c.loginCredentialsManager.GetConfluentCredentialsFromEnvVar(cmd),
		c.loginCredentialsManager.GetCredentialsFromNetrc(cmd, netrcFilterParams),
		c.loginCredentialsManager.GetConfluentCredentialsFromPrompt(cmd),
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
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return "", err
	}
	if url != "" {
		return url, nil
	}

	if url := os.Getenv(pauth.ConfluentURLEnvVar); url != "" {
		return url, nil
	}

	return pauth.CCloudURL, nil
}

func (c *Command) saveLoginToNetrc(cmd *cobra.Command, credentials *pauth.Credentials) error {
	save, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}

	if save {
		if err := c.netrcHandler.WriteNetrcCredentials(c.Config.CLIName, credentials.IsSSO, c.Config.Config.Context().Name, credentials.Username, credentials.Password); err != nil {
			return err
		}
		utils.ErrPrintf(cmd, errors.WroteCredentialsToNetrcMsg, c.netrcHandler.GetFileName())
	}

	return nil
}

func validateURL(url string, isCCloud bool) (string, bool, string) {
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

	return url, matched, strings.Join(msg, " and ")
}

func (c *Command) isCCloudURL(url string) bool {
	for _, hostname := range v3.CCloudHostnames {
		if strings.Contains(url, hostname) {
			return true
		}
	}

	if c.isTest {
		return strings.Contains(url, testserver.TestCloudURL.Host)
	}

	return false
}
