package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
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
	Logger          *log.Logger
	analyticsClient analytics.Client
	// for testing
	MDSClientManager      pauth.MDSClientManager
	prompt                pcmd.Prompt
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client
	jwtHTTPClientFactory  func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client
	netrcHandler          *pauth.NetrcHandler
}

func NewLoginCommand(cliName string, prerunner pcmd.PreRunner, log *log.Logger, prompt pcmd.Prompt,
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client,
	jwtHTTPClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client,
	mdsClientManager pauth.MDSClientManager, analyticsClient analytics.Client, netrcHandler *pauth.NetrcHandler) *loginCommand {
	cmd := &loginCommand{
		Logger:                log,
		prompt:                prompt,
		analyticsClient:       analyticsClient,
		anonHTTPClientFactory: anonHTTPClientFactory,
		jwtHTTPClientFactory:  jwtHTTPClientFactory,
		MDSClientManager:      mdsClientManager,
		netrcHandler:          netrcHandler,
	}
	cmd.init(cliName, prerunner)
	return cmd
}

func (a *loginCommand) init(cliName string, prerunner pcmd.PreRunner) {
	remoteAPIName := getRemoteAPIName(cliName)
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: fmt.Sprintf("Log in to %s.", remoteAPIName),
		Args:  cobra.NoArgs,
		PersistentPreRunE: pcmd.NewCLIPreRunnerE(func(cmd *cobra.Command, args []string) error {
			a.analyticsClient.SetCommandType(analytics.Login)
			return a.CLICommand.PersistentPreRunE(cmd, args)
		}),
	}
	if cliName == "ccloud" {
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
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	noBrowser, err := cmd.Flags().GetBool("no-browser")
	if err != nil {
		return err
	}
	a.Config.NoBrowser = noBrowser

	client := a.anonHTTPClientFactory(url, a.Config.Logger)
	email, password, err := a.credentials(cmd, "Email", client)
	if err != nil {
		return err
	}

	token, refreshToken, err := pauth.GetCCloudAuthToken(client, url, email, password, noBrowser, a.Logger)
	if err != nil {
		err = errors.CatchEmailNotFoundError(err, email)
		return err
	}

	client = a.jwtHTTPClientFactory(context.Background(), token, url, a.Config.Logger)
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return err
	}

	if len(user.Accounts) == 0 {
		return errors.Errorf(errors.NoEnvironmentFoundErrorMsg)
	}
	username := user.User.Email
	name := generateContextName(username, url)
	var state *v2.ContextState
	ctx, err := a.Config.FindContext(name)
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

	err = a.addOrUpdateContext(state.Auth.User.Email, url, state, "")
	if err != nil {
		return err
	}
	err = a.Config.Save()
	if err != nil {
		return errors.Wrap(err, errors.UnableToSaveUserAuthErrorMsg)
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

	pcmd.Printf(cmd, errors.LoggedInAsMsg, email)
	pcmd.Printf(cmd, errors.LoggedInUsingEnvMsg, state.Auth.Account.Id, state.Auth.Account.Name)
	return err
}

func (a *loginCommand) loginMDS(cmd *cobra.Command, _ []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	email, password, err := a.credentials(cmd, "Username", nil)
	if err != nil {
		return err
	}
	dynamicContext, err := a.Config.Context(cmd)
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
		return err
	}
	mdsClient, err := a.MDSClientManager.GetMDSClient(ctx, caCertPath, flagChanged, url, a.Logger)
	if err != nil {
		return err
	}
	authToken, err := pauth.GetConfluentAuthToken(mdsClient, email, password)
	if err != nil {
		return err
	}
	if err != nil {
		return err
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
	pcmd.Printf(cmd, errors.LoggedInAsMsg, email)
	return nil
}

func (a *loginCommand) credentials(cmd *cobra.Command, userField string, cloudClient *ccloud.Client) (string, string, error) {
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
		f := form.New(form.Field{ID: "email", Prompt: userField})
		if err := f.Prompt(cmd, a.prompt); err != nil {
			return "", "", err
		}
		email = f.Responses["email"].(string)
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
		f := form.New(form.Field{ID: "password", Prompt: "Password", IsHidden: true})
		if err := f.Prompt(cmd, a.prompt); err != nil {
			return "", "", err
		}
		password = f.Responses["password"].(string)
	}
	a.Logger.Trace("Successfully obtained password")

	return email, password, nil
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

func (a *loginCommand) saveToNetrc(cmd *cobra.Command, email, password, refreshToken string) error {
	// sso if refresh token is empty
	var err error
	if refreshToken == "" {
		err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, false, a.Config.Config.Context().Name, email, password)
	} else {
		err = a.netrcHandler.WriteNetrcCredentials(a.Config.CLIName, true, a.Config.Config.Context().Name, email, refreshToken)
	}
	if err != nil {
		return err
	}
	pcmd.ErrPrintf(cmd, errors.WroteCredentialsToNetrcMsg, a.netrcHandler.FileName)
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
