package auth

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type loginCommand struct {
	*pcmd.CLICommand
	cliName         string
	Logger          *log.Logger
	analyticsClient analytics.Client
	// for testing
	ccloudClientFactory     pauth.CCloudClientFactory
	MDSClientManager        pauth.MDSClientManager
	netrcHandler            netrc.NetrcHandler
	loginCredentialsManager pauth.LoginCredentialsManager
	authTokenHandler        pauth.AuthTokenHandler
}

func NewLoginCommand(cliName string, prerunner pcmd.PreRunner, log *log.Logger, ccloudClientFactory pauth.CCloudClientFactory,
	mdsClientManager pauth.MDSClientManager, analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler,
	loginCredentialsManager pauth.LoginCredentialsManager, authTokenHandler pauth.AuthTokenHandler) *loginCommand {
	cmd := &loginCommand{
		cliName:                 cliName,
		Logger:                  log,
		analyticsClient:         analyticsClient,
		MDSClientManager:        mdsClientManager,
		ccloudClientFactory:     ccloudClientFactory,
		netrcHandler:            netrcHandler,
		loginCredentialsManager: loginCredentialsManager,
		authTokenHandler:        authTokenHandler,
	}
	cmd.init(prerunner)
	return cmd
}

func (a *loginCommand) init(prerunner pcmd.PreRunner) {
	var longDesc string
	
	remoteAPIName := getRemoteAPIName(a.cliName)
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: fmt.Sprintf("Log in to %s.", remoteAPIName),
		Args:  cobra.NoArgs,
		PersistentPreRunE: pcmd.NewCLIPreRunnerE(func(cmd *cobra.Command, args []string) error {
			a.analyticsClient.SetCommandType(analytics.Login)
			return a.CLICommand.PersistentPreRunE(cmd, args)
		}),
	}
	if a.cliName == "ccloud" {
		longDesc = fmt.Sprintf("Log in to %s.\n\n%s\n\n%s", remoteAPIName, "Starting in the 1.20.1 release, you can log in to Confluent Cloud non-interactively using the CCLOUD_EMAIL and CCLOUD_PASSWORD environment variables.", "Even with the above environment variables set, you can force an interactive login using the --prompt flag.")
		loginCmd.Long = longDesc
		loginCmd.RunE = pcmd.NewCLIRunE(a.login)
		loginCmd.Flags().String("url", pauth.CCloudURL, "Confluent Cloud service URL.")
	} else {
		longDesc=fmt.Sprintf("Log in to %s.\n\n%s\n\n", remoteAPIName, "Starting in the 1.20.1 release, you can log in to Confluent Platform non-interactively using the following environment variables: CONFLUENT_USERNAME, CONFLUENT_PASSWORD, CONFLUENT_MDS_URL, CONFLUENT_CA_CERT_PATH")
		longDesc += "In a non-interactive login, CONFLUENT_MDS_URL replaces the ``--url`` flag, and CONFLUENT_CA_CERT_PATH replaces the ``--ca-cert-path`` flag.\n\n"
		longDesc += "Even with the above environment variables set, you can force an interactive login using the --prompt flag."
		loginCmd.Long = longDesc
		loginCmd.RunE = pcmd.NewCLIRunE(a.loginMDS)
		loginCmd.Flags().String("url", "", "Metadata service URL.")
		loginCmd.Flags().String("ca-cert-path", "", "Self-signed certificate chain in PEM format.")
		loginCmd.Short = strings.ReplaceAll(loginCmd.Short, ".", " (required for RBAC).")
		loginCmd.Long = fmt.Sprintf("%s \n\n%s", loginCmd.Long, "This command is required for RBAC.")
		check(loginCmd.MarkFlagRequired("url")) // because https://confluent.cloud isn't an MDS endpoint
	}
	loginCmd.Flags().Bool("no-browser", false, "Do not open browser when authenticating via Single Sign-On.")
	loginCmd.Flags().Bool("prompt", false, "Bypass non-interactive login and prompt for login credentials.")
	loginCmd.Flags().Bool("save", false, "Save login credentials or refresh token (in the case of SSO) to local netrc file.")
	loginCmd.Flags().SortFlags = false
	cliLoginCmd := pcmd.NewAnonymousCLICommand(loginCmd, prerunner)
	a.CLICommand = cliLoginCmd
}

func getRemoteAPIName(cliName string) string {
	if cliName == "ccloud" {
		return "Confluent Cloud"
	}
	return "Confluent Platform"
}

