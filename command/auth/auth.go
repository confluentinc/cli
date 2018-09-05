package auth

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/command/common"
	chttp "github.com/confluentinc/cli/http"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

type AnonHttpClientFactory func(baseURL string, logger *log.Logger) *chttp.Client
type JwtHttpClientFactory func(ctx context.Context, authToken string, baseURL string, logger *log.Logger) *chttp.Client

type commands struct {
	Commands []*cobra.Command
	config   *shared.Config
	// for testing
	Stdin                 command.Prompt
	anonHttpClientFactory AnonHttpClientFactory
	jwtHttpClientFactory  JwtHttpClientFactory
}

// New returns a list of auth-related Cobra commands.
func New(config *shared.Config) []*cobra.Command {
	var defaultAnonHttpClientFactory = func(baseURL string, logger *log.Logger) *chttp.Client {
		return chttp.NewClient(chttp.BaseClient, baseURL, logger)
	}
	var defaultJwtHttpClientFactory = func(ctx context.Context, jwt string, baseURL string, logger *log.Logger) *chttp.Client {
		return chttp.NewClientWithJWT(ctx, jwt, baseURL, logger)
	}
	return newForTesting(config, command.NewTerminalPrompt(os.Stdin), defaultAnonHttpClientFactory, defaultJwtHttpClientFactory)
}

func newForTesting(config *shared.Config, stdin command.Prompt, anonHttpClientFactory AnonHttpClientFactory, jwtHttpClientFactory JwtHttpClientFactory) []*cobra.Command {
	cmd := &commands{config: config, Stdin: stdin, anonHttpClientFactory: anonHttpClientFactory, jwtHttpClientFactory: jwtHttpClientFactory}
	cmd.init()
	return cmd.Commands
}

func (a *commands) init() {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to a Confluent Control Plane.",
		RunE:  a.login,
	}
	loginCmd.Flags().String("url", "https://confluent.cloud", "Confluent Control Plane URL")
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout of a Confluent Control Plane.",
		RunE:  a.logout,
	}
	a.Commands = []*cobra.Command{loginCmd, logoutCmd}
}

func (a *commands) login(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	a.config.AuthURL = url
	email, password, err := a.credentials(cmd.OutOrStderr())
	if err != nil {
		return err
	}

	client := a.anonHttpClientFactory(a.config.AuthURL, a.config.Logger)
	token, err := client.Auth.Login(email, password)
	if err != nil {
		err = shared.ConvertAPIError(err)
		if err == shared.ErrUnauthorized { // special case for login failure
			err = shared.ErrIncorrectAuth
		}
		return common.HandleError(err, cmd)
	}
	a.config.AuthToken = token

	client = a.jwtHttpClientFactory(context.Background(), a.config.AuthToken, a.config.AuthURL, a.config.Logger)
	user, err := client.Auth.User()
	if err != nil {
		return common.HandleError(shared.ConvertAPIError(err), cmd)
	}
	a.config.Auth = user

	a.createOrUpdateContext(user)

	err = a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to save user auth")
	}
	fmt.Fprintln(cmd.OutOrStderr(), "Logged in as", email)
	return nil
}

func (a *commands) logout(cmd *cobra.Command, args []string) error {
	a.config.AuthToken = ""
	a.config.Auth = nil
	err := a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to delete user auth")
	}
	fmt.Fprintln(cmd.OutOrStderr(), "You are now logged out")
	return nil
}

func (a *commands) credentials(out io.Writer) (string, string, error) {
	fmt.Fprintln(out, "Enter your Confluent Cloud credentials:")

	fmt.Fprint(out, "Email: ")
	email, err := a.Stdin.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Fprint(out, "Password: ")
	bytePassword, err := a.Stdin.ReadPassword(0)
	fmt.Fprintln(out)
	if err != nil {
		return "", "", err
	}
	password := string(bytePassword)

	return strings.TrimSpace(email), strings.TrimSpace(password), nil
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
