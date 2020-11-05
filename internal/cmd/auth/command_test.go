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

	"github.com/spf13/cobra"

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
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	envUser        = "env-user"
	envPassword    = "env-password"
	testToken      = "y0ur.jwt.T0kEn"
	promptUser     = "prompt-user@confluent.io"
	promptPassword = " prompt-password "
	netrcFile      = "netrc-file"
	ccloudURL      = "https://confluent.cloud"
)

var (
	envCreds = &pauth.Credentials{
		Username: envUser,
		Password: envPassword,
	}
	mockAuth = &sdkMock.Auth{
		UserFunc: func(ctx context.Context) (*orgv1.GetUserReply, error) {
			return &orgv1.GetUserReply{
				User: &orgv1.User{
					Id:        23,
					Email:     "",
					FirstName: "",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	mockUser = &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: "",
			}, nil
		},
	}
	mockLoginTokenHandler = &cliMock.MockLoginTokenHandler{
		GetCCloudTokenAndCredentialsFromEnvVarFunc: func(cmd *cobra.Command, client *ccloud.Client) (string, *pauth.Credentials, error) {
			return "", nil, nil
		},
		GetCCloudTokenAndCredentialsFromNetrcFunc: func(cmd *cobra.Command, client *ccloud.Client, url string, filterParams netrc.GetMatchingNetrcMachineParams) (string, *pauth.Credentials, error) {
			return "", nil, nil
		},
		GetCCloudTokenAndCredentialsFromPromptFunc: func(cmd *cobra.Command, client *ccloud.Client, url string) (string, *pauth.Credentials, error) {
			return testToken, &pauth.Credentials{
				Username: promptUser,
				Password: promptPassword,
			}, nil
		},
		GetConfluentTokenAndCredentialsFromEnvVarFunc: func(cmd *cobra.Command, client *mds.APIClient) (string, *pauth.Credentials, error) {
			return "", nil, nil
		},
		GetConfluentTokenAndCredentialsFromNetrcFunc: func(cmd *cobra.Command, client *mds.APIClient, filterParams netrc.GetMatchingNetrcMachineParams) (string, *pauth.Credentials, error) {
			return "", nil, nil
		},
		GetConfluentTokenAndCredentialsFromPromptFunc: func(cmd *cobra.Command, client *mds.APIClient) (string, *pauth.Credentials, error) {
			return testToken, &pauth.Credentials{
				Username: promptUser,
				Password: promptPassword,
			}, nil
		},
	}
	mockAuthTokenHandler = &pmock.MockAuthTokenHandler{
		GetCCloudCredentialsTokenFunc: func(client *ccloud.Client, email, password string) (string, error) {
			return testToken, nil
		},
		GetCCloudSSOTokenFunc: func(client *ccloud.Client, url string, noBrowser bool, email string, logger *log.Logger) (string, string, error) {
			return testToken, "refreshToken", nil
		},
		GetConfluentAuthTokenFunc: func(mdsClient *mds.APIClient, username, password string, logger *log.Logger) (string, error) {
			return testToken, nil
		},
	}
	mockNetrcHandler = &pmock.MockNetrcHandler{
		GetFileNameFunc: func() string { return netrcFile },
		WriteNetrcCredentialsFunc: func(cliName string, isSSO bool, ctxName, username, password string) error {
			return nil
		},
	}
)

func TestCredentialsOverride(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(ctx context.Context) (*orgv1.GetUserReply, error) {
			return &orgv1.GetUserReply{
				User: &orgv1.User{
					Id:        23,
					Email:     envUser,
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: envUser,
			}, nil
		},
	}
	mockLoginTokenHandler := &cliMock.MockLoginTokenHandler{
		GetCCloudTokenAndCredentialsFromEnvVarFunc: func(cmd *cobra.Command, client *ccloud.Client) (string, *pauth.Credentials, error) {
			return testToken, envCreds, nil
		},
	}
	loginCmd, cfg := newLoginCmd(auth, user, "ccloud", req, mockNetrcHandler, mockAuthTokenHandler, mockLoginTokenHandler)

	output, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.NoError(err)
	req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, envUser))
	ctx := cfg.Context()
	req.NotNil(ctx)

	req.Equal(testToken, ctx.State.AuthToken)
	req.Equal(&orgv1.User{Id: 23, Email: envUser, FirstName: "Cody"}, ctx.State.Auth.User)
}

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(ctx context.Context) (*orgv1.GetUserReply, error) {
			return &orgv1.GetUserReply{
				User: &orgv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: promptUser,
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
		loginCmd, cfg := newLoginCmd(auth, user, s.cliName, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginTokenHandler)
		output, err := pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		verifyLoggedInState(t, cfg, s.cliName)
	}
}

