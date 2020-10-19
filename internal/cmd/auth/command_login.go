package auth

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type loginCommand struct {
	*pcmd.CLICommand
	cliName         string
	Logger          *log.Logger
	analyticsClient analytics.Client
	// for testing
	MDSClientManager      pauth.MDSClientManager
	prompt                pcmd.Prompt
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client
	jwtHTTPClientFactory  func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client
	netrcHandler          pauth.NetrcHandler
}

func NewLoginCommand(cliName string, prerunner pcmd.PreRunner, log *log.Logger, prompt pcmd.Prompt,
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client,
	jwtHTTPClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client,
	mdsClientManager pauth.MDSClientManager, analyticsClient analytics.Client, netrcHandler pauth.NetrcHandler) *loginCommand {
	cmd := &loginCommand{
		cliName:               cliName,
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

func (a *loginCommand) init(prerunner pcmd.PreRunner) {
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
		loginCmd.RunE = pcmd.NewCLIRunE(a.login)
		loginCmd.Flags().String("url", "https://confluent.cloud", "Confluent Cloud service URL.")
	} else {
		loginCmd.RunE = pcmd.NewCLIRunE(a.loginMDS)
		loginCmd.Flags().String("url", "", "Metadata service URL.")
		loginCmd.Flags().String("ca-cert-path", "", "Self-signed certificate chain in PEM format.")
		loginCmd.Short = strings.ReplaceAll(loginCmd.Short, ".", " (required for RBAC).")
		loginCmd.Long = strings.ReplaceAll(loginCmd.Long, ".", " (required for RBAC).")
		check(loginCmd.MarkFlagRequired("url")) // because https://confluent.cloud isn't an MDS endpoint
	}
	loginCmd.Flags().Bool("no-browser", false, "Do not open browser when authenticating via Single Sign-On.")
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

	anonClient := a.anonHTTPClientFactory(url, a.Config.Logger)
	email, password, err := a.getCCloudLoginCredentials(cmd, anonClient)
	if err != nil {
		return err
	}
	token, refreshToken, err := a.getCCloudTokens(cmd, anonClient, url, email, password)
	if err != nil {
		return err
	}

	state, err := a.getCCloudContextState(cmd, url, email, token)
	if err != nil {
		return err
	}

	err = a.addOrUpdateContext(state.Auth.User.Email, url, state, "")
	if err != nil {
		return err
	}

	err = a.saveLogin(cmd, email, password, refreshToken)
	if err != nil {
		return err
	}

	pcmd.Printf(cmd, errors.LoggedInAsMsg, email)
	pcmd.Printf(cmd, errors.LoggedInUsingEnvMsg, state.Auth.Account.Id, state.Auth.Account.Name)
	return err
}

func (a *loginCommand) getCCloudLoginCredentials(cmd *cobra.Command, client *ccloud.Client) (string, string, error) {
	email, password := a.getEnvVarCredentials(cmd, pauth.CCloudEmailEnvVar, pauth.CCloudPasswordEnvVar)
	var err error
	if len(email) == 0 {
		email, password, err = a.GetNetrcCredentials(cmd)
		if err != nil {
			return "", "", err
		}
	}
	if len(email) == 0 {
		email, password = a.getPromptedCredentials(cmd, "Email", client)
	}
	return email, password, nil
}

func (a *loginCommand) getCCloudTokens(cmd *cobra.Command, client *ccloud.Client, url string, email string, password string) (string, string, error) {
	noBrowser, err := cmd.Flags().GetBool("no-browser")
	if err != nil {
		return "", "", err
	}
	token, refreshToken, err := pauth.GetCCloudAuthToken(client, url, email, password, noBrowser, a.Logger)
	if err != nil {
		err = errors.CatchEmailNotFoundError(err, email)
		return "", "", err
	}
	return token, refreshToken, nil
}

func (a *loginCommand) getCCloudContextState(cmd *cobra.Command, url string, email string, token string) (*v2.ContextState, error) {
	ctxName := generateContextName(email, url)
	user, err := a.getCCloudUser(cmd, url, token)
	if err != nil {
		return nil, err
	}
	var state *v2.ContextState
	ctx, err := a.Config.FindContext(ctxName)
	if err == nil {
		state = ctx.State
	} else {
		state = new(v2.ContextState)
	}
	state.AuthToken = token

	if state.Auth == nil {
		state.Auth = &v1.AuthConfig{}
	}

	// Always overwrite the user, organization, and list of accounts when logging in -- but don't necessarily
	// overwrite `Account` (current/active environment) since we want that to be remembered
	// between CLI sessions.
	state.Auth.User = user.User
	state.Auth.Accounts = user.Accounts
	state.Auth.Organization = user.Organization

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

	return state, nil
}

func (a *loginCommand) getCCloudUser(cmd *cobra.Command, url string, token string) (*orgv1.GetUserReply, error) {
	client := a.jwtHTTPClientFactory(context.Background(), token, url, a.Config.Logger)
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return nil, err
	}
	if len(user.Accounts) == 0 {
		return nil, errors.Errorf(errors.NoEnvironmentFoundErrorMsg)
	}
	return user, nil
}

func (a *loginCommand) loginMDS(cmd *cobra.Command, _ []string) error {
	url, err := a.getURL(cmd)
	if err != nil {
		return err
	}

	caCertPath, err := cmd.Flags().GetString("ca-cert-path")
	if err != nil {
		return err
	}

	mdsClient, err := a.getMDSClient(cmd, url, caCertPath)
	if err != nil {
		return err
	}

	username, password, err := a.getConfluentLoginCredentials(cmd)
	if err != nil {
		return err
	}

	authToken, err := pauth.GetConfluentAuthToken(mdsClient, username, password)
	if err != nil {
		return err
	}

	state := &v2.ContextState{
		Auth:      nil,
		AuthToken: authToken,
	}

	err = a.addOrUpdateContext(username, url, state, caCertPath)
	if err != nil {
		return err
	}

	err = a.saveLogin(cmd, username, password, "")
	if err != nil {
		return err
	}

	pcmd.Printf(cmd, errors.LoggedInAsMsg, username)
	return nil
}

