package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type commands struct {
	Commands        []*pcmd.CLICommand
	Logger          *log.Logger
	config          *v3.Config
	analyticsClient analytics.Client
	// for testing
	MDSClientManager      pauth.MDSClientManager
	prompt                pcmd.Prompt
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client
	jwtHTTPClientFactory  func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client
	netrcHandler          *pauth.NetrcHandler
}

var (
	LoginIndex = 0
)

// New returns a list of auth-related Cobra commands.
func New(prerunner pcmd.PreRunner, config *v3.Config, logger *log.Logger, userAgent string, analyticsClient analytics.Client, netrcHandler *pauth.NetrcHandler) []*cobra.Command {
	var defaultAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClient(&ccloud.Params{BaseURL: baseURL, HttpClient: ccloud.BaseClient, Logger: logger, UserAgent: userAgent})
	}
	var defaultJwtHTTPClientFactory = func(ctx context.Context, jwt string, baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClientWithJWT(ctx, jwt, &ccloud.Params{BaseURL: baseURL, Logger: logger, UserAgent: userAgent})
	}
	cmds := newCommands(prerunner, config, logger, pcmd.NewPrompt(os.Stdin),
		defaultAnonHTTPClientFactory, defaultJwtHTTPClientFactory, &pauth.MDSClientManagerImpl{},
		analyticsClient, netrcHandler,
	)
	var cobraCmds []*cobra.Command
	for _, cmd := range cmds.Commands {
		cobraCmds = append(cobraCmds, cmd.Command)
	}
	return cobraCmds
}

func newCommands(prerunner pcmd.PreRunner, config *v3.Config, log *log.Logger, prompt pcmd.Prompt,
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client,
	jwtHTTPClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client,
	mdsClientManager pauth.MDSClientManager, analyticsClient analytics.Client, netrcHandler *pauth.NetrcHandler) *commands {
	cmd := &commands{
		config:                config,
		Logger:                log,
		prompt:                prompt,
		analyticsClient:       analyticsClient,
		anonHTTPClientFactory: anonHTTPClientFactory,
		jwtHTTPClientFactory:  jwtHTTPClientFactory,
		MDSClientManager:      mdsClientManager,
		netrcHandler:          netrcHandler,
	}
	cmd.init(prerunner)
	return cmd
}

func (a *commands) init(prerunner pcmd.PreRunner) {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: fmt.Sprintf("Log in to %s.", a.config.APIName()),
		Long:  fmt.Sprintf("Log in to %s.", a.config.APIName()),
		Args:  cobra.NoArgs,
	}
	if a.config.CLIName == "ccloud" {
		loginCmd.RunE = a.login
		loginCmd.Flags().String("url", "https://confluent.cloud", "Confluent Cloud service URL.")
	} else {
		loginCmd.RunE = a.loginMDS
		loginCmd.Flags().String("url", "", "Metadata service URL.")
		loginCmd.Flags().String("ca-cert-path", "", "Self-signed certificate chain in PEM format.")
		loginCmd.Short = strings.Replace(loginCmd.Short, ".", " (required for RBAC).", -1)
		loginCmd.Long = strings.Replace(loginCmd.Long, ".", " (required for RBAC).", -1)
		check(loginCmd.MarkFlagRequired("url")) // because https://confluent.cloud isn't an MDS endpoint
	}
	loginCmd.Flags().Bool("no-browser", false, "Do not open browser when authenticating via Single Sign-On.")
	loginCmd.Flags().Bool("save", false, "Save login credentials or refresh token (in the case of SSO) to local netrc file.")
	loginCmd.Flags().SortFlags = false
	cliLoginCmd := pcmd.NewAnonymousCLICommand(loginCmd, a.config, prerunner)
	loginCmd.PersistentPreRunE = a.analyticsPreRunCover(cliLoginCmd, analytics.Login, prerunner)
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: fmt.Sprintf("Logout of %s.", a.config.APIName()),
		Long:  fmt.Sprintf("Logout of %s.", a.config.APIName()),

		RunE: a.logout,
		Args: cobra.NoArgs,
	}
	cliLogoutCmd := pcmd.NewAnonymousCLICommand(logoutCmd, a.config, prerunner)
	logoutCmd.PersistentPreRunE = a.analyticsPreRunCover(cliLogoutCmd, analytics.Logout, prerunner)
	a.Commands = []*pcmd.CLICommand{cliLoginCmd, cliLogoutCmd}
}

