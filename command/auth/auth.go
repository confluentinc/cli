package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

type commands struct {
	Commands []*cobra.Command
	config   *shared.Config
	// for testing
	prompt                command.Prompt
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *chttp.Client
	jwtHTTPClientFactory  func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *chttp.Client
}

// New returns a list of auth-related Cobra commands.
func New(config *shared.Config) []*cobra.Command {
	var defaultAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *chttp.Client {
		return chttp.NewClient(baseURL, chttp.BaseClient, logger)
	}
	var defaultJwtHTTPClientFactory = func(ctx context.Context, jwt string, baseURL string, logger *log.Logger) *chttp.Client {
		return chttp.NewClientWithJWT(ctx, jwt, baseURL, logger)
	}
	return newCommands(config, command.NewTerminalPrompt(os.Stdin), defaultAnonHTTPClientFactory, defaultJwtHTTPClientFactory).Commands
}

func newCommands(config *shared.Config, prompt command.Prompt,
	anonHTTPClientFactory func(baseURL string, logger *log.Logger) *chttp.Client,
	jwtHTTPClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *chttp.Client,
) *commands {
	cmd := &commands{config: config, prompt: prompt, anonHTTPClientFactory: anonHTTPClientFactory, jwtHTTPClientFactory: jwtHTTPClientFactory}
	cmd.init()
	return cmd
}

func (a *commands) init() {
	var preRun = func(cmd *cobra.Command, args []string) error {
		if err := common.SetLoggingVerbosity(cmd, a.config.Logger); err != nil {
			return common.HandleError(err, cmd)
		}
		a.prompt.SetOutput(cmd.OutOrStderr())
		return nil
	}
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Confluent Cloud",
		RunE:  a.login,
	}
	loginCmd.Flags().String("url", "https://confluent.cloud", "Confluent Control Plane URL")
	loginCmd.PersistentPreRunE = preRun
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout of Confluent Cloud",
		RunE:  a.logout,
	}
	logoutCmd.PersistentPreRunE = preRun
	a.Commands = []*cobra.Command{loginCmd, logoutCmd}
}

func (a *commands) login(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	a.config.AuthURL = url
	email, password, err := a.credentials()
	if err != nil {
		return err
	}

	client := a.anonHTTPClientFactory(a.config.AuthURL, a.config.Logger)

	token, err := client.Auth.Login(context.Background(), email, password)
	if err != nil {
		err = shared.ConvertAPIError(err)
		if err == shared.ErrUnauthorized { // special case for login failure
			err = shared.ErrIncorrectAuth
		}
		return common.HandleError(err, cmd)
	}
	a.config.AuthToken = token

	client = a.jwtHTTPClientFactory(context.Background(), a.config.AuthToken, a.config.AuthURL, a.config.Logger)
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return common.HandleError(shared.ConvertAPIError(err), cmd)
	}
	a.config.Auth = &shared.AuthConfig{User: user.User, Account: user.Account}

	a.createOrUpdateContext(a.config.Auth)

	err = a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to save user auth")
	}
	_, err = a.prompt.Println("Logged in as", email)
	return err
}

func (a *commands) logout(cmd *cobra.Command, args []string) error {
	a.config.AuthToken = ""
	a.config.Auth = nil
	err := a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to delete user auth")
	}
	_, err = a.prompt.Println("You are now logged out")
	return err
}

func (a *commands) credentials() (string, string, error) {
	email := os.Getenv("XX_CCLOUD_EMAIL")
	password := os.Getenv("XX_CCLOUD_PASSWORD")
	if len(email) == 0 || len(password) == 0 {
		if _, err := a.prompt.Println("Enter your Confluent Cloud credentials:"); err != nil {
			return "", "", err
		}
	}
	if len(email) == 0 {
		a.prompt.Print("Email: ")
		emailFromPrompt, err := a.prompt.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		email = emailFromPrompt
	}

	if len(password) == 0 {
		a.prompt.Print("Password: ")
		bytePassword, err := a.prompt.ReadPassword(0)
		if err != nil {
			return "", "", err
		}
		_, err = a.prompt.Println()
		if err != nil {
			return "", "", err
		}
		password = string(bytePassword)
	}

	return strings.TrimSpace(email), password, nil
}

func (a *commands) createOrUpdateContext(user *shared.AuthConfig) {
	name := fmt.Sprintf("login-%s-%s", user.User.Email, a.config.AuthURL)
	if _, ok := a.config.Platforms[name]; !ok {
		a.config.Platforms[name] = &shared.Platform{
			Server: a.config.AuthURL,
		}
	}
	if _, ok := a.config.Credentials[name]; !ok {
		a.config.Credentials[name] = &shared.Credential{
			Username: user.User.Email,
			// don't save password if they entered it interactively
		}
	}
	if _, ok := a.config.Contexts[name]; !ok {
		a.config.Contexts[name] = &shared.Context{
			Platform:   name,
			Credential: name,
		}
	}
	a.config.CurrentContext = name
}
