package auth

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/ccloud-sdk-go"
	sdkMock "github.com/confluentinc/ccloud-sdk-go/mock"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
)

func TestCredentialsOverride(t *testing.T) {
	req := require.New(t)
	currentEmail := os.Getenv("XX_CCLOUD_EMAIL")
	currentPassword := os.Getenv("XX_CCLOUD_PASSWORD")

	os.Setenv("XX_CCLOUD_EMAIL", "test-email")
	os.Setenv("XX_CCLOUD_PASSWORD", "test-password")

	prompt := prompt("cody@confluent.io", "iambatman")
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, username string, password string) (string, error) {
			return "y0ur.jwt.T0kEn", nil
		},
		UserFunc: func(ctx context.Context) (*orgv1.GetUserReply, error) {
			return &orgv1.GetUserReply{
				User: &orgv1.User{
					Id:        23,
					Email:     "test-email",
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	cmds, cfg := newAuthCommand(prompt, auth, req)

	output, err := pcmd.ExecuteCommand(cmds.Commands[0])
	req.NoError(err)
	req.Contains(output, "Logged in as test-email")

	req.Equal("y0ur.jwt.T0kEn", cfg.AuthToken)
	req.Equal(&orgv1.User{Id: 23, Email: "test-email", FirstName: "Cody"}, cfg.Auth.User)

	os.Setenv("XX_CCLOUD_EMAIL", currentEmail)
	os.Setenv("XX_CCLOUD_PASSWORD", currentPassword)
}

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)

	prompt := prompt("cody@confluent.io", "iambatman")
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, username string, password string) (string, error) {
			return "y0ur.jwt.T0kEn", nil
		},
		UserFunc: func(ctx context.Context) (*orgv1.GetUserReply, error) {
			return &orgv1.GetUserReply{
				User: &orgv1.User{
					Id:        23,
					Email:     "cody@confluent.io",
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	cmds, cfg := newAuthCommand(prompt, auth, req)

	output, err := pcmd.ExecuteCommand(cmds.Commands[0])
	req.NoError(err)
	req.Contains(output, "Logged in as cody@confluent.io")

	req.Equal("y0ur.jwt.T0kEn", cfg.AuthToken)
	req.Equal(&orgv1.User{Id: 23, Email: "cody@confluent.io", FirstName: "Cody"}, cfg.Auth.User)

	cfg = config.New()
	req.NoError(cfg.Load())
	name := "login-cody@confluent.io-https://confluent.cloud"
	req.Contains(cfg.Platforms, name)
	req.Equal("https://confluent.cloud", cfg.Platforms[name].Server)
	req.Contains(cfg.Credentials, name)
	req.Equal("cody@confluent.io", cfg.Credentials[name].Username)
	req.Contains(cfg.Contexts, name)
	req.Equal(name, cfg.Contexts[name].Platform)
	req.Equal(name, cfg.Contexts[name].Credential)
}

func TestLoginFail(t *testing.T) {
	req := require.New(t)

	prompt := prompt("cody@confluent.io", "iamrobin")
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, username string, password string) (string, error) {
			return "", errors.ErrIncorrectAuth
		},
	}
	cmds, _ := newAuthCommand(prompt, auth, req)

	_, err := pcmd.ExecuteCommand(cmds.Commands[0])
	req.Contains(err.Error(), "You have entered an incorrect username or password.")
}

func TestLogout(t *testing.T) {
	req := require.New(t)

	prompt := prompt("cody@confluent.io", "iamrobin")
	auth := &sdkMock.Auth{}
	cmds, cfg := newAuthCommand(prompt, auth, req)

	cfg.AuthToken = "some.token.here"
	cfg.Auth = &config.AuthConfig{User: &orgv1.User{Id: 23}}
	req.NoError(cfg.Save())

	output, err := pcmd.ExecuteCommand(cmds.Commands[1])
	req.NoError(err)
	req.Contains(output, "You are now logged out")

	cfg = config.New()
	req.NoError(cfg.Load())
	req.Empty(cfg.AuthToken)
	req.Empty(cfg.Auth)
}

func Test_credentials_NoSpacesAroundEmail_ShouldSupportSpacesAtBeginOrEnd(t *testing.T) {
	req := require.New(t)

	prompt := prompt(" cody@confluent.io ", " iamrobin ")
	auth := &sdkMock.Auth{}
	cmds, _ := newAuthCommand(prompt, auth, req)

	user, pass, err := cmds.credentials(cmds.Commands[0])
	req.NoError(err)
	req.Equal("cody@confluent.io", user)
	req.Equal(" iamrobin ", pass)
}

func prompt(username, password string) *cliMock.Prompt {
	return &cliMock.Prompt{
		Strings:   []string{username},
		Passwords: []string{password},
	}
}

func newAuthCommand(prompt pcmd.Prompt, auth *sdkMock.Auth, req *require.Assertions) (*commands, *config.Config) {
	var mockAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *ccloud.Client {
		req.Equal("https://confluent.cloud", baseURL)
		return &ccloud.Client{Auth: auth}
	}
	var mockJwtHTTPClientFactory = func(ctx context.Context, jwt, baseURL string, logger *log.Logger) *ccloud.Client {
		return &ccloud.Client{Auth: auth}
	}
	cfg := config.New()
	cfg.Logger = log.New()
	commands := newCommands(&cliMock.Commander{}, cfg, prompt, mockAnonHTTPClientFactory, mockJwtHTTPClientFactory)
	for _, c := range commands.Commands {
		c.PersistentFlags().CountP("verbose", "v", "increase output verbosity")
	}
	return commands, cfg
}
