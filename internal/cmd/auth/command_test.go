package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/confluentinc/cli/internal/pkg/errors"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	sdkMock "github.com/confluentinc/ccloud-sdk-go/mock"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	mdsMock "github.com/confluentinc/mds-sdk-go/mdsv1/mock"
	"github.com/stretchr/testify/require"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

func TestCredentialsOverride(t *testing.T) {
	req := require.New(t)
	currentEmail := os.Getenv("XX_CCLOUD_EMAIL")
	currentPassword := os.Getenv("XX_CCLOUD_PASSWORD")

	os.Setenv("XX_CCLOUD_EMAIL", "test-email")
	os.Setenv("XX_CCLOUD_PASSWORD", "test-password")

	prompt := prompt()
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
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
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: "test-email",
			}, nil
		},
	}
	loginCmd, cfg := newLoginCmd(prompt, auth, user, "ccloud", req)

	output, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.NoError(err)
	req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, "test-email"))
	ctx := cfg.Context()
	req.NotNil(ctx)

	req.Equal("y0ur.jwt.T0kEn", ctx.State.AuthToken)
	req.Equal(&orgv1.User{Id: 23, Email: "test-email", FirstName: "Cody"}, ctx.State.Auth.User)

	os.Setenv("XX_CCLOUD_EMAIL", currentEmail)
	os.Setenv("XX_CCLOUD_PASSWORD", currentPassword)
}

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)

	prompt := prompt()
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
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
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: "test-email",
			}, nil
		},
	}

	suite := []struct {
		cliName string
		args    []string
	}{
		{
			cliName: "ccloud",
			args:    []string{},
		},
		{
			cliName: "confluent",
			args: []string{
				"--url=http://localhost:8090",
			},
		},
	}

	for _, s := range suite {
		// Login to the CLI control plane
		loginCmd, cfg := newLoginCmd(prompt, auth, user, s.cliName, req)
		output, err := pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, "cody@confluent.io"))
		verifyLoggedInState(t, cfg, s.cliName)
	}
}

func TestLoginFail(t *testing.T) {
	req := require.New(t)

	prompt := prompt()
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return "", &ccloud.InvalidLoginError{}
		},
	}
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: "test-email",
			}, nil
		},
	}
	loginCmd, _ := newLoginCmd(prompt, auth, user, "ccloud", req)

	_, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.Contains(err.Error(), errors.InvalidLoginErrorMsg)
	errors.VerifyErrorAndSuggestions(req, err, errors.InvalidLoginErrorMsg, errors.CCloudInvalidLoginSuggestions)
}

func TestURLRequiredWithMDS(t *testing.T) {
	req := require.New(t)

	prompt := prompt()
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return "", &ccloud.InvalidLoginError{}
		},
	}
	loginCmd, _ := newLoginCmd(prompt, auth, nil, "confluent", req)

	_, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.Contains(err.Error(), "required flag(s) \"url\" not set")
}

func TestLogout(t *testing.T) {
	req := require.New(t)

	cfg := v3.AuthenticatedCloudConfigMock()
	contextName := cfg.Context().Name
	logoutCmd, cfg := newLogoutCmd("ccloud", cfg)
	output, err := pcmd.ExecuteCommand(logoutCmd.Command)
	req.NoError(err)
	req.Contains(output, errors.LoggedOutMsg)
	verifyLoggedOutState(t, cfg, contextName)
}

func Test_credentials_NoSpacesAroundEmail_ShouldSupportSpacesAtBeginOrEnd(t *testing.T) {
	req := require.New(t)

	prompt := prompt()
	auth := &sdkMock.Auth{}
	loginCmd, _ := newLoginCmd(prompt, auth, nil, "ccloud", req)

	user, pass, err := loginCmd.credentials(loginCmd.Command, "Email", nil)
	req.NoError(err)
	req.Equal("cody@confluent.io", user)
	req.Equal(" iamrobin ", pass)
}