func TestLoginOrderOfPrecedence(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	netrcUser := "netrc@confleunt.io"
	netrcPassword := "netrcpassword"
	netrcCreds := &pauth.Credentials{
		Username: netrcUser,
		Password: netrcPassword,
	}

	tests := []struct {
		name         string
		cliName      string
		setEnvVar    bool
		setNetrcUser bool
		wantUser     string
	}{
		{
			name:         "CCLOUD env var over all other credentials",
			cliName:      "ccloud",
			setEnvVar:    true,
			setNetrcUser: true,
			wantUser:     envUser,
		},
		{
			name:         "CCLOUD netrc credential over prompt",
			cliName:      "ccloud",
			setEnvVar:    false,
			setNetrcUser: true,
			wantUser:     netrcUser,
		},
		{
			name:         "CCLOUD prompt",
			cliName:      "ccloud",
			setEnvVar:    false,
			setNetrcUser: false,
			wantUser:     promptUser,
		},
		{
			name:         "CONFLUENT env var over all other credentials",
			cliName:      "confluent",
			setEnvVar:    true,
			setNetrcUser: true,
			wantUser:     envUser,
		},
		{
			name:         "CONFLUENT netrc credential over prompt",
			cliName:      "confluent",
			setEnvVar:    false,
			setNetrcUser: true,
			wantUser:     netrcUser,
		},
		{
			name:         "CONFLUENT prompt",
			cliName:      "confluent",
			setEnvVar:    false,
			setNetrcUser: false,
			wantUser:     promptUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonInteractiveLoginHandler := &cliMock.MockLoginTokenHandler{
				GetCCloudTokenAndCredentialsFromEnvVarFunc: func(cmd *cobra.Command, client *ccloud.Client) (string, *pauth.Credentials, error) {
					return "", nil, nil
				},
				GetCCloudTokenAndCredentialsFromNetrcFunc: func(cmd *cobra.Command, client *ccloud.Client, url string, filterParams netrc.GetMatchingNetrcMachineParams) (string, *pauth.Credentials, error) {
					return "", nil, nil
				},
				GetCCloudTokenAndCredentialsFromPromptFunc: func(cmd *cobra.Command, client *ccloud.Client, url string) (string, *pauth.Credentials, error) {
					return testToken, &pauth.Credentials{
						Username: promptUser,
						Password: promptPassword,
					}, nil
				},
				GetConfluentTokenAndCredentialsFromEnvVarFunc: func(cmd *cobra.Command, client *mds.APIClient) (string, *pauth.Credentials, error) {
					return "", nil, nil
				},
				GetConfluentTokenAndCredentialsFromNetrcFunc: func(cmd *cobra.Command, client *mds.APIClient, filterParams netrc.GetMatchingNetrcMachineParams) (string, *pauth.Credentials, error) {
					return "", nil, nil
				},
				GetConfluentTokenAndCredentialsFromPromptFunc: func(cmd *cobra.Command, client *mds.APIClient) (string, *pauth.Credentials, error) {
					return testToken, &pauth.Credentials{
						Username: promptUser,
						Password: promptPassword,
					}, nil
				},
			}
			if tt.cliName == "ccloud" {
				if tt.setEnvVar {
					nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromEnvVarFunc = func(cmd *cobra.Command, client *ccloud.Client) (string, *pauth.Credentials, error) {
						return testToken, envCreds, nil
					}
				}
				if tt.setNetrcUser {
					nonInteractiveLoginHandler.GetCCloudTokenAndCredentialsFromNetrcFunc = func(cmd *cobra.Command, client *ccloud.Client, url string, filterParams netrc.GetMatchingNetrcMachineParams) (string, *pauth.Credentials, error) {
						return testToken, netrcCreds, nil
					}
				}
			} else {
				if tt.setEnvVar {
					nonInteractiveLoginHandler.GetConfluentTokenAndCredentialsFromEnvVarFunc = func(cmd *cobra.Command, client *mds.APIClient) (string, *pauth.Credentials, error) {
						return testToken, envCreds, nil
					}
				}
				if tt.setNetrcUser {
					nonInteractiveLoginHandler.GetConfluentTokenAndCredentialsFromNetrcFunc = func(cmd *cobra.Command, client *mds.APIClient, filterParams netrc.GetMatchingNetrcMachineParams) (string, *pauth.Credentials, error) {
						return testToken, netrcCreds, nil
					}
				}
			}
			loginCmd, _ := newLoginCmd(mockAuth, mockUser, tt.cliName, req, mockNetrcHandler, mockAuthTokenHandler, nonInteractiveLoginHandler)
			var loginArgs []string
			if tt.cliName == "confluent" {
				loginArgs = []string{"--url=http://localhost:8090"}
			}
			output, err := pcmd.ExecuteCommand(loginCmd.Command, loginArgs...)
			req.NoError(err)
			req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, tt.wantUser))
		})
	}
}

