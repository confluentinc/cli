package login

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	sdkMock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	mdsMock "github.com/confluentinc/mds-sdk-go/mdsv1/mock"

	"github.com/confluentinc/cli/internal/cmd/logout"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
		UserFunc: func(_ context.Context) (*flowv1.GetMeReply, error) {
			return &flowv1.GetMeReply{
				User: &orgv1.User{
					Id:        23,
					Email:     "",
					FirstName: "",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	mockUser                    = &sdkMock.User{}
	mockLoginCredentialsManager = &cliMock.MockLoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},
		GetOnPremCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetOnPremCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(arg0 *ccloud.Client) {
		},
	}
	mockAuthTokenHandler = &cliMock.MockAuthTokenHandler{
		GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, _ *pauth.Credentials, _ bool) (s string, s2 string, e error) {
			return testToken, "refreshToken", nil
		},
		GetConfluentTokenFunc: func(mdsClient *mds.APIClient, credentials *pauth.Credentials) (s string, e error) {
			return testToken, nil
		},
	}
	mockNetrcHandler = &pmock.MockNetrcHandler{
		GetFileNameFunc: func() string { return netrcFile },
		WriteNetrcCredentialsFunc: func(isCloud bool, isSSO bool, ctxName, username, password string) error {
			return nil
		},
		RemoveNetrcCredentialsFunc: func(isCloud bool, ctxName string) (string, error) {
			return "", nil
		},
		CheckCredentialExistFunc: func(isCloud bool, ctxName string) (bool, error) {
			return false, nil
		},
	}
)

func TestCredentialsOverride(t *testing.T) {
	req := require.New(t)
	auth := &sdkMock.Auth{
		LoginFunc: func(_ context.Context, _, _, _, _ string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(_ context.Context) (*flowv1.GetMeReply, error) {
			return &flowv1.GetMeReply{
				User: &orgv1.User{
					Id:        23,
					Email:     envUser,
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{}
	mockLoginCredentialsManager := &cliMock.MockLoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return envCreds, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(arg0 *ccloud.Client) {
		},
	}
	loginCmd, cfg := newLoginCmd(auth, user, true, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginCredentialsManager)

	output, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.NoError(err)
	req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, envUser))

	ctx := cfg.Context()
	req.NotNil(ctx)
	req.Equal(pauth.GenerateContextName(envUser, ccloudURL, ""), ctx.Name)

	req.Equal(testToken, ctx.State.AuthToken)
	req.Equal(&orgv1.User{Id: 23, Email: envUser, FirstName: "Cody"}, ctx.State.Auth.User)
}

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)
	auth := &sdkMock.Auth{
		LoginFunc: func(_ context.Context, _, _, _, _ string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(_ context.Context) (*flowv1.GetMeReply, error) {
			return &flowv1.GetMeReply{
				User: &orgv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{}

	suite := []struct {
		isCloud bool
		args    []string
		setEnv  bool
	}{
		{
			isCloud: true,
			args:    []string{},
		},
		{
			args: []string{"--url=http://localhost:8090"},
		},
		{
			setEnv: true,
		},
	}

	for _, s := range suite {
		// Log in to the CLI control plane
		if s.setEnv {
			_ = os.Setenv(pauth.ConfluentPlatformMDSURL, "http://localhost:8090")
		}
		loginCmd, cfg := newLoginCmd(auth, user, s.isCloud, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginCredentialsManager)
		output, err := pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		verifyLoggedInState(t, cfg, s.isCloud)
		if s.setEnv {
			_ = os.Unsetenv(pauth.ConfluentPlatformMDSURL)
		}
	}
}

func TestLoginOrderOfPrecedence(t *testing.T) {
	req := require.New(t)
	netrcUser := "netrc@confleunt.io"
	netrcPassword := "netrcpassword"
	netrcCreds := &pauth.Credentials{
		Username: netrcUser,
		Password: netrcPassword,
	}

	tests := []struct {
		name         string
		isCloud      bool
		setEnvVar    bool
		setNetrcUser bool
		wantUser     string
	}{
		{
			name:         "cloud env var over all other credentials",
			isCloud:      true,
			setEnvVar:    true,
			setNetrcUser: true,
			wantUser:     envUser,
		},
		{
			name:         "cloud netrc credential over prompt",
			isCloud:      true,
			setEnvVar:    false,
			setNetrcUser: true,
			wantUser:     netrcUser,
		},
		{
			name:         "cloud prompt",
			isCloud:      true,
			setEnvVar:    false,
			setNetrcUser: false,
			wantUser:     promptUser,
		},
		{
			name:         "on-prem env var over all other credentials",
			setEnvVar:    true,
			setNetrcUser: true,
			wantUser:     envUser,
		},
		{
			name:         "on-prem netrc credential over prompt",
			setEnvVar:    false,
			setNetrcUser: true,
			wantUser:     netrcUser,
		},
		{
			name:         "on-prem prompt",
			setEnvVar:    false,
			setNetrcUser: false,
			wantUser:     promptUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginCredentialsManager := &cliMock.MockLoginCredentialsManager{
				GetCloudCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCloudCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{
							Username: promptUser,
							Password: promptPassword,
						}, nil
					}
				},
				GetOnPremCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetOnPremCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{
							Username: promptUser,
							Password: promptPassword,
						}, nil
					}
				},
				GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				SetCloudClientFunc: func(arg0 *ccloud.Client) {
				},
			}
			if tt.setNetrcUser {
				loginCredentialsManager.GetCredentialsFromNetrcFunc = func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return netrcCreds, nil
					}
				}
			}
			if tt.isCloud {
				if tt.setEnvVar {
					loginCredentialsManager.GetCloudCredentialsFromEnvVarFunc = func(_ *cobra.Command) func() (*pauth.Credentials, error) {
						return func() (*pauth.Credentials, error) {
							return envCreds, nil
						}
					}
				}
			} else {
				if tt.setEnvVar {
					loginCredentialsManager.GetOnPremCredentialsFromEnvVarFunc = func(_ *cobra.Command) func() (*pauth.Credentials, error) {
						return func() (*pauth.Credentials, error) {
							return envCreds, nil
						}
					}
				}
			}
			loginCmd, _ := newLoginCmd(mockAuth, mockUser, tt.isCloud, req, mockNetrcHandler, mockAuthTokenHandler, loginCredentialsManager)
			var loginArgs []string
			if !tt.isCloud {
				loginArgs = []string{"--url=http://localhost:8090"}
			}
			output, err := pcmd.ExecuteCommand(loginCmd.Command, loginArgs...)
			req.NoError(err)
			req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, tt.wantUser))
		})
	}
}