func (a *commands) login(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	noBrowser, err := cmd.Flags().GetBool("no-browser")
	if err != nil {
		return err
	}
	a.config.NoBrowser = noBrowser

	client := a.anonHTTPClientFactory(url, a.config.Logger)
	email, password, err := a.credentials(cmd, "Email", client)
	if err != nil {
		return err
	}

	token, refreshToken, err := pauth.GetCCloudAuthToken(client, url, email, password, noBrowser)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	client = a.jwtHTTPClientFactory(context.Background(), token, url, a.config.Logger)
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if len(user.Accounts) == 0 {
		return errors.HandleCommon(errors.New("No environments found for authenticated user!"), cmd)
	}
	username := user.User.Email
	name := generateContextName(username, url)
	var state *v2.ContextState
	ctx, err := a.config.FindContext(name)
	if err == nil {
		state = ctx.State
	} else {
		state = new(v2.ContextState)
	}
	state.AuthToken = token

	if state.Auth == nil {
		state.Auth = &v1.AuthConfig{}
	}

	// Always overwrite the user and list of accounts when logging in -- but don't necessarily
	// overwrite `Account` (current/active environment) since we want that to be remembered
	// between CLI sessions.
	state.Auth.User = user.User
	state.Auth.Accounts = user.Accounts

	// Default to 0th environment if no suitable environment is already configured
	hasGoodEnv := false
	if state.Auth.Account != nil {
		for _, acc := range state.Auth.Accounts {
			if acc.Id == state.Auth.Account.Id {
				hasGoodEnv = true
			}
		}
	}
	if !hasGoodEnv {
		state.Auth.Account = state.Auth.Accounts[0]
	}

	err = a.addOrUpdateContext(state.Auth.User.Email, url, state, "")
	if err != nil {
		return err
	}
	err = a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to save user authentication")
	}

	saveToNetrc, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}
	if saveToNetrc {
		err = a.saveToNetrc(cmd, email, password, refreshToken)
		if err != nil {
			return err
		}
	}

	pcmd.Println(cmd, "Logged in as", email)
	pcmd.Print(cmd, "Using environment ", state.Auth.Account.Id,
		" (\"", state.Auth.Account.Name, "\")\n")
	return err
}

func (a *commands) loginMDS(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	email, password, err := a.credentials(cmd, "Username", nil)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	dynamicContext, err := a.Commands[0].Config.Context(cmd)
	if err != nil {
		return err
	}
	var ctx *v3.Context
	if dynamicContext != nil {
		ctx = dynamicContext.Context
	}
	flagChanged := cmd.Flags().Changed("ca-cert-path")
	caCertPath, err := cmd.Flags().GetString("ca-cert-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	mdsClient, err := a.MDSClientManager.GetMDSClient(ctx, caCertPath, flagChanged, url, a.Logger)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	authToken, err := pauth.GetConfluentAuthToken(mdsClient, email, password)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	basicContext := context.WithValue(context.Background(), mds.ContextBasicAuth, mds.BasicAuth{UserName: email, Password: password})
	_, _, err = mdsClient.TokensAndAuthenticationApi.GetToken(basicContext)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	state := &v2.ContextState{
		Auth:      nil,
		AuthToken: authToken,
	}
	err = a.addOrUpdateContext(email, url, state, caCertPath)
	if err != nil {
		return err
	}
	saveToNetrc, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}
	if saveToNetrc {
		err = a.saveToNetrc(cmd, email, password, "")
		if err != nil {
			return err
		}
	}
	pcmd.Println(cmd, "Logged in as", email)
	return nil
}

func (a *commands) logout(cmd *cobra.Command, args []string) error {
	ctx := a.config.Context()
	if ctx == nil {
		return nil
	}
	err := ctx.DeleteUserAuth()
	if err != nil {
		return err
	}
	err = a.config.Save()
	if err != nil {
		return err
	}
	pcmd.Println(cmd, "You are now logged out")
	return nil
}