func (a *loginCommand) login(cmd *cobra.Command, _ []string) error {
	url, err := a.getURL(cmd)
	if err != nil {
		return err
	}

	credentials, err := a.getCCloudCredentials(cmd, url)
	if err != nil {
		return err
	}

	noBrowser, err := cmd.Flags().GetBool("no-browser")
	if err != nil {
		return err
	}

	client := a.ccloudClientFactory.AnonHTTPClientFactory(url)
	token, refreshToken, err := a.authTokenHandler.GetCCloudTokens(client, credentials, noBrowser)
	if err != nil {
		return err
	}

	client = a.ccloudClientFactory.JwtHTTPClientFactory(context.Background(), token, url)

	currentEnv, err := pauth.PersistCCloudLoginToConfig(a.Config.Config, credentials.Username, url, token, client)
	if err != nil {
		return err
	}

	// If refresh token is available, we want to save that in the place of password
	if refreshToken != "" {
		credentials.Password = refreshToken
	}
	err = a.saveLoginToNetrc(cmd, credentials)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.LoggedInAsMsg, credentials.Username)
	utils.Printf(cmd, errors.LoggedInUsingEnvMsg, currentEnv.Id, currentEnv.Name)
	return err
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (a *loginCommand) getCCloudCredentials(cmd *cobra.Command, url string) (*pauth.Credentials, error) {
	client := a.ccloudClientFactory.AnonHTTPClientFactory(url)
	promptOnly, err := cmd.Flags().GetBool("prompt")
	if err != nil {
		return nil, err
	}

	if promptOnly {
		return pauth.GetLoginCredentials(a.loginCredentialsManager.GetCCloudCredentialsFromPrompt(cmd, client))
	}
	netrcFilterParams := netrc.GetMatchingNetrcMachineParams{
		CLIName: a.cliName,
		URL:     url,
	}
	return pauth.GetLoginCredentials(
		a.loginCredentialsManager.GetCCloudCredentialsFromEnvVar(cmd),
		a.loginCredentialsManager.GetCredentialsFromNetrc(cmd, netrcFilterParams),
		a.loginCredentialsManager.GetCCloudCredentialsFromPrompt(cmd, client),
	)
}

func (a *loginCommand) loginMDS(cmd *cobra.Command, _ []string) error {
	url, err := a.getURL(cmd)
	if err != nil {
		return err
	}

	credentials, err := a.getConfluentCredentials(cmd, url)
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
	caCertPath, err := cmd.Flags().GetString("ca-cert-path")
	if err != nil {
		return err
	}
	if caCertPath == "" {
		contextName := pauth.GenerateContextName(credentials.Username, url, "")
		caCertPath, err = a.checkLegacyContextCaCertPath(cmd, contextName)
		if err != nil {
			return err
		}
		isLegacyContext = caCertPath != ""
	}

	client, err := a.MDSClientManager.GetMDSClient(url, caCertPath, a.Logger)
	if err != nil {
		return err
	}

	token, err := a.authTokenHandler.GetConfluentToken(client, credentials)
	if err != nil {
		return err
	}

	err = pauth.PersistConfluentLoginToConfig(a.Config.Config, credentials.Username, url, token, caCertPath, isLegacyContext)
	if err != nil {
		return err
	}

	err = a.saveLoginToNetrc(cmd, credentials)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.LoggedInAsMsg, credentials.Username)
	return nil
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (a *loginCommand) getConfluentCredentials(cmd *cobra.Command, url string) (*pauth.Credentials, error) {
	promptOnly, err := cmd.Flags().GetBool("prompt")
	if err != nil {
		return nil, err
	}

	if promptOnly {
		return pauth.GetLoginCredentials(a.loginCredentialsManager.GetConfluentCredentialsFromPrompt(cmd))
	}
	netrcFilterParams := netrc.GetMatchingNetrcMachineParams{
		CLIName: a.cliName,
		URL:     url,
	}
	return pauth.GetLoginCredentials(
		a.loginCredentialsManager.GetConfluentCredentialsFromEnvVar(cmd),
		a.loginCredentialsManager.GetCredentialsFromNetrc(cmd, netrcFilterParams),
		a.loginCredentialsManager.GetConfluentCredentialsFromPrompt(cmd),
	)
}

func (a *loginCommand) checkLegacyContextCaCertPath(cmd *cobra.Command, contextName string) (string, error) {
	changed := cmd.Flags().Changed("ca-cert-path")
	// if flag used but empty string is passed then user intends to reset the ca-cert-path
	if changed {
		return "", nil
	}
	ctx, ok := a.Config.Contexts[contextName]
	if !ok {
		return "", nil
	}
	return ctx.Platform.CaCertPath, nil
}

func (a *loginCommand) getURL(cmd *cobra.Command) (string, error) {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return "", err
	}
	url, valid, errMsg := validateURL(url, a.cliName)
	if !valid {
		return "", errors.Errorf(errors.InvalidLoginURLMsg)
	}
	if errMsg != "" {
		utils.ErrPrintf(cmd, errors.UsingLoginURLDefaults, errMsg)
	}
	return url, nil
}

func (a *loginCommand) saveLoginToNetrc(cmd *cobra.Command, credentials *pauth.Credentials) error {
	saveToNetrc, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}
	if saveToNetrc {
		err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, credentials.IsSSO, a.Config.Config.Context().Name, credentials.Username, credentials.Password)
		if err != nil {
			return err
		}
		utils.ErrPrintf(cmd, errors.WroteCredentialsToNetrcMsg, a.netrcHandler.GetFileName())
	}
	return nil
}

func validateURL(url string, cli string) (string, bool, string) {
	protocol_rgx, _ := regexp.Compile(`(\w+)://`)
	port_rgx, _ := regexp.Compile(`:(\d+\/?)`)

	protocol_match := protocol_rgx.MatchString(url)
	port_match := port_rgx.MatchString(url)

	var msg []string
	if !protocol_match {
		if cli == "ccloud" {
			url = "https://" + url
			msg = append(msg, "https protocol")
		} else {
			url = "http://" + url
			msg = append(msg, "http protocol")
		}
	}
	if !port_match && cli == "confluent" {
		url = url + ":8090"
		msg = append(msg, "default MDS port 8090")
	}
	var pattern string
	if cli == "confluent" {
		pattern = `^\w+://[^/ ]+:\d+(?:\/|$)`
	} else {
		pattern = `^\w+://[^/ ]+`
	}
	matched, _ := regexp.Match(pattern, []byte(url))

	return url, matched, strings.Join(msg, " and ")
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
