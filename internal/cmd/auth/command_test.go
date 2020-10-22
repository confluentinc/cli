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
	"github.com/confluentinc/cli/internal/pkg/netrc"

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
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	envUser        = "env-user"
	envPassword    = "env-password"
	depEnvUser     = "dep-env-user"
	depEnvPassword = "dep-env-password"
	testToken      = "y0ur.jwt.T0kEn"
	promptUser     = "cody@confluent.io"
	promptPassword = " iamrobin "
	netrcUser      = "netrc@confleunt.io"
	netrcPassword  = "netrcpassword"
	netrcFile      = "netrc-file"
)

func TestCredentialsOverride(t *testing.T) {
	req := require.New(t)
	currentEmail := os.Getenv(pauth.CCloudEmailEnvVar)
	currentPassword := os.Getenv(pauth.CCloudPasswordEnvVar)

	os.Setenv(pauth.CCloudEmailEnvVar, envUser)
	os.Setenv(pauth.CCloudPasswordEnvVar, envPassword)

	prompt := prompt()
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
	loginCmd, cfg := newLoginCmd(prompt, auth, user, "ccloud", req, netrcHandlerNoCredential())

	output, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.NoError(err)
	req.Contains(output, fmt.Sprintf(errors.FoundEnvCredMsg, envUser, pauth.CCloudEmailEnvVar, pauth.CCloudPasswordEnvVar))
	req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, envUser))
	ctx := cfg.Context()
	req.NotNil(ctx)

	req.Equal(testToken, ctx.State.AuthToken)
	req.Equal(&orgv1.User{Id: 23, Email: envUser, FirstName: "Cody"}, ctx.State.Auth.User)

	os.Setenv(pauth.CCloudEmailEnvVar, currentEmail)
	os.Setenv(pauth.CCloudPasswordEnvVar, currentPassword)
}

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)
	clearEnvironmentVariables()
	prompt := prompt()
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
		loginCmd, cfg := newLoginCmd(prompt, auth, user, s.cliName, req, netrcHandlerNoCredential())
		output, err := pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		verifyLoggedInState(t, cfg, s.cliName)
	}
}

func TestLoginOrderOfPrecedence(t *testing.T) {
	req := require.New(t)
	prompt := prompt()
	auth := &sdkMock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return testToken, nil
		},
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
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: "",
			}, nil
		},
	}

	tests := []struct {
		name                string
		cliName             string
		setEnvVar           bool
		setDeprecatedEnvVar bool
		setNetrcUser        bool
		wantUser            string
	}{
		{
			name:                "CCLOUD: env var over all other credentials",
			cliName:             "ccloud",
			setEnvVar:           true,
			setDeprecatedEnvVar: true,
			setNetrcUser:        true,
			wantUser:            envUser,
		},
		{
			name:                "CCLOUD: deprecated env var over other credentials",
			cliName:             "ccloud",
			setEnvVar:           false,
			setDeprecatedEnvVar: true,
			setNetrcUser:        true,
			wantUser:            depEnvUser,
		},
		{
			name:                "CCLOUD: netrc credential over prompt",
			cliName:             "ccloud",
			setEnvVar:           false,
			setDeprecatedEnvVar: false,
			setNetrcUser:        true,
			wantUser:            netrcUser,
		},
		{
			name:                "CCLOUD prompt",
			cliName:             "ccloud",
			setEnvVar:           false,
			setDeprecatedEnvVar: false,
			setNetrcUser:        false,
			wantUser:            promptUser,
		},
		{
			name:                "CONFLUENT: env var over all other credentials",
			cliName:             "confluent",
			setEnvVar:           true,
			setDeprecatedEnvVar: true,
			setNetrcUser:        true,
			wantUser:            envUser,
		},
		{
			name:                "CONFLUENT: deprecated env var over other credentials",
			cliName:             "confluent",
			setEnvVar:           false,
			setDeprecatedEnvVar: true,
			setNetrcUser:        true,
			wantUser:            depEnvUser,
		},
		{
			name:                "CONFLUENT: netrc credential over prompt",
			cliName:             "confluent",
			setEnvVar:           false,
			setDeprecatedEnvVar: false,
			setNetrcUser:        true,
			wantUser:            netrcUser,
		},
		{
			name:                "CONFLUENT prompt",
			cliName:             "confluent",
			setEnvVar:           false,
			setDeprecatedEnvVar: false,
			setNetrcUser:        false,
			wantUser:            promptUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvironmentVariables()
			loginArgs := []string{}
			if tt.cliName == "ccloud" {
				if tt.setEnvVar {
					setCCloudEnvironmentVariables()
				}
				if tt.setDeprecatedEnvVar {
					setCCloudDeprecatedEnvironmentVariables()
				}
			}
			if tt.cliName == "confluent" {
				loginArgs = []string{"--url=http://localhost:8090"}
				if tt.setEnvVar {
					setConfluentEnvironmentVariables()
				}
				if tt.setDeprecatedEnvVar {
					setConfluentDeprecatedEnvironmentVariables()
				}
			}
			var netrcHandler netrc.NetrcHandler
			if tt.setNetrcUser {
				netrcHandler = netrcHandlerWithCredential()
			} else {
				netrcHandler = netrcHandlerNoCredential()
			}
			loginCmd, _ := newLoginCmd(prompt, auth, user, tt.cliName, req, netrcHandler)
			output, err := pcmd.ExecuteCommand(loginCmd.Command, loginArgs...)
			req.NoError(err)
			req.Contains(output, getFoundCredentialMessage(tt.wantUser, tt.cliName, tt.setEnvVar, tt.setDeprecatedEnvVar, tt.setNetrcUser))
			req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsg, tt.wantUser))
		})
	}
}