func (a *commands) credentials(cmd *cobra.Command, userField string, cloudClient *ccloud.Client) (string, string, error) {
	email := os.Getenv("XX_CCLOUD_EMAIL")
	if len(email) == 0 {
		email = os.Getenv("XX_CONFLUENT_USERNAME")
	}
	password := os.Getenv("XX_CCLOUD_PASSWORD")
	if len(password) == 0 {
		password = os.Getenv("XX_CONFLUENT_PASSWORD")
	}
	if len(email) == 0 || len(password) == 0 {
		pcmd.Println(cmd, "Enter your Confluent credentials:")
	}
	if len(email) == 0 {
		pcmd.Print(cmd, userField+": ")
		emailFromPrompt, err := a.prompt.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		email = strings.TrimSpace(emailFromPrompt)
	}

	a.Logger.Trace("Successfully obtained email")

	// In the case of MDS login (`confluent`) or in the case of some of the mocks,
	// cloudClient will be nll, so we need this check
	if cloudClient != nil {
		// If SSO user, don't prompt for password
		userSSO, err := cloudClient.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
		// Fine to ignore non-nil err for this request: e.g. what if this fails due to invalid/malicious
		// email, we want to silently continue and give the illusion of password prompt.
		// If Auth0ConnectionName is blank ("local" user) still prompt for password
		if err == nil && userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
			a.Logger.Trace("User is SSO-enabled so won't prompt for password")
			return email, password, nil
		}
	}

	if len(password) == 0 {
		var err error
		pcmd.Print(cmd, "Password: ")
		bytePassword, err := a.prompt.ReadPassword()
		if err != nil {
			return "", "", err
		}
		pcmd.Println(cmd)
		password = string(bytePassword)
	}

	a.Logger.Trace("Successfully obtained password")

	return email, password, nil
}

func (a *commands) addOrUpdateContext(username string, url string, state *v2.ContextState, caCertPath string) error {
	ctxName := generateContextName(username, url)
	credName := generateCredentialName(username)
	platform := &v2.Platform{
		Name:       strings.TrimPrefix(url, "https://"),
		Server:     url,
		CaCertPath: caCertPath,
	}
	credential := &v2.Credential{
		Name:     credName,
		Username: username,
		// don't save password if they entered it interactively.
	}
	err := a.config.SavePlatform(platform)
	if err != nil {
		return err
	}
	err = a.config.SaveCredential(credential)
	if err != nil {
		return err
	}
	if ctx, ok := a.config.Contexts[ctxName]; ok {
		a.config.ContextStates[ctxName] = state
		ctx.State = state
	} else {
		err = a.config.AddContext(ctxName, platform.Name, credential.Name, map[string]*v1.KafkaClusterConfig{},
			"", nil, state)
	}
	if err != nil {
		return err
	}
	err = a.config.SetContext(ctxName)
	if err != nil {
		return err
	}
	return nil
}

func (a *commands) analyticsPreRunCover(command *pcmd.CLICommand, commandType analytics.CommandType, prerunner pcmd.PreRunner) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		a.analyticsClient.SetCommandType(commandType)
		return prerunner.Anonymous(command)(cmd, args)
	}
}

func (a *commands) saveToNetrc(cmd *cobra.Command, email, password, refreshToken string) error {
	// sso if refresh token is empty
	var err error
	if refreshToken == "" {
		err = a.netrcHandler.WriteNetrcCredentials(a.config.CLIName, false, a.config.Context().Name, email, password)
	} else {
		err = a.netrcHandler.WriteNetrcCredentials(a.config.CLIName, true, a.config.Context().Name, email, refreshToken)
	}
	if err != nil {
		return err
	}
	pcmd.ErrPrintf(cmd, "Written credentials to file %s\n", a.netrcHandler.FileName)
	return nil
}

func generateContextName(username string, url string) string {
	return fmt.Sprintf("login-%s-%s", username, url)
}

func generateCredentialName(username string) string {
	return fmt.Sprintf("username-%s", username)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
