package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"

	mds "github.com/confluentinc/mds-sdk-go"
)

type commands struct {
	Commands  []*cobra.Command
	config    *config.Config
	mdsClient *mds.APIClient
	Logger    *log.Logger
	// for testing
	prompt                pcmd.Prompt
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client
	jwtHTTPClientFactory  func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client
}

// New returns a list of auth-related Cobra commands.
func New(prerunner pcmd.PreRunner, config *config.Config, logger *log.Logger, mdsClient *mds.APIClient) []*cobra.Command {
	var defaultAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClient(baseURL, ccloud.BaseClient, logger)
	}
	var defaultJwtHTTPClientFactory = func(ctx context.Context, jwt string, baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClientWithJWT(ctx, jwt, baseURL, logger)
	}
	return newCommands(prerunner, config, logger, mdsClient, pcmd.NewPrompt(os.Stdin),
		defaultAnonHTTPClientFactory, defaultJwtHTTPClientFactory,
	).Commands
}

func newCommands(prerunner pcmd.PreRunner, config *config.Config, log *log.Logger, mdsClient *mds.APIClient, prompt pcmd.Prompt,
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client,
	jwtHTTPClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client,
) *commands {
	cmd := &commands{
		config:                config,
		mdsClient:             mdsClient,
		Logger:                log,
		prompt:                prompt,
		anonHTTPClientFactory: anonHTTPClientFactory,
		jwtHTTPClientFactory:  jwtHTTPClientFactory,
	}
	cmd.init(prerunner)
	return cmd
}

func (a *commands) init(prerunner pcmd.PreRunner) {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: fmt.Sprintf("Login to %s.", a.config.APIName()),
		Long:  fmt.Sprintf("Login to %s.", a.config.APIName()),
		Args:  cobra.NoArgs,
	}
	if a.config.CLIName == "ccloud" {
		loginCmd.RunE = a.login
		loginCmd.Flags().String("url", "https://confluent.cloud", "Confluent Control Plane URL.")
	} else {
		loginCmd.RunE = a.loginMDS
		loginCmd.Flags().String("url", "", "Confluent Control Plane URL.")
		check(loginCmd.MarkFlagRequired("url")) // because https://confluent.cloud isn't an MDS endpoint
	}
	loginCmd.Flags().SortFlags = false
	loginCmd.PersistentPreRunE = prerunner.Anonymous()
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: fmt.Sprintf("Logout of %s.", a.config.APIName()),
		Long:  fmt.Sprintf("Logout of %s.", a.config.APIName()),

		RunE: a.logout,
		Args: cobra.NoArgs,
	}
	logoutCmd.PersistentPreRunE = prerunner.Anonymous()
	a.Commands = []*cobra.Command{loginCmd, logoutCmd}
}

func (a *commands) login(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	a.config.AuthURL = url
	email, password, err := a.credentials(cmd, "Email")
	if err != nil {
		return err
	}

	client := a.anonHTTPClientFactory(a.config.AuthURL, a.config.Logger)

	token, err := client.Auth.Login(context.Background(), email, password)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	a.config.AuthToken = token

	client = a.jwtHTTPClientFactory(context.Background(), a.config.AuthToken, a.config.AuthURL, a.config.Logger)
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if len(user.Accounts) == 0 {
		return errors.HandleCommon(errors.New("No environments found for authenticated user!"), cmd)
	}

	// If no auth config exists, initialize it
	if a.config.Auth == nil {
		a.config.Auth = &config.AuthConfig{}
	}

	// Always overwrite the user and list of accounts when logging in -- but don't necessarily
	// overwrite `Account` (current/active environment) since we want that to be remembered
	// between CLI sessions.
	a.config.Auth.User = user.User
	a.config.Auth.Accounts = user.Accounts

	// Default to 0th environment if no suitable environment is already configured
	hasGoodEnv := false
	if a.config.Auth.Account != nil {
		for _, acc := range a.config.Auth.Accounts {
			if acc.Id == a.config.Auth.Account.Id {
				hasGoodEnv = true
			}
		}
	}
	if !hasGoodEnv {
		a.config.Auth.Account = a.config.Auth.Accounts[0]
	}

	a.createOrUpdateContext(a.config.Auth)

	err = a.config.Save()
	if err != nil {
		return errors.Wrap(err, "Unable to save user authentication.")
	}
	pcmd.Println(cmd, "Logged in as", email)
	pcmd.Print(cmd, "Using environment ", a.config.Auth.Account.Id,
		" (\"", a.config.Auth.Account.Name, "\")\n")
	return err
}

func (a *commands) loginMDS(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	a.config.AuthURL = url
	a.mdsClient.ChangeBasePath(a.config.AuthURL)
	email, password, err := a.credentials(cmd, "Username")
	if err != nil {
		return err
	}

	basicContext := context.WithValue(context.Background(), mds.ContextBasicAuth, mds.BasicAuth{UserName: email, Password: password})
	resp, _, err := a.mdsClient.TokensAuthenticationApi.GetToken(basicContext, "")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	a.config.AuthToken = resp.AuthToken

	err = a.config.Save()
	if err != nil {
		return errors.Wrap(err, "Unable to save user authentication.")
	}

	pcmd.Println(cmd, "Logged in as", email)

	return err
}

func (a *commands) logout(cmd *cobra.Command, args []string) error {
	a.config.AuthToken = ""
	a.config.Auth = nil
	err := a.config.Save()
	if err != nil {
		return errors.Wrap(err, "Unable to delete user auth")
	}
	pcmd.Println(cmd, "You are now logged out")
	return nil
}

func (a *commands) credentials(cmd *cobra.Command, userField string) (string, string, error) {
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
		email = emailFromPrompt
	}

	a.Logger.Trace("Successfully obtained email")

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

	return strings.TrimSpace(email), password, nil
}

func (a *commands) createOrUpdateContext(user *config.AuthConfig) {
	name := fmt.Sprintf("login-%s-%s", user.User.Email, a.config.AuthURL)
	if _, ok := a.config.Platforms[name]; !ok {
		a.config.Platforms[name] = &config.Platform{
			Server: a.config.AuthURL,
		}
	}
	if _, ok := a.config.Credentials[name]; !ok {
		a.config.Credentials[name] = &config.Credential{
			Username: user.User.Email,
			// don't save password if they entered it interactively
		}
	}
	if _, ok := a.config.Contexts[name]; !ok {
		a.config.Contexts[name] = &config.Context{
			Platform:      name,
			Credential:    name,
			KafkaClusters: map[string]*config.KafkaClusterConfig{},
		}
	}
	a.config.CurrentContext = name
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