func TestLoginFail(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	mockLoginTokenHandler := &cliMock.MockLoginTokenHandler{
		GetCCloudTokenAndCredentialsFromEnvVarFunc: func(cmd *cobra.Command, client *ccloud.Client) (string, *pauth.Credentials, error) {
			return "", nil, errors.New("DO NOT RETURN THIS ERR")
		},
		GetCCloudTokenAndCredentialsFromNetrcFunc: func(cmd *cobra.Command, client *ccloud.Client, url string, params netrc.GetMatchingNetrcMachineParams) (string, *pauth.Credentials, error) {
			return "", nil, errors.New("DO NOT RETURN THIS ERR")
		},
		GetCCloudTokenAndCredentialsFromPromptFunc: func(cmd *cobra.Command, client *ccloud.Client, url string) (string, *pauth.Credentials, error) {
			return "", nil, &ccloud.InvalidLoginError{}
		},
	}
	loginCmd, _ := newLoginCmd(mockAuth, mockUser, "ccloud", req, mockNetrcHandler, mockAuthTokenHandler, mockLoginTokenHandler)
	_, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.Contains(err.Error(), errors.InvalidLoginErrorMsg)
	errors.VerifyErrorAndSuggestions(req, err, errors.InvalidLoginErrorMsg, errors.CCloudInvalidLoginSuggestions)
}

func TestURLRequiredWithMDS(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return "", &ccloud.InvalidLoginError{}
		},
	}
	loginCmd, _ := newLoginCmd(auth, nil, "confluent", req, mockNetrcHandler, mockAuthTokenHandler, mockLoginTokenHandler)

	_, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.Contains(err.Error(), "required flag(s) \"url\" not set")
}

func TestLogout(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	cfg := v3.AuthenticatedCloudConfigMock()
	contextName := cfg.Context().Name
	logoutCmd, cfg := newLogoutCmd("ccloud", cfg)
	output, err := pcmd.ExecuteCommand(logoutCmd.Command)
	req.NoError(err)
	req.Contains(output, errors.LoggedOutMsg)
	verifyLoggedOutState(t, cfg, contextName)
}