func Test_SelfSignedCerts(t *testing.T) {
	req := require.New(t)
	mdsConfig := mds.NewConfiguration()
	mdsClient := mds.NewAPIClient(mdsConfig)
	cfg := v3.New(&config.Params{
		CLIName:    "confluent",
		MetricSink: nil,
		Logger:     log.New(),
	})
	prompt := prompt()
	prerunner := cliMock.NewPreRunnerMock(nil, nil, cfg)

	// Create a test certificate to be read in by the command
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject:      pkix.Name{Organization: []string{"testorg"}},
	}
	priv, err := rsa.GenerateKey(rand.Reader, 512)
	req.NoError(err, "Couldn't generate private key")
	certBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)
	req.NoError(err, "Couldn't generate certificate from private key")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certReader := bytes.NewReader(pemBytes)

	cert, err := x509.ParseCertificate(certBytes)
	req.NoError(err, "Couldn't reparse certificate")
	expectedSubject := cert.RawSubject
	mdsClient.TokensAndAuthenticationApi = &mdsMock.TokensAndAuthenticationApi{
		GetTokenFunc: func(ctx context.Context) (mds.AuthenticationResponse, *http.Response, error) {
			req.NotEqual(http.DefaultClient, mdsClient)
			transport, ok := mdsClient.GetConfig().HTTPClient.Transport.(*http.Transport)
			req.True(ok)
			req.NotEqual(http.DefaultTransport, transport)
			found := false
			for _, actualSubject := range transport.TLSClientConfig.RootCAs.Subjects() {
				if bytes.Equal(expectedSubject, actualSubject) {
					found = true
					break
				}
			}
			req.True(found, "Certificate not found in client.")
			return mds.AuthenticationResponse{
				AuthToken: "y0ur.jwt.T0kEn",
				TokenType: "JWT",
				ExpiresIn: 100,
			}, nil, nil
		},
	}
	mdsClientManager := &cliMock.MockMDSClientManager{
		GetMDSClientFunc: func(ctx *v3.Context, caCertPath string, flagChanged bool, url string, logger *log.Logger) (client *mds.APIClient, e error) {
			mdsClient.GetConfig().HTTPClient, err = pauth.SelfSignedCertClient(certReader, logger)
			if err != nil {
				return nil, err
			}
			return mdsClient, nil
		},
	}
	loginCmd := NewLoginCommand("confluent", prerunner, log.New(), prompt, nil, nil, mdsClientManager, cliMock.NewDummyAnalyticsMock(), nil)
	loginCmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	_, err = pcmd.ExecuteCommand(loginCmd.Command, "--url=http://localhost:8090", "--ca-cert-path=testcert.pem")
	req.NoError(err)
}

func TestLoginWithExistingContext(t *testing.T) {
	req := require.New(t)

	prompt := prompt()
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
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
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: "test-email",
			}, nil
		},
	}

	suite := []struct {
		cliName string
		args    []string
	}{
		{
			cliName: "ccloud",
			args:    []string{},
		},
		{
			cliName: "confluent",
			args: []string{
				"--url=http://localhost:8090",
			},
		},
	}

	activeApiKey := "bo"
	kafkaCluster := &v1.KafkaClusterConfig{
		ID:          "lkc-0000",
		Name:        "bob",
		Bootstrap:   "http://bobby",
		APIEndpoint: "http://bobbyboi",
		APIKeys: map[string]*v0.APIKeyPair{
			activeApiKey: {
				Key:    activeApiKey,
				Secret: "bo",
			},
		},
		APIKey: activeApiKey,
	}

	for _, s := range suite {
		loginCmd, cfg := newLoginCmd(prompt, auth, user, s.cliName, req)

		// Login to the CLI control plane
		output, err := pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, "cody@confluent.io"))
		verifyLoggedInState(t, cfg, s.cliName)

		// Set kafka related states for the logged in context
		ctx := cfg.Context()
		ctx.KafkaClusterContext.AddKafkaClusterConfig(kafkaCluster)
		ctx.KafkaClusterContext.SetActiveKafkaCluster(kafkaCluster.ID)

		// Executing logout
		logoutCmd, _ := newLogoutCmd(cfg.CLIName, cfg)
		output, err = pcmd.ExecuteCommand(logoutCmd.Command)
		req.NoError(err)
		req.Contains(output, errors.LoggedOutMsg)
		verifyLoggedOutState(t, cfg, ctx.Name)

		// logging back in the the same context
		output, err = pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, "cody@confluent.io"))
		verifyLoggedInState(t, cfg, s.cliName)

		// verify that kafka cluster info persists between logging back in again
		req.Equal(kafkaCluster.ID, ctx.KafkaClusterContext.GetActiveKafkaClusterId())
		reflect.DeepEqual(kafkaCluster, ctx.KafkaClusterContext.GetKafkaClusterConfig(kafkaCluster.ID))
	}
}

func TestValidateUrl(t *testing.T) {
	req := require.New(t)

	suite := []struct {
		url_in      string
		valid       bool
		url_out     string
		warning_msg string
		cli         string
	}{
		{
			url_in:      "https:///test.com",
			valid:       false,
			url_out:     "",
			warning_msg: "default MDS port 8090",
			cli:         "confluent",
		},
		{
			url_in:      "test.com",
			valid:       true,
			url_out:     "http://test.com:8090",
			warning_msg: "http protocol and default MDS port 8090",
			cli:         "confluent",
		},
		{
			url_in:      "test.com:80",
			valid:       true,
			url_out:     "http://test.com:80",
			warning_msg: "http protocol",
			cli:         "confluent",
		},
		{
			url_in:      "http://test.com",
			valid:       true,
			url_out:     "http://test.com:8090",
			warning_msg: "default MDS port 8090",
			cli:         "confluent",
		},
		{
			url_in:      "https://127.0.0.1:8090",
			valid:       true,
			url_out:     "https://127.0.0.1:8090",
			warning_msg: "",
			cli:         "confluent",
		},
		{
			url_in:      "127.0.0.1",
			valid:       true,
			url_out:     "http://127.0.0.1:8090",
			warning_msg: "http protocol and default MDS port 8090",
			cli:         "confluent",
		},
		{
			url_in:      "devel.cpdev.cloud",
			valid:       true,
			url_out:     "https://devel.cpdev.cloud",
			warning_msg: "https protocol",
			cli:         "ccloud",
		},
	}
	for _, s := range suite {
		url, matched, msg := validateURL(s.url_in, s.cli)
		req.Equal(s.valid, matched)
		if s.valid {
			req.Equal(s.url_out, url)
		}
		req.Equal(s.warning_msg, msg)
	}
}