func TestPromptLoginFlag(t *testing.T) {
	req := require.New(t)
	wrongCreds := &pauth.Credentials{
		Username: "wrong_user",
		Password: "wrong_password",
	}

	tests := []struct {
		name    string
		isCloud bool
	}{
		{
			name:    "cloud login prompt flag",
			isCloud: true,
		},
		{
			name: "on-prem login prompt flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoginCredentialsManager := &cliMock.MockLoginCredentialsManager{
				GetCloudCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return wrongCreds, nil
					}
				},
				GetCloudCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{
							Username: promptUser,
							Password: promptPassword,
						}, nil
					}
				},
				GetOnPremCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return wrongCreds, nil
					}
				},
				GetOnPremCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{
							Username: promptUser,
							Password: promptPassword,
						}, nil
					}
				},
				GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return wrongCreds, nil
					}
				},
				SetCloudClientFunc: func(arg0 *ccloud.Client) {
				},
			}
			loginCmd, _ := newLoginCmd(mockAuth, mockUser, tt.isCloud, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginCredentialsManager)
			loginArgs := []string{"--prompt"}
			if !tt.isCloud {
				loginArgs = append(loginArgs, "--url=http://localhost:8090")
			}
			output, err := pcmd.ExecuteCommand(loginCmd.Command, loginArgs...)
			req.NoError(err)

			req.False(mockLoginCredentialsManager.GetCloudCredentialsFromEnvVarCalled())
			req.False(mockLoginCredentialsManager.GetOnPremCredentialsFromEnvVarCalled())
			req.False(mockLoginCredentialsManager.GetCredentialsFromNetrcCalled())

			req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		})
	}
}