func Test_SelfSignedCerts(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	mdsConfig := mds.NewConfiguration()
	mdsClient := mds.NewAPIClient(mdsConfig)
	cfg := v3.New(&config.Params{
		CLIName:    "confluent",
		MetricSink: nil,
		Logger:     log.New(),
	})
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
				AuthToken: testToken,
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
	loginCmd := NewLoginCommand("confluent", prerunner, log.New(), nil, nil,
		mdsClientManager, cliMock.NewDummyAnalyticsMock(), mockNetrcHandler, mockLoginTokenHandler)
	loginCmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	_, err = pcmd.ExecuteCommand(loginCmd.Command, "--url=http://localhost:8090", "--ca-cert-path=testcert.pem")
	req.NoError(err)
}

func TestLoginWithExistingContext(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(ctx context.Context) (*orgv1.GetUserReply, error) {
			return &orgv1.GetUserReply{
				User: &orgv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: promptUser,
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
		loginCmd, cfg := newLoginCmd(auth, user, s.cliName, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginTokenHandler)

		// Login to the CLI control plane
		output, err := pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
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
		req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		verifyLoggedInState(t, cfg, s.cliName)

		// verify that kafka cluster info persists between logging back in again
		req.Equal(kafkaCluster.ID, ctx.KafkaClusterContext.GetActiveKafkaClusterId())
		reflect.DeepEqual(kafkaCluster, ctx.KafkaClusterContext.GetKafkaClusterConfig(kafkaCluster.ID))
	}
}

func TestValidateUrl(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
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
	req.Equal(testToken, ctx.State.AuthToken)
	contextName := fmt.Sprintf("login-%s-%s", promptUser, ctx.Platform.Server)
	credName := fmt.Sprintf("username-%s", ctx.Credential.Username)
	req.Contains(cfg.Platforms, ctx.Platform.Name)
	req.Equal(ctx.Platform, cfg.Platforms[ctx.PlatformName])
	req.Contains(cfg.Credentials, credName)
	req.Equal(promptUser, cfg.Credentials[credName].Username)
	req.Contains(cfg.Contexts, contextName)
	req.Equal(ctx.Platform, cfg.Contexts[contextName].Platform)
	req.Equal(ctx.Credential, cfg.Contexts[contextName].Credential)
	if cliName == "ccloud" {
		// MDS doesn't set some things like cfg.Auth.User since e.g. an MDS user != an orgv1 (ccloud) User
		req.Equal(&orgv1.User{Id: 23, Email: promptUser, FirstName: "Cody"}, ctx.State.Auth.User)
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

func newLoginCmd(auth *sdkMock.Auth, user *sdkMock.User, cliName string, req *require.Assertions, netrcHandler netrc.NetrcHandler,
	authTokenHandler pauth.AuthTokenHandler, nonInteractiveLoginHandler pauth.LoginTokenHandler) (*loginCommand, *v3.Config) {
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
					AuthToken: testToken,
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
	prerunner := cliMock.NewPreRunnerMock(mockAnonHTTPClientFactory(ccloudURL, nil), mdsClient, cfg)
	loginCmd := NewLoginCommand(cliName, prerunner, log.New(),
		mockAnonHTTPClientFactory, mockJwtHTTPClientFactory, mdsClientManager,
		cliMock.NewDummyAnalyticsMock(), netrcHandler, nonInteractiveLoginHandler)
	return loginCmd, cfg
}

func newLogoutCmd(cliName string, cfg *v3.Config) (*logoutCommand, *v3.Config) {
	logoutCmd := NewLogoutCmd(cliName, cliMock.NewPreRunnerMock(nil, nil, cfg), cliMock.NewDummyAnalyticsMock())
	return logoutCmd, cfg
}

// XX_CCLOUD_EMAIL is used for integration test hack
func clearCCloudDeprecatedEnvVar(req *require.Assertions) {
	req.NoError(os.Setenv(pauth.CCloudEmailDeprecatedEnvVar, ""))
}