func verifyLoggedInState(t *testing.T, cfg *v3.Config, cliName string) {
	req := require.New(t)
	ctx := cfg.Context()
	req.NotNil(ctx)
	req.Equal("y0ur.jwt.T0kEn", ctx.State.AuthToken)
	contextName := fmt.Sprintf("login-cody@confluent.io-%s", ctx.Platform.Server)
	credName := fmt.Sprintf("username-%s", ctx.Credential.Username)
	req.Contains(cfg.Platforms, ctx.Platform.Name)
	req.Equal(ctx.Platform, cfg.Platforms[ctx.PlatformName])
	req.Contains(cfg.Credentials, credName)
	req.Equal("cody@confluent.io", cfg.Credentials[credName].Username)
	req.Contains(cfg.Contexts, contextName)
	req.Equal(ctx.Platform, cfg.Contexts[contextName].Platform)
	req.Equal(ctx.Credential, cfg.Contexts[contextName].Credential)
	if cliName == "ccloud" {
		// MDS doesn't set some things like cfg.Auth.User since e.g. an MDS user != an orgv1 (ccloud) User
		req.Equal(&orgv1.User{Id: 23, Email: "cody@confluent.io", FirstName: "Cody"}, ctx.State.Auth.User)
	} else {
		req.Equal("http://localhost:8090", ctx.Platform.Server)
	}
}

func verifyLoggedOutState(t *testing.T, cfg *v3.Config, loggedOutContext string) {
	req := require.New(t)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}

func prompt() form.Prompt {
	return &pmock.Prompt{
		ReadLineFunc: func() (string, error) {
			return "cody@confluent.io", nil
		},
		ReadLineMaskedFunc: func() (string, error) {
			return " iamrobin ", nil
		},
	}
}

func newLoginCmd(prompt form.Prompt, auth *sdkMock.Auth, user *sdkMock.User, cliName string, req *require.Assertions) (*loginCommand, *v3.Config) {
	var mockAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *ccloud.Client {
		req.Equal("https://confluent.cloud", baseURL)
		return &ccloud.Client{Auth: auth, User: user}
	}
	var mockJwtHTTPClientFactory = func(ctx context.Context, jwt, baseURL string, logger *log.Logger) *ccloud.Client {
		return &ccloud.Client{Auth: auth, User: user}
	}
	cfg := v3.New(&config.Params{
		CLIName:    cliName,
		MetricSink: nil,
		Logger:     nil,
	})
	var mdsClient *mds.APIClient
	if cliName == "confluent" {
		mdsConfig := mds.NewConfiguration()
		mdsClient = mds.NewAPIClient(mdsConfig)
		mdsClient.TokensAndAuthenticationApi = &mdsMock.TokensAndAuthenticationApi{
			GetTokenFunc: func(ctx context.Context) (mds.AuthenticationResponse, *http.Response, error) {
				return mds.AuthenticationResponse{
					AuthToken: "y0ur.jwt.T0kEn",
					TokenType: "JWT",
					ExpiresIn: 100,
				}, nil, nil
			},
		}
	}
	mdsClientManager := &cliMock.MockMDSClientManager{
		GetMDSClientFunc: func(ctx *v3.Context, caCertPath string, flagChanged bool, url string, logger *log.Logger) (client *mds.APIClient, e error) {
			return mdsClient, nil
		},
	}
	prerunner := cliMock.NewPreRunnerMock(mockAnonHTTPClientFactory("https://confluent.cloud", nil), mdsClient, cfg)
	loginCmd := NewLoginCommand(cliName, prerunner, log.New(), prompt, mockAnonHTTPClientFactory, mockJwtHTTPClientFactory, mdsClientManager,
		cliMock.NewDummyAnalyticsMock(), nil)
	return loginCmd, cfg
}

func newLogoutCmd(cliName string, cfg *v3.Config) (*logoutCommand, *v3.Config) {
	logoutCmd := NewLogoutCmd(cliName, cliMock.NewPreRunnerMock(nil, nil, cfg), cliMock.NewDummyAnalyticsMock())
	return logoutCmd, cfg
}