func TestLoginFail(t *testing.T) {
	req := require.New(t)
	mockLoginCredentialsManager := &cliMock.MockLoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, errors.New("DO NOT RETURN THIS ERR")
			}
		},
		GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, errors.New("DO NOT RETURN THIS ERR")
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ *cobra.Command) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, &ccloud.InvalidLoginError{}
			}
		},
		SetCloudClientFunc: func(arg0 *ccloud.Client) {
		},
	}
	loginCmd, _ := newLoginCmd(mockAuth, mockUser, true, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginCredentialsManager)
	_, err := pcmd.ExecuteCommand(loginCmd.Command)
	req.Contains(err.Error(), errors.InvalidLoginErrorMsg)
	errors.VerifyErrorAndSuggestions(req, err, errors.InvalidLoginErrorMsg, errors.CCloudInvalidLoginSuggestions)
}

func Test_SelfSignedCerts(t *testing.T) {
	req := require.New(t)
	tests := []struct {
		name                string
		caCertPathFlag      string
		expectedContextName string
		setEnv              bool
		envCertPath         string
	}{
		{
			name:                "specified ca-cert-path",
			caCertPathFlag:      "testcert.pem",
			expectedContextName: "login-prompt-user@confluent.io-http://localhost:8090?cacertpath=%s",
		},
		{
			name:                "no ca-cert-path flag",
			caCertPathFlag:      "",
			expectedContextName: "login-prompt-user@confluent.io-http://localhost:8090",
		},
		{
			name:                "env var ca-cert-path flag",
			setEnv:              true,
			envCertPath:         "testcert.pem",
			expectedContextName: "login-prompt-user@confluent.io-http://localhost:8090?cacertpath=%s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(pauth.ConfluentPlatformCACertPath, "testcert.pem")
			}
			cfg := v1.New(&config.Params{})
			var expectedCaCert string
			if tt.setEnv {
				expectedCaCert = tt.envCertPath
			} else {
				expectedCaCert = tt.caCertPathFlag
			}
			loginCmd := getNewLoginCommandForSelfSignedCertTest(req, cfg, expectedCaCert)
			_, err := pcmd.ExecuteCommand(loginCmd.Command, "--url=http://localhost:8090", fmt.Sprintf("--ca-cert-path=%s", tt.caCertPathFlag))
			req.NoError(err)

			ctx := cfg.Context()

			if tt.setEnv {
				req.Contains(ctx.Platform.CaCertPath, tt.envCertPath)
			} else {
				// ensure the right CaCertPath is stored in Config
				req.Contains(ctx.Platform.CaCertPath, tt.caCertPathFlag)
			}

			if tt.caCertPathFlag != "" || tt.envCertPath != "" {
				req.Equal(fmt.Sprintf(tt.expectedContextName, ctx.Platform.CaCertPath), ctx.Name)
			} else {
				req.Equal(tt.expectedContextName, ctx.Name)
			}
			if tt.setEnv {
				os.Unsetenv(pauth.ConfluentPlatformCACertPath)
			}
		})
	}
}

func Test_SelfSignedCertsLegacyContexts(t *testing.T) {
	originalCaCertPath, _ := filepath.Abs("ogcert.pem")

	req := require.New(t)
	tests := []struct {
		name               string
		useCaCertPathFlag  bool
		expectedCaCertPath string
	}{
		{
			name:               "use existing caCertPath in config",
			useCaCertPathFlag:  false,
			expectedCaCertPath: originalCaCertPath,
		},
		{
			name:              "reset ca-cert-path",
			useCaCertPathFlag: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			legacyContextName := "login-prompt-user@confluent.io-http://localhost:8090"
			cfg := v1.AuthenticatedConfigMockWithContextName(legacyContextName)
			cfg.Contexts[legacyContextName].Platform.CaCertPath = originalCaCertPath

			loginCmd := getNewLoginCommandForSelfSignedCertTest(req, cfg, tt.expectedCaCertPath)
			args := []string{"--url=http://localhost:8090"}
			if tt.useCaCertPathFlag {
				args = append(args, "--ca-cert-path=")
			}
			_, err := pcmd.ExecuteCommand(loginCmd.Command, args...)
			req.NoError(err)

			ctx := cfg.Context()
			// ensure the right CaCertPath is stored in Config and context name is the right name
			req.Equal(tt.expectedCaCertPath, ctx.Platform.CaCertPath)
			req.Equal(legacyContextName, ctx.Name)
		})
	}
}

