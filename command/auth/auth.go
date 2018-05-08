package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/confluentinc/cli/command/common"
	chttp "github.com/confluentinc/cli/http"
	"github.com/confluentinc/cli/shared"
)

type Authentication struct {
	Commands  []*cobra.Command
	config    *shared.Config
}

func New(config *shared.Config) []*cobra.Command {
	cmd := &Authentication{config: config}
	cmd.init()
	return cmd.Commands
}

func (a *Authentication) init() {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to a Confluent Control Plane.",
		RunE:  a.login,
	}
	loginCmd.Flags().StringVar(&a.config.AuthURL, "url", "https://confluent.cloud", "Confluent Control Plane URL")
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout of a Confluent Control Plane.",
		RunE:  a.logout,
	}
	a.Commands = []*cobra.Command{loginCmd, logoutCmd}
}

func (a *Authentication) login(cmd *cobra.Command, args []string) error {
	email, password, err := credentials()
	if err != nil {
		return err
	}

	client := chttp.NewClient(chttp.BaseClient, a.config.AuthURL, a.config.Logger)
	token, err := client.Auth.Login(email, password)
	if err != nil {
		err := shared.ConvertAPIError(err)
		if err == shared.ErrUnauthorized { // special case for login failure
			err = shared.ErrIncorrectAuth
		}
		return common.HandleError(err)
	}
	a.config.AuthToken = token

	client = chttp.NewClientWithJWT(context.Background(), a.config.AuthToken, a.config.AuthURL, a.config.Logger)
	user, err := client.Auth.User()
	if err != nil {
		return common.HandleError(shared.ConvertAPIError(err))
	}
	a.config.Auth = user

	err = a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to save user auth")
	}
	fmt.Println("Logged in as", email)
	return nil
}

func (a *Authentication) logout(cmd *cobra.Command, args []string) error {
	a.config.AuthToken = ""
	a.config.Auth = nil
	err := a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to delete user auth")
	}
	fmt.Println("You are now logged out")
	return nil
}

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter your Confluent Cloud credentials:")

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	fmt.Println()
	if err != nil {
		return "", "", err
	}
	password := string(bytePassword)

	return strings.TrimSpace(email), strings.TrimSpace(password), nil
}
