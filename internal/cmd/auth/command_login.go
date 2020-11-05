package auth

import (
	"context"
	"fmt"
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
	MDSClientManager      pauth.MDSClientManager
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client
	jwtHTTPClientFactory  func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client
	netrcHandler          netrc.NetrcHandler
	loginTokenHandler     pauth.LoginTokenHandler
}

func NewLoginCommand(cliName string, prerunner pcmd.PreRunner, log *log.Logger,
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client,
	jwtHTTPClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client,
	mdsClientManager pauth.MDSClientManager, analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler,
	loginTokenHandler pauth.LoginTokenHandler) *loginCommand {
	cmd := &loginCommand{
		cliName:               cliName,
		Logger:                log,
		analyticsClient:       analyticsClient,
		anonHTTPClientFactory: anonHTTPClientFactory,
		jwtHTTPClientFactory:  jwtHTTPClientFactory,
		MDSClientManager:      mdsClientManager,
		netrcHandler:          netrcHandler,
		loginTokenHandler:     loginTokenHandler,
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

	token, creds, err := a.getCCloudTokenAndCredentials(cmd, url)
	if err != nil {
		return err
	}

	state, err := a.getCCloudContextState(cmd, url, creds.Username, token)
	if err != nil {
		return err
	}

	err = a.addOrUpdateContext(creds.Username, url, state, "")
	if err != nil {
		return err
	}

	err = a.saveLoginToNetrc(cmd, creds)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.LoggedInAsMsg, creds.Username)
	utils.Printf(cmd, errors.LoggedInUsingEnvMsg, state.Auth.Account.Id, state.Auth.Account.Name)
	return err
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (a *loginCommand) getCCloudTokenAndCredentials(cmd *cobra.Command, url string) (string, *pauth.Credentials, error) {
	client := a.anonHTTPClientFactory(url, a.Logger)

	token, creds, err := a.loginTokenHandler.GetCCloudTokenAndCredentialsFromEnvVar(cmd, client)
	if err == nil && len(token) > 0 {
		return token, creds, nil
	}

	token, creds, err = a.loginTokenHandler.GetCCloudTokenAndCredentialsFromNetrc(cmd, client, url, netrc.GetMatchingNetrcMachineParams{
		CLIName: a.cliName,
		URL:     url,
	})
	if err == nil && len(token) > 0 {
		return token, creds, nil
	}

	return a.loginTokenHandler.GetCCloudTokenAndCredentialsFromPrompt(cmd, client, url)
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

	token, creds, err := a.getConfluentTokenAndCredentials(cmd, url, caCertPath)
	if err != nil {
		return err
	}

	state := &v2.ContextState{
		Auth:      nil,
		AuthToken: token,
	}

	err = a.addOrUpdateContext(creds.Username, url, state, caCertPath)
	if err != nil {
		return err
	}

	err = a.saveLoginToNetrc(cmd, creds)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.LoggedInAsMsg, creds.Username)
	return nil
}

// Order of precedence: env vars > netrc > prompt
// i.e. if login credentials found in env vars then acquire token using env vars and skip checking for credentials else where
func (a *loginCommand) getConfluentTokenAndCredentials(cmd *cobra.Command, url string, caCertPath string) (string, *pauth.Credentials, error) {
	client, err := a.getMDSClient(cmd, url, caCertPath)
	if err != nil {
		return "", nil, err
	}

	token, creds, err := a.loginTokenHandler.GetConfluentTokenAndCredentialsFromEnvVar(cmd, client)
	if err == nil && len(token) > 0 {
		return token, creds, nil
	}

	token, creds, err = a.loginTokenHandler.GetConfluentTokenAndCredentialsFromNetrc(cmd, client, netrc.GetMatchingNetrcMachineParams{
		CLIName: a.cliName,
		URL:     url,
	})
	if err == nil && len(token) > 0 {
		return token, creds, nil
	}

	return a.loginTokenHandler.GetConfluentTokenAndCredentialsFromPrompt(cmd, client)
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

func (a *loginCommand) saveLoginToNetrc(cmd *cobra.Command, creds *pauth.Credentials) error {
	saveToNetrc, err := cmd.Flags().GetBool("save")
	if err != nil {
		return err
	}
	if saveToNetrc {
		var err error
		if creds.RefreshToken == "" {
			err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, false, a.Config.Config.Context().Name, creds.Username, creds.Password)
		} else {
			err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, true, a.Config.Config.Context().Name, creds.Username, creds.RefreshToken)
		}
		if err != nil {
			return err
		}
		utils.ErrPrintf(cmd, errors.WroteCredentialsToNetrcMsg, a.netrcHandler.GetFileName())
	}
	return nil
}

func generateContextName(username string, url string) string {
	return fmt.Sprintf("login-%s-%s", username, url)
}

func generateCredentialName(username string) string {
	return fmt.Sprintf("username-%s", username)
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