func getNewLoginCommandForSelfSignedCertTest(req *require.Assertions, cfg *v1.Config, expectedCaCertPath string) *Command {
	mdsConfig := mds.NewConfiguration()
	mdsClient := mds.NewAPIClient(mdsConfig)

	prerunner := cliMock.NewPreRunnerMock(nil, nil, nil, cfg)

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
		GetMDSClientFunc: func(url string, caCertPath string) (client *mds.APIClient, e error) {
			// ensure the right caCertPath is used
			req.Contains(caCertPath, expectedCaCertPath)
			mdsClient.GetConfig().HTTPClient, err = utils.SelfSignedCertClient(certReader, tls.Certificate{})
			if err != nil {
				return nil, err
			}
			return mdsClient, nil
		},
	}
	loginCmd := New(prerunner, nil, mdsClientManager, cliMock.NewDummyAnalyticsMock(), mockNetrcHandler,
		mockLoginCredentialsManager, mockAuthTokenHandler, true)
	loginCmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")

	return loginCmd
}

func TestLoginWithExistingContext(t *testing.T) {
	req := require.New(t)
	auth := &sdkMock.Auth{
		LoginFunc: func(_ context.Context, _, _, _, _ string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(_ context.Context) (*flowv1.GetMeReply, error) {
			return &flowv1.GetMeReply{
				User: &orgv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &sdkMock.User{}

	suite := []struct {
		isCloud bool
		args    []string
	}{
		{
			isCloud: true,
			args:    []string{},
		},
		{
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
		APIKeys: map[string]*v1.APIKeyPair{
			activeApiKey: {
				Key:    activeApiKey,
				Secret: "bo",
			},
		},
		APIKey: activeApiKey,
	}

	for _, s := range suite {
		loginCmd, cfg := newLoginCmd(auth, user, s.isCloud, req, mockNetrcHandler, mockAuthTokenHandler, mockLoginCredentialsManager)

		// Login to the CLI control plane
		output, err := pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		verifyLoggedInState(t, cfg, s.isCloud)

		// Set kafka related states for the logged in context
		ctx := cfg.Context()
		ctx.KafkaClusterContext.AddKafkaClusterConfig(kafkaCluster)
		ctx.KafkaClusterContext.SetActiveKafkaCluster(kafkaCluster.ID)

		// Executing logout
		logoutCmd, _ := newLogoutCmd(cfg, mockNetrcHandler)
		output, err = pcmd.ExecuteCommand(logoutCmd.Command)
		req.NoError(err)
		req.Contains(output, errors.LoggedOutMsg)
		verifyLoggedOutState(t, cfg, ctx.Name)

		// logging back in the the same context
		output, err = pcmd.ExecuteCommand(loginCmd.Command, s.args...)
		req.NoError(err)
		req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		verifyLoggedInState(t, cfg, s.isCloud)

		// verify that kafka cluster info persists between logging back in again
		req.Equal(kafkaCluster.ID, ctx.KafkaClusterContext.GetActiveKafkaClusterId())
		reflect.DeepEqual(kafkaCluster, ctx.KafkaClusterContext.GetKafkaClusterConfig(kafkaCluster.ID))
	}
}

func TestValidateUrl(t *testing.T) {
	req := require.New(t)
	suite := []struct {
		urlIn      string
		urlOut     string
		warningMsg string
		isCCloud   bool
		errMsg     string
	}{
		{
			urlIn:      "https:///test.com",
			urlOut:     "",
			warningMsg: "default MDS port 8090",
			errMsg:     errors.InvalidLoginURLMsg,
		},
		{
			urlIn:      "test.com",
			urlOut:     "http://test.com:8090",
			warningMsg: "http protocol and default MDS port 8090",
		},
		{
			urlIn:      "test.com:80",
			urlOut:     "http://test.com:80",
			warningMsg: "http protocol",
		},
		{
			urlIn:      "http://test.com",
			urlOut:     "http://test.com:8090",
			warningMsg: "default MDS port 8090",
		},
		{
			urlIn:      "https://127.0.0.1:8090",
			urlOut:     "https://127.0.0.1:8090",
			warningMsg: "",
		},
		{
			urlIn:      "127.0.0.1",
			urlOut:     "http://127.0.0.1:8090",
			warningMsg: "http protocol and default MDS port 8090",
		},
		{
			urlIn:      "devel.cpdev.cloud/",
			urlOut:     "https://devel.cpdev.cloud/",
			warningMsg: "https protocol",
			isCCloud:   true,
		},
		{
			urlIn:    "confluent.cloud:123",
			isCCloud: true,
			errMsg:   errors.NewErrorWithSuggestions(errors.UnneccessaryUrlFlagForCloudLoginErrorMsg, errors.UnneccessaryUrlFlagForCloudLoginSuggestions).Error(),
		},
		{
			urlIn:    "https://confluent.cloud/login/sso/company",
			isCCloud: true,
			errMsg:   errors.NewErrorWithSuggestions(errors.UnneccessaryUrlFlagForCloudLoginErrorMsg, errors.UnneccessaryUrlFlagForCloudLoginSuggestions).Error(),
		},
		{
			urlIn:    "https://devel.cpdev.cloud//",
			isCCloud: true,
			errMsg:   errors.NewErrorWithSuggestions(errors.UnneccessaryUrlFlagForCloudLoginErrorMsg, errors.UnneccessaryUrlFlagForCloudLoginSuggestions).Error(),
		},
	}
	for _, s := range suite {
		url, warningMsg, err := validateURL(s.urlIn, s.isCCloud)
		if s.errMsg == "" {
			req.NoError(err)
			req.Equal(s.urlOut, url)
			req.Equal(s.warningMsg, warningMsg)
		} else {
			req.Equal(s.errMsg, err.Error())
		}
	}
}

func newLoginCmd(auth *sdkMock.Auth, user *sdkMock.User, isCloud bool, req *require.Assertions, netrcHandler netrc.NetrcHandler,
	authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager) (*Command, *v1.Config) {
	cfg := v1.New(new(config.Params))
	var mdsClient *mds.APIClient
	if !isCloud {
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
	ccloudClientFactory := &cliMock.MockCCloudClientFactory{
		AnonHTTPClientFactoryFunc: func(baseURL string) *ccloud.Client {
			req.Equal("https://confluent.cloud", baseURL)
			return &ccloud.Client{Params: &ccloud.Params{HttpClient: new(http.Client)}, Auth: auth, User: user}
		},
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloud.Client {
			return &ccloud.Client{Auth: auth, User: user}
		},
	}
	mdsClientManager := &cliMock.MockMDSClientManager{
		GetMDSClientFunc: func(url string, caCertPath string) (client *mds.APIClient, e error) {
			return mdsClient, nil
		},
	}
	prerunner := cliMock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), mdsClient, nil, cfg)
	loginCmd := New(prerunner, ccloudClientFactory, mdsClientManager, cliMock.NewDummyAnalyticsMock(),
		netrcHandler, loginCredentialsManager, authTokenHandler, true)
	return loginCmd, cfg
}

func newLogoutCmd(cfg *v1.Config, netrcHandler netrc.NetrcHandler) (*logout.Command, *v1.Config) {
	logoutCmd := logout.New(cfg, cliMock.NewPreRunnerMock(nil, nil, nil, cfg), cliMock.NewDummyAnalyticsMock(), netrcHandler)
	return logoutCmd, cfg
}

func verifyLoggedInState(t *testing.T, cfg *v1.Config, isCloud bool) {
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
	if isCloud {
		// MDS doesn't set some things like cfg.Auth.User since e.g. an MDS user != an orgv1 (ccloud) User
		req.Equal(&orgv1.User{Id: 23, Email: promptUser, FirstName: "Cody"}, ctx.State.Auth.User)
	} else {
		req.Equal("http://localhost:8090", ctx.Platform.Server)
	}
}

func verifyLoggedOutState(t *testing.T, cfg *v1.Config, loggedOutContext string) {
	req := require.New(t)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}

func TestIsCCloudURL_True(t *testing.T) {
	for _, url := range []string{
		"confluent.cloud",
		"https://confluent.cloud",
		"https://devel.cpdev.cloud/",
		"devel.cpdev.cloud",
		"stag.cpdev.cloud",
		"premium-oryx.gcp.priv.cpdev.cloud",
	} {
		c := new(Command)
		isCCloud := c.isCCloudURL(url)
		require.True(t, isCCloud, url+" should return true")
	}
}

func TestIsCCloudURL_False(t *testing.T) {
	for _, url := range []string{
		"example.com",
		"example.com:8090",
		"https://example.com",
	} {
		c := new(Command)
		isCCloud := c.isCCloudURL(url)
		require.False(t, isCCloud, url+" should return false")
	}
}