func getFoundCredentialMessage(user, cliName string, setEnvVar, setDeprecatedEnvVar, setNetrcUser bool) string {
	if setEnvVar {
		if cliName == "ccloud" {
			return fmt.Sprintf(errors.FoundEnvCredMsg, user, pauth.CCloudEmailEnvVar, pauth.CCloudPasswordEnvVar)
		} else {
			return fmt.Sprintf(errors.FoundEnvCredMsg, user, pauth.ConfluentUsernameEnvVar, pauth.ConfluentPasswordEnvVar)
		}
	}
	if setDeprecatedEnvVar {
		if cliName == "ccloud" {
			return fmt.Sprintf(errors.FoundEnvCredMsg, user, pauth.CCloudEmailDeprecatedEnvVar, pauth.CCloudPasswordDeprecatedEnvVar)
		} else {
			return fmt.Sprintf(errors.FoundEnvCredMsg, user, pauth.ConfluentUsernameDeprecatedEnvVar, pauth.ConfluentPasswordDeprecatedEnvVar)
		}
	}
	if setNetrcUser {
		return fmt.Sprint(errors.FoundNetrcCredMsg, user, netrcFile)
	}
	return ""
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
				Email: envUser,
			}, nil
		},
	}
	loginCmd, _ := newLoginCmd(prompt, auth, user, "ccloud", req, netrcHandlerNoCredential())

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
	loginCmd, _ := newLoginCmd(prompt, auth, nil, "confluent", req, netrcHandlerNoCredential())

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

func Test_getCCloudLoginCredentials_NoSpacesAroundEmail_ShouldSupportSpacesAtBeginOrEnd(t *testing.T) {
	req := require.New(t)
	clearEnvironmentVariables()

	prompt := prompt()
	auth := &sdkMock.Auth{}
	user := &sdkMock.User{
		CheckEmailFunc: func(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				Email: promptUser,
			}, nil
		},
	}
	loginCmd, _ := newLoginCmd(prompt, auth, user, "ccloud", req, netrcHandlerNoCredential())
	email, password, err := loginCmd.getCCloudLoginCredentials(loginCmd.Command, loginCmd.anonHTTPClientFactory("https://confluent.cloud", log.New()))
	req.NoError(err)
	req.Equal(promptUser, email)
	req.Equal(promptPassword, password)
}

func Test_SelfSignedCerts(t *testing.T) {
	req := require.New(t)
	clearEnvironmentVariables()
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
	loginCmd := NewLoginCommand("confluent", prerunner, log.New(), prompt, nil, nil,
		mdsClientManager, cliMock.NewDummyAnalyticsMock(),
		&cliMock.MockNetrcHandler{
			GetMatchingNetrcCredentialsFunc: func(netrc.GetMatchingNetrcCredentialsParams) (string, string, error) { return "", "", nil },
		},
	)
	loginCmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")
	_, err = pcmd.ExecuteCommand(loginCmd.Command, "--url=http://localhost:8090", "--ca-cert-path=testcert.pem")
	req.NoError(err)
}

