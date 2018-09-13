package auth

import (
	"context"
	"testing"

	"github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/command"
	chttp "github.com/confluentinc/cli/http"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)

	prompt := prompt("cody@confluent.io", "iambatman")
	auth := &mock.Auth{
		LoginFunc: func(username string, password string) (string, error) {
			return "y0ur.jwt.T0kEn", nil
		},
		UserFunc: func() (*shared.AuthConfig, error) {
			return &shared.AuthConfig{
				User: &orgv1.User{
					Id:        23,
					Email:     "cody@confluent.io",
					FirstName: "Cody",
				},
			}, nil
		},
	}
	cmds, config := newAuthCommand(prompt, auth, req)

	output, err := command.ExecuteCommand(cmds[0])
	req.NoError(err)
	req.Contains(output, "Logged in as cody@confluent.io")

	req.Equal("y0ur.jwt.T0kEn", config.AuthToken)
	req.Equal(&orgv1.User{Id: 23, Email: "cody@confluent.io", FirstName: "Cody"}, config.Auth.User)

	config = shared.NewConfig()
	config.Load()
	name := "login-cody@confluent.io-https://confluent.cloud"
	req.Contains(config.Platforms, name)
	req.Equal("https://confluent.cloud", config.Platforms[name].Server)
	req.Contains(config.Credentials, name)
	req.Equal("cody@confluent.io", config.Credentials[name].Username)
	req.Contains(config.Contexts, name)
	req.Equal(name, config.Contexts[name].Platform)
	req.Equal(name, config.Contexts[name].Credential)
}

func TestLoginFail(t *testing.T) {
	req := require.New(t)

	prompt := prompt("cody@confluent.io", "iamrobin")
	auth := &mock.Auth{
		LoginFunc: func(username string, password string) (string, error) {
			return "", shared.ErrIncorrectAuth
		},
	}
	cmds, _ := newAuthCommand(prompt, auth, req)

	output, err := command.ExecuteCommand(cmds[0])
	req.NoError(err)
	req.Contains(output, "You have entered an incorrect username or password.")
}

func TestLogout(t *testing.T) {
	req := require.New(t)

	prompt := prompt("cody@confluent.io", "iamrobin")
	auth := &mock.Auth{}
	cmds, config := newAuthCommand(prompt, auth, req)

	config.AuthToken = "some.token.here"
	config.Auth = &shared.AuthConfig{User: &orgv1.User{Id: 23}}
	config.Save()

	output, err := command.ExecuteCommand(cmds[1])
	req.NoError(err)
	req.Contains(output, "You are now logged out")

	config = shared.NewConfig()
	config.Load()
	req.Empty(config.AuthToken)
	req.Empty(config.Auth)
}

func prompt(username, password string) *mock.Prompt {
	return &mock.Prompt{
		Strings:   []string{username},
		Passwords: []string{password},
	}
}

func newAuthCommand(prompt command.Prompt, auth *mock.Auth, req *require.Assertions) ([]*cobra.Command, *shared.Config) {
	var mockAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *chttp.Client {
		req.Equal("https://confluent.cloud", baseURL)
		return &chttp.Client{Auth: auth}
	}
	var mockJwtHTTPClientFactory = func(ctx context.Context, jwt, baseURL string, logger *log.Logger) *chttp.Client {
		return &chttp.Client{Auth: auth}
	}
	config := shared.NewConfig()
	return newCommands(config, prompt, mockAnonHTTPClientFactory, mockJwtHTTPClientFactory), config
}
