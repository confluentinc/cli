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
)

type commands struct {
	Commands []*cobra.Command
	config   *config.Config
	// for testing
	prompt                pcmd.Prompt
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client
	jwtHTTPClientFactory  func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client
}

// New returns a list of auth-related Cobra commands.
func New(prerunner pcmd.PreRunner, config *config.Config) []*cobra.Command {
	var defaultAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClient(baseURL, ccloud.BaseClient, logger)
	}
	var defaultJwtHTTPClientFactory = func(ctx context.Context, jwt string, baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClientWithJWT(ctx, jwt, baseURL, logger)
	}
	return newCommands(prerunner, config, pcmd.NewPrompt(os.Stdin),
		defaultAnonHTTPClientFactory, defaultJwtHTTPClientFactory,
	).Commands
}

func newCommands(prerunner pcmd.PreRunner, config *config.Config, prompt pcmd.Prompt,
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *ccloud.Client,
	jwtHTTPClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *ccloud.Client,
) *commands {
	cmd := &commands{
		config:                config,
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
		Short: "Login to Confluent Cloud",
		RunE:  a.login,
		Args:  cobra.NoArgs,
	}
	loginCmd.Flags().String("url", "https://confluent.cloud", "Confluent Control Plane URL")
	loginCmd.Flags().SortFlags = false
	loginCmd.PersistentPreRunE = prerunner.Anonymous()
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout of Confluent Cloud",
		RunE:  a.logout,
		Args:  cobra.NoArgs,
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
	email, password, err := a.credentials(cmd)
	if err != nil {
		return err
	}

	client := a.anonHTTPClientFactory(a.config.AuthURL, a.config.Logger)

	token, err := client.Auth.Login(context.Background(), email, password)
	if err != nil {
		err = errors.ConvertAPIError(err)
		if err == errors.ErrUnauthorized { // special case for login failure
			err = errors.ErrIncorrectAuth
		}
		return errors.HandleCommon(err, cmd)
	}
	a.config.AuthToken = token

	client = a.jwtHTTPClientFactory(context.Background(), a.config.AuthToken, a.config.AuthURL, a.config.Logger)
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return errors.HandleCommon(errors.ConvertAPIError(err), cmd)
	}

	if len(user.Accounts) == 0 {
		return errors.HandleCommon(errors.New("no environments found for authenticated user!"), cmd)
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
		return errors.Wrap(err, "unable to save user auth")
	}
	pcmd.Println(cmd, "Logged in as", email)
	pcmd.Print(cmd, "Using environment ", a.config.Auth.Account.Id,
		" (\"", a.config.Auth.Account.Name, "\")\n")
	return err
}

func (a *commands) logout(cmd *cobra.Command, args []string) error {
	a.config.AuthToken = ""
	a.config.Auth = nil
	err := a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to delete user auth")
	}
	pcmd.Println(cmd, "You are now logged out")
	return nil
}

func (a *commands) credentials(cmd *cobra.Command) (string, string, error) {
	email := os.Getenv("XX_CCLOUD_EMAIL")
	password := os.Getenv("XX_CCLOUD_PASSWORD")
	if len(email) == 0 || len(password) == 0 {
		pcmd.Println(cmd, "Enter your Confluent Cloud credentials:")
	}
	if len(email) == 0 {
		pcmd.Print(cmd, "Email: ")
		emailFromPrompt, err := a.prompt.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		email = emailFromPrompt
	}

	if len(password) == 0 {
		pcmd.Print(cmd, "Password: ")
		bytePassword, err := a.prompt.ReadPassword(0)
		if err != nil {
			return "", "", err
		}
		pcmd.Println(cmd)
		password = string(bytePassword)
	}

	return strings.TrimSpace(email), password, nil
}

func (a *commands) createOrUpdateContext(user *config.AuthConfig) {
	name := fmt.Sprintf("login-%s-%s", user.User.Email, a.config.AuthURL)
	if _, ok := a.config.Platforms[name]; !ok {
		a.config.Platforms[name] = &config.Platform{
			Server:        a.config.AuthURL,
			KafkaClusters: map[string]config.KafkaClusterConfig{},
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
			Platform:   name,
			Credential: name,
		}
	}
	a.config.CurrentContext = name
}