func TestLoginWithExistingContext(t *testing.T) {
	req := require.New(t)
	clearEnvironmentVariables()

	prompt := prompt()
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
		loginCmd, cfg := newLoginCmd(prompt, auth, user, s.cliName, req, netrcHandlerNoCredential())

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
	contextName := fmt.Sprintf("login-cody@confluent.io-%s", ctx.Platform.Server)
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

func prompt() *cliMock.Prompt {
	return &cliMock.Prompt{
		ReadLineFunc: func() (string, error) {
			return promptUser, nil
		},
		ReadLineMaskedFunc: func() (string, error) {
			return promptPassword, nil
		},
	}
}

func netrcHandlerNoCredential() netrc.NetrcHandler {
	return &cliMock.MockNetrcHandler{
		GetMatchingNetrcCredentialsFunc: func(netrc.GetMatchingNetrcCredentialsParams) (string, string, error) { return "", "", nil },
		GetFileNameFunc:                 func() string { return netrcFile },
	}
}

func netrcHandlerWithCredential() netrc.NetrcHandler {
	return &cliMock.MockNetrcHandler{
		GetMatchingNetrcCredentialsFunc: func(netrc.GetMatchingNetrcCredentialsParams) (string, string, error) {
			return netrcUser, netrcPassword, nil
		},
		GetFileNameFunc: func() string { return netrcFile },
	}
}

func newLoginCmd(prompt pcmd.Prompt, auth *sdkMock.Auth, user *sdkMock.User, cliName string, req *require.Assertions, netrcHandler netrc.NetrcHandler) (*loginCommand, *v3.Config) {
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
	prerunner := cliMock.NewPreRunnerMock(mockAnonHTTPClientFactory("https://confluent.cloud", nil), mdsClient, cfg)
	loginCmd := NewLoginCommand(cliName, prerunner, log.New(), prompt,
		mockAnonHTTPClientFactory, mockJwtHTTPClientFactory, mdsClientManager, cliMock.NewDummyAnalyticsMock(), netrcHandler,
	)
	return loginCmd, cfg
}

func newLogoutCmd(cliName string, cfg *v3.Config) (*logoutCommand, *v3.Config) {
	logoutCmd := NewLogoutCmd(cliName, cliMock.NewPreRunnerMock(nil, nil, cfg), cliMock.NewDummyAnalyticsMock())
	return logoutCmd, cfg
}

func clearEnvironmentVariables() {
	os.Setenv(pauth.CCloudEmailEnvVar, "")
	os.Setenv(pauth.CCloudPasswordEnvVar, "")
	os.Setenv(pauth.CCloudEmailDeprecatedEnvVar, "")
	os.Setenv(pauth.CCloudPasswordDeprecatedEnvVar, "")

	os.Setenv(pauth.ConfluentUsernameEnvVar, "")
	os.Setenv(pauth.ConfluentPasswordEnvVar, "")
	os.Setenv(pauth.ConfluentUsernameDeprecatedEnvVar, "")
	os.Setenv(pauth.ConfluentPasswordDeprecatedEnvVar, "")
}

func setCCloudEnvironmentVariables() {
	os.Setenv(pauth.CCloudEmailEnvVar, envUser)
	os.Setenv(pauth.CCloudPasswordEnvVar, envPassword)
}

func setCCloudDeprecatedEnvironmentVariables() {
	os.Setenv(pauth.CCloudEmailDeprecatedEnvVar, depEnvUser)
	os.Setenv(pauth.CCloudPasswordDeprecatedEnvVar, depEnvPassword)
}

func setConfluentEnvironmentVariables() {
	os.Setenv(pauth.ConfluentUsernameEnvVar, envUser)
	os.Setenv(pauth.ConfluentPasswordEnvVar, envPassword)
}

func setConfluentDeprecatedEnvironmentVariables() {
	os.Setenv(pauth.ConfluentUsernameEnvVar, depEnvUser)
	os.Setenv(pauth.ConfluentPasswordDeprecatedEnvVar, depEnvPassword)
}