func (a *loginCommand) getMDSClient(cmd *cobra.Command, url string, caCertPath string) (*mds.APIClient, error) {
	ctx, err := a.getContext(cmd)
	if err != nil {
		return nil, err
	}
	caCertPathFlagChanged := cmd.Flags().Changed("ca-cert-path")
	mdsClient, err := a.MDSClientManager.GetMDSClient(ctx, caCertPath, caCertPathFlagChanged, url, a.Logger)
	if err != nil {
		return nil, err
	}
	return mdsClient, nil
}

func (a *loginCommand) getContext(cmd *cobra.Command) (*v3.Context, error) {
	dynamicContext, err := a.Config.Context(cmd)
	if err != nil {
		return nil, err
	}
	var ctx *v3.Context
	if dynamicContext != nil {
		ctx = dynamicContext.Context
	}
	return ctx, nil
}

func (a *loginCommand) getConfluentLoginCredentials(cmd *cobra.Command) (string, string, error) {
	username, password := a.getEnvVarCredentials(cmd, pauth.ConfluentUsernameEnvVar, pauth.ConfluentPasswordEnvVar)
	var err error
	if len(username) == 0 {
		username, password, err = a.GetNetrcCredentials(cmd)
		if err != nil {
			return "", "", err
		}
	}
	if len(username) == 0 {
		username, password = a.getPromptedCredentials(cmd, "Username", nil)
	}
	return username, password, nil
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
		pcmd.ErrPrintf(cmd, errors.UsingLoginURLDefaults, errMsg)
	}
	return url, nil
}

func (a *loginCommand) addOrUpdateContext(username string, url string, state *v2.ContextState, caCertPath string) error {
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
	err := a.Config.SavePlatform(platform)
	if err != nil {
		return err
	}
	err = a.Config.SaveCredential(credential)
	if err != nil {
		return err
	}
	if ctx, ok := a.Config.Contexts[ctxName]; ok {
		a.Config.ContextStates[ctxName] = state
		ctx.State = state
	} else {
		err = a.Config.AddContext(ctxName, platform.Name, credential.Name, map[string]*v1.KafkaClusterConfig{},
			"", nil, state)
	}
	if err != nil {
		return err
	}
	err = a.Config.SetContext(ctxName)
	if err != nil {
		return err
	}
	return nil
}

func (a *loginCommand) saveLogin(cmd *cobra.Command, email, password, refreshToken string) error {
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
	return nil
}

func (a *loginCommand) saveToNetrc(cmd *cobra.Command, username, password, refreshToken string) error {
	// sso if refresh token is empty
	var err error
	if refreshToken == "" {
		err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, false, a.Config.Config.Context().Name, username, password)
	} else {
		err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, true, a.Config.Config.Context().Name, username, refreshToken)
	}
	if err != nil {
		return err
	}
	pcmd.ErrPrintf(cmd, errors.WroteCredentialsToNetrcMsg, a.netrcHandler.GetFileName())
	return nil
}

func (a *loginCommand) GetNetrcCredentials(cmd *cobra.Command) (string, string, error) {
	return "", "", nil
}

func (a *loginCommand) getEnvVarCredentials(cmd *cobra.Command, userEnvVar string, passwordEnvVar string) (string, string) {
	user := os.Getenv(userEnvVar)
	if len(user) == 0 {
		return "", ""
	}
	password := os.Getenv(passwordEnvVar)
	if len(passwordEnvVar) == 0 {
		return "", ""
	}
	return user, password
}

// For Confluent login command, pass nil as ccloudClient argument
func (a *loginCommand) getPromptedCredentials(cmd *cobra.Command, userField string, ccloudClient *ccloud.Client) (string, string) {
	username := a.promptForUser(cmd, userField)
	var password string
	if ccloudClient != nil && isSSOUser(username, ccloudClient) {
		a.Logger.Trace("User is SSO-enabled so won't prompt for password")
	} else {
		password = a.promptForPassword(cmd)
	}
	return username, password
}

func (a *loginCommand) promptForUser(cmd *cobra.Command, userField string) string {
	pcmd.Println(cmd, "Enter your Confluent credentials:")
	f := form.New(form.Field{ID: userField, Prompt: userField})
	if err := f.Prompt(cmd, a.prompt); err != nil {
		return ""
	}
	return f.Responses[userField].(string)
}

func (a *loginCommand) promptForPassword(cmd *cobra.Command) string {
	passwordField := "Password"
	f := form.New(form.Field{ID: passwordField, Prompt: passwordField, IsHidden: true})
	if err := f.Prompt(cmd, a.prompt); err != nil {
		return ""
	}
	return f.Responses[passwordField].(string)
}

func generateContextName(username string, url string) string {
	return fmt.Sprintf("login-%s-%s", username, url)
}

func generateCredentialName(username string) string {
	return fmt.Sprintf("username-%s", username)
}

func isSSOUser(email string, cloudClient *ccloud.Client) bool {
	userSSO, err := cloudClient.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
	// Fine to ignore non-nil err for this request: e.g. what if this fails due to invalid/malicious
	// email, we want to silently continue and give the illusion of password prompt.
	// If Auth0ConnectionName is blank ("local" user) still prompt for password
	if err == nil && userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
		return true
	}
	return false
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
