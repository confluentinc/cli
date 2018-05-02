package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	chttp "github.com/confluentinc/cli/http"
	"github.com/confluentinc/cli/shared"
	//"io/ioutil"
	"golang.org/x/oauth2"
	"context"
)

const (
	loginPath = "/api/sessions"
	mePath = "/api/me"
)

var (
	ErrUnauthorized = fmt.Errorf("unauthorized")
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
	username, password, err := credentials()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{"email": username, "password": password})
	if err != nil {
		return err
	}
	response, err := a.http().Post(a.config.AuthURL + loginPath, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case http.StatusNotFound:
		return ErrUnauthorized
	case http.StatusOK:
		var token string
		for _, cookie := range response.Cookies() {
			if cookie.Name == "auth_token" {
				token = cookie.Value
				break
			}
		}
		if token == "" {
			return ErrUnauthorized
		}
		a.config.AuthToken = token

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: a.config.AuthToken})
		tc := oauth2.NewClient(context.Background(), ts)
		s := sling.New().Client(tc).Base(a.config.AuthURL)

		me := &shared.AuthConfig{}
		confluentError := &chttp.ConfluentError{}
		_, err := s.New().Get("/api/me").Receive(me, confluentError)
		if err != nil {
			return errors.Wrap(err, "unable to fetch user info") // you just don't get me
		}
		a.config.Auth = me

		err = a.config.Save()
		if err != nil {
			return errors.Wrap(err, "unable to save user auth")
		}
		fmt.Println("Logged in as", username)
	}
	return nil
}

func (a *Authentication) logout(cmd *cobra.Command, args []string) error {
	a.config.AuthToken = ""
	err := a.config.Save()
	if err != nil {
		return errors.Wrap(err, "unable to delete auth token")
	}
	fmt.Println("You are now logged out")
	return nil
}

func (a *Authentication) http() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10,
	}
}

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter your Confluent Cloud credentials:")

	fmt.Print("Email: ")
	username, err := reader.ReadString('\n')
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

	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}
