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
	"reflect"
	"slices"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	mdsmock "github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"

	"github.com/confluentinc/cli/v4/internal/logout"
	climock "github.com/confluentinc/cli/v4/mock"
	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const (
	envUser         = "envvar-user"
	envPassword     = "env-password"
	testToken1      = "org-1-y0ur.jwt.T0kEn"
	testToken2      = "org-2-y0ur.jwt.T0kEn"
	promptUser      = "prompt-user@confluent.io"
	promptPassword  = " prompt-password "
	ccloudURL       = "https://confluent.cloud"
	organizationId1 = "o-001"
	organizationId2 = "o-002"
	refreshToken    = "refreshToken"
)

var (
	envCreds = &pauth.Credentials{
		Username: envUser,
		Password: envPassword,
	}
	mockAuth = &ccloudv1mock.Auth{
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User:         &ccloudv1.User{Id: 23},
				Organization: &ccloudv1.Organization{ResourceId: organizationId1},
				Accounts:     []*ccloudv1.Account{{Id: "env-596", Name: "Default"}},
			}, nil
		},
	}
	mockUserInterface           = &ccloudv1mock.UserInterface{}
	mockLoginCredentialsManager = &climock.LoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},
		GetOnPremCredentialsFromEnvVarFunc: func() func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetOnPremCredentialsFromPromptFunc: func() func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},
		GetCredentialsFromConfigFunc: func(_ *config.Config, _ config.MachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetOnPremSsoCredentialsFunc: func(_, _, _, _ string, _ bool) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetOnPremSsoCredentialsFromConfigFunc: func(_ *config.Config, _ bool) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromKeychainFunc: func(_ bool, _, _ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(_ *ccloudv1.Client) {},
	}
	LoginOrganizationManager = &climock.LoginOrganizationManager{
		GetLoginOrganizationFromFlagFunc: func(cmd *cobra.Command) func() string {
			return pauth.NewLoginOrganizationManagerImpl().GetLoginOrganizationFromFlag(cmd)
		},
		GetLoginOrganizationFromEnvironmentVariableFunc: func() func() string {
			return pauth.NewLoginOrganizationManagerImpl().GetLoginOrganizationFromEnvironmentVariable()
		},
	}
	AuthTokenHandler = &climock.AuthTokenHandler{
		GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, credentials *pauth.Credentials, noBrowser bool, organizationId string) (string, string, error) {
			if organizationId == "" || organizationId == organizationId1 {
				return testToken1, refreshToken, nil
			} else if organizationId == organizationId2 {
				return testToken2, refreshToken, nil
			} else {
				return "", "", &ccloudv1.Error{Message: "invalid user", Code: http.StatusUnauthorized}
			}
		},
		GetConfluentTokenFunc: func(_ *mdsv1.APIClient, _ *pauth.Credentials, _ bool) (string, string, error) {
			return testToken1, "", nil
		},
	}
	mockAuthResponse = mdsv1.AuthenticationResponse{
		AuthToken: testToken1,
		TokenType: "JWT",
		ExpiresIn: 100,
	}
)

func TestCredentialsOverride(t *testing.T) {
	req := require.New(t)
	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken1}, nil
		},
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     envUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: organizationId1},
				Accounts:     []*ccloudv1.Account{{Id: "env-596", Name: "Default"}},
			}, nil
		},
	}
	userInterface := &ccloudv1mock.UserInterface{}
	mockLoginCredentialsManager := &climock.LoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return envCreds, nil
			}
		},
		GetCredentialsFromConfigFunc: func(_ *config.Config, _ config.MachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromKeychainFunc: func(_ bool, _, _ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(_ *ccloudv1.Client) {},
	}
	loginCmd, cfg := newLoginCmd(auth, userInterface, true, req, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)

	output, err := pcmd.ExecuteCommand(loginCmd)
	req.NoError(err)
	req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, envUser, organizationId1, ""))

	ctx := cfg.Context()
	req.NotNil(ctx)
	req.Equal(pauth.GenerateContextName(envUser, ccloudURL, ""), ctx.Name)

	req.Equal(testToken1, ctx.GetAuthToken())
	req.Equal(&ccloudv1.User{Id: 23, Email: envUser, FirstName: "Cody"}, ctx.GetUser())
}

func TestOrgIdOverride(t *testing.T) {
	req := require.New(t)
	auth := &ccloudv1mock.Auth{
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: organizationId2},
				Accounts:     []*ccloudv1.Account{{Id: "env-596", Name: "Default"}},
			}, nil
		},
	}
	userInterface := &ccloudv1mock.UserInterface{}

	loginOrganizationManager := &climock.LoginOrganizationManager{
		GetLoginOrganizationFromFlagFunc: LoginOrganizationManager.GetLoginOrganizationFromFlagFunc,
		GetLoginOrganizationFromEnvironmentVariableFunc: func() func() string {
			return func() string { return organizationId2 }
		},
	}
	loginCmd, cfg := newLoginCmd(auth, userInterface, true, req, AuthTokenHandler, mockLoginCredentialsManager, loginOrganizationManager)

	output, err := pcmd.ExecuteCommand(loginCmd)
	req.NoError(err)
	req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, promptUser, organizationId2, ""))

	ctx := cfg.Context()
	req.NotNil(ctx)
	req.Equal(pauth.GenerateContextName(promptUser, ccloudURL, ""), ctx.Name)

	req.Equal(testToken2, ctx.GetAuthToken())
	req.Equal(&ccloudv1.User{Id: 23, Email: promptUser, FirstName: "Cody"}, ctx.GetUser())
	verifyLoggedInState(t, cfg, true, organizationId2)
}

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)
	useOrgTwo := false
	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken1}, nil
		},
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			org := organizationId1
			if useOrgTwo {
				org = organizationId2
			}
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: org},
				Accounts:     []*ccloudv1.Account{{Id: "env-596", Name: "Default"}},
			}, nil
		},
	}
	userInterface := &ccloudv1mock.UserInterface{}

	suite := []struct {
		isCloud bool
		args    []string
		orgId   string
		setEnv  bool
	}{
		{
			isCloud: true,
			args:    []string{},
		},
		{
			args: []string{"--url", "http://localhost:8090"},
		},
		{
			isCloud: true,
			orgId:   organizationId1,
		},
		{
			isCloud: true,
			orgId:   organizationId2,
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
		if s.isCloud && s.orgId == "" {
			s.orgId = organizationId1 // org1Id treated as default org
		} else if s.orgId != "" {
			useOrgTwo = s.orgId == organizationId2
			s.args = append(s.args, "--organization", s.orgId)
		}

		loginCmd, cfg := newLoginCmd(auth, userInterface, s.isCloud, req, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
		output, err := pcmd.ExecuteCommand(loginCmd, s.args...)
		req.NoError(err)
		if s.isCloud {
			req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, promptUser, s.orgId, ""))
		} else {
			req.Empty(output) // on-prem has no login message
		}

		verifyLoggedInState(t, cfg, s.isCloud, s.orgId)
		if s.setEnv {
			_ = os.Unsetenv(pauth.ConfluentPlatformMDSURL)
		}
	}
}

func TestLoginOrderOfPrecedence(t *testing.T) {
	req := require.New(t)

	tests := []struct {
		name      string
		isCloud   bool
		setEnvVar bool
		wantUser  string
	}{
		{
			name:      "cloud env var over all other credentials",
			isCloud:   true,
			setEnvVar: true,
			wantUser:  envUser,
		},
		{
			name:      "cloud prompt",
			isCloud:   true,
			setEnvVar: false,
			wantUser:  promptUser,
		},
		{
			name:      "on-prem env var over all other credentials",
			setEnvVar: true,
			wantUser:  envUser,
		},
		{
			name:      "on-prem prompt",
			setEnvVar: false,
			wantUser:  promptUser,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loginCredentialsManager := &climock.LoginCredentialsManager{
				GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCloudCredentialsFromPromptFunc: func(_ string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{
							Username: promptUser,
							Password: promptPassword,
						}, nil
					}
				},
				GetOnPremCredentialsFromEnvVarFunc: func() func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetOnPremCredentialsFromPromptFunc: func() func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{
							Username: promptUser,
							Password: promptPassword,
						}, nil
					}
				},
				GetCredentialsFromConfigFunc: func(_ *config.Config, _ config.MachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetOnPremSsoCredentialsFunc: func(_, _, _, _ string, _ bool) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetOnPremSsoCredentialsFromConfigFunc: func(_ *config.Config, _ bool) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCredentialsFromKeychainFunc: func(_ bool, _, _ string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				SetCloudClientFunc: func(_ *ccloudv1.Client) {},
			}
			if test.isCloud {
				if test.setEnvVar {
					loginCredentialsManager.GetCloudCredentialsFromEnvVarFunc = func(_ string) func() (*pauth.Credentials, error) {
						return func() (*pauth.Credentials, error) {
							return envCreds, nil
						}
					}
				}
			} else {
				if test.setEnvVar {
					loginCredentialsManager.GetOnPremCredentialsFromEnvVarFunc = func() func() (*pauth.Credentials, error) {
						return func() (*pauth.Credentials, error) {
							return envCreds, nil
						}
					}
				}
			}

			loginCmd, _ := newLoginCmd(mockAuth, mockUserInterface, test.isCloud, req, AuthTokenHandler, loginCredentialsManager, LoginOrganizationManager)
			var loginArgs []string
			if !test.isCloud {
				loginArgs = []string{"--url", "http://localhost:8090"}
			}
			output, err := pcmd.ExecuteCommand(loginCmd, loginArgs...)
			req.NoError(err)
			if test.isCloud {
				user := promptUser
				if test.setEnvVar {
					user = envUser
				}
				req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, user, organizationId1, ""))
			} else {
				req.Empty(output) // on-prem has no login message
			}
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLoginCredentialsManager := &climock.LoginCredentialsManager{
				GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return wrongCreds, nil
					}
				},
				GetCloudCredentialsFromPromptFunc: func(_ string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{
							Username: promptUser,
							Password: promptPassword,
						}, nil
					}
				},
				GetOnPremCredentialsFromEnvVarFunc: func() func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return wrongCreds, nil
					}
				},
				GetOnPremCredentialsFromPromptFunc: func() func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{
							Username: promptUser,
							Password: promptPassword,
						}, nil
					}
				},
				SetCloudClientFunc: func(_ *ccloudv1.Client) {},
			}
			loginCmd, _ := newLoginCmd(mockAuth, mockUserInterface, test.isCloud, req, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
			loginArgs := []string{"--prompt"}
			if !test.isCloud {
				loginArgs = append(loginArgs, "--url", "http://localhost:8090")
			}
			output, err := pcmd.ExecuteCommand(loginCmd, loginArgs...)
			req.NoError(err)

			req.False(mockLoginCredentialsManager.GetCloudCredentialsFromEnvVarCalled())
			req.False(mockLoginCredentialsManager.GetOnPremCredentialsFromEnvVarCalled())

			if test.isCloud {
				req.Contains(output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, promptUser, organizationId1, ""))
			} else {
				req.Empty(output) // on-prem has no login message
			}
		})
	}
}

func TestLoginFail(t *testing.T) {
	req := require.New(t)
	mockLoginCredentialsManager := &climock.LoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, fmt.Errorf("DO NOT RETURN THIS ERR")
			}
		},
		GetCredentialsFromConfigFunc: func(_ *config.Config, _ config.MachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, &ccloudv1.InvalidLoginError{}
			}
		},
		GetCredentialsFromKeychainFunc: func(_ bool, _, _ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(_ *ccloudv1.Client) {},
	}
	loginCmd, _ := newLoginCmd(mockAuth, mockUserInterface, true, req, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
	_, err := pcmd.ExecuteCommand(loginCmd)
	req.Error(err)
	req.Equal(new(ccloudv1.InvalidLoginError), err)
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
			name:                "specified certificate-authority-path",
			caCertPathFlag:      "testcert.pem",
			expectedContextName: "login-prompt-user@confluent.io-http://localhost:8090?cacertpath=%s",
		},
		{
			name:                "no certificate-authority-path flag",
			caCertPathFlag:      "",
			expectedContextName: "login-prompt-user@confluent.io-http://localhost:8090",
		},
		{
			name:                "env var certificate-authority-path flag",
			setEnv:              true,
			envCertPath:         "testcert.pem",
			expectedContextName: "login-prompt-user@confluent.io-http://localhost:8090?cacertpath=%s",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setEnv {
				os.Setenv(pauth.ConfluentPlatformCertificateAuthorityPath, "testcert.pem")
			}
			config.SetTempHomeDir()
			cfg := config.New()
			var expectedCaCert string
			if test.setEnv {
				expectedCaCert = test.envCertPath
			} else {
				expectedCaCert = test.caCertPathFlag
			}
			loginCmd := getNewLoginCommandForSelfSignedCertTest(req, cfg, expectedCaCert)
			_, err := pcmd.ExecuteCommand(loginCmd, "--url", "http://localhost:8090", "--certificate-authority-path", test.caCertPathFlag)
			req.NoError(err)

			ctx := cfg.Context()

			if test.setEnv {
				req.Contains(ctx.Platform.CaCertPath, test.envCertPath)
			} else {
				// ensure the right CaCertPath is stored in Config
				req.Contains(ctx.Platform.CaCertPath, test.caCertPathFlag)
			}

			if test.caCertPathFlag != "" || test.envCertPath != "" {
				req.Equal(fmt.Sprintf(test.expectedContextName, ctx.Platform.CaCertPath), ctx.Name)
			} else {
				req.Equal(test.expectedContextName, ctx.Name)
			}
			if test.setEnv {
				os.Unsetenv(pauth.ConfluentPlatformCertificateAuthorityPath)
			}
		})
	}
}

func getNewLoginCommandForSelfSignedCertTest(req *require.Assertions, cfg *config.Config, expectedCaCertPath string) *cobra.Command {
	mdsConfig := mdsv1.NewConfiguration()
	mdsClient := mdsv1.NewAPIClient(mdsConfig)

	prerunner := climock.NewPreRunnerMock(nil, nil, nil, nil, cfg)

	// Create a test certificate to be read in by the command
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject:      pkix.Name{Organization: []string{"testorg"}},
	}
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	req.NoError(err, "Couldn't generate private key")
	certBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)
	req.NoError(err, "Couldn't generate certificate from private key")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certReader := bytes.NewReader(pemBytes)

	cert, err := x509.ParseCertificate(certBytes)
	req.NoError(err, "Couldn't reparse certificate")
	mdsClient.TokensAndAuthenticationApi = &mdsmock.TokensAndAuthenticationApi{
		GetTokenFunc: func(_ context.Context) (mdsv1.AuthenticationResponse, *http.Response, error) {
			req.NotEqual(http.DefaultClient, mdsClient)
			transport, ok := mdsClient.GetConfig().HTTPClient.Transport.(*http.Transport)
			req.True(ok)
			req.NotEqual(http.DefaultTransport, transport)
			found := slices.ContainsFunc(transport.TLSClientConfig.RootCAs.Subjects(), func(subject []byte) bool { //nolint
				return bytes.Equal(cert.RawSubject, subject)
			})
			req.True(found, "Certificate not found in client.")
			return mockAuthResponse, nil, nil
		},
	}
	mdsClientManager := &climock.MDSClientManager{
		GetMDSClientFunc: func(_, caCertPath, _, _ string, _ bool) (*mdsv1.APIClient, error) {
			// ensure the right caCertPath is used
			req.Contains(caCertPath, expectedCaCertPath)
			mdsClient.GetConfig().HTTPClient, err = utils.SelfSignedCertClient(certReader, tls.Certificate{})
			if err != nil {
				return nil, err
			}
			return mdsClient, nil
		},
	}
	loginCmd := New(cfg, prerunner, nil, mdsClientManager, mockLoginCredentialsManager, LoginOrganizationManager, AuthTokenHandler)
	loginCmd.Flags().Bool("unsafe-trace", false, "")
	loginCmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity")

	return loginCmd
}

func TestLoginWithExistingContext(t *testing.T) {
	req := require.New(t)
	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken1}, nil
		},
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: organizationId1},
				Accounts:     []*ccloudv1.Account{{Id: "env-596", Name: "Default"}},
			}, nil
		},
		LogoutFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken1}, nil
		},
	}
	userInterface := &ccloudv1mock.UserInterface{}

	suite := []struct {
		isCloud bool
		args    []string
	}{
		{
			isCloud: true,
			args:    []string{},
		},
		{
			args: []string{"--url", "http://localhost:8090"},
		},
	}

	activeApiKey := "bo"
	kafkaCluster := &config.KafkaClusterConfig{
		ID:        "lkc-0000",
		Name:      "bob",
		Bootstrap: "http://bobby",
		APIKeys: map[string]*config.APIKeyPair{
			activeApiKey: {
				Key:    activeApiKey,
				Secret: "bo",
			},
		},
		APIKey: activeApiKey,
	}

	for _, s := range suite {
		loginCmd, cfg := newLoginCmd(auth, userInterface, s.isCloud, req, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)

		// Login to the CLI control plane
		_, err := pcmd.ExecuteCommand(loginCmd, s.args...)
		req.NoError(err)
		verifyLoggedInState(t, cfg, s.isCloud, organizationId1)

		// Set kafka related states for the logged in context
		ctx := cfg.Context()
		ctx.KafkaClusterContext.AddKafkaClusterConfig(kafkaCluster)
		ctx.KafkaClusterContext.SetActiveKafkaCluster(kafkaCluster.ID)

		// Executing logout
		logoutCmd := newLogoutCmd(auth, userInterface, s.isCloud, req, AuthTokenHandler, cfg)
		_, err = pcmd.ExecuteCommand(logoutCmd)
		req.NoError(err)
		verifyLoggedOutState(t, cfg, ctx.Name)

		// logging back in the same context
		_, err = pcmd.ExecuteCommand(loginCmd, s.args...)
		req.NoError(err)
		verifyLoggedInState(t, cfg, s.isCloud, organizationId1)

		// verify that kafka cluster info persists between logging back in again
		req.Equal(kafkaCluster.ID, ctx.KafkaClusterContext.GetActiveKafkaClusterId())
		reflect.DeepEqual(kafkaCluster, ctx.KafkaClusterContext.GetKafkaClusterConfig(kafkaCluster.ID))
	}
}

func TestSuspendedOrganizationError(t *testing.T) {
	req := require.New(t)

	mockLoginCredentialsManager := &climock.LoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return &pauth.Credentials{
					Username: promptUser,
					Password: promptPassword,
				}, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(_ *ccloudv1.Client) {},
	}

	mockAuthTokenHandler := &climock.AuthTokenHandler{
		GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, _ *pauth.Credentials, _ bool, _ string) (string, string, error) {
			return "", "", &ccloudv1.SuspendedOrganizationError{}
		},
	}

	auth := &ccloudv1mock.Auth{
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:    23,
					Email: promptUser,
				},
				Organization: &ccloudv1.Organization{ResourceId: organizationId1},
				Accounts:     []*ccloudv1.Account{{Id: "env-596", Name: "Default"}},
			}, nil
		},
	}
	userInterface := &ccloudv1mock.UserInterface{}

	loginCmd, _ := newLoginCmd(auth, userInterface, true, req, mockAuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
	_, err := pcmd.ExecuteCommand(loginCmd)
	req.Error(err)
	req.IsType(&errors.ErrorWithSuggestionsImpl{}, err)
	req.Contains(err.Error(), "suspended")
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
			errMsg:     errors.InvalidLoginURLErrorMsg,
		},
		{
			urlIn:      "test.com",
			urlOut:     "https://test.com:8090",
			warningMsg: "https protocol and default MDS port 8090",
		},
		{
			urlIn:      "test.com:80",
			urlOut:     "https://test.com:80",
			warningMsg: "https protocol",
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
			urlOut:     "https://127.0.0.1:8090",
			warningMsg: "https protocol and default MDS port 8090",
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
			errMsg:   errors.NewErrorWithSuggestions("there is no need to pass the `--url` flag if you are logging in to Confluent Cloud", "Log in to Confluent Cloud with `confluent login`.").Error(),
		},
		{
			urlIn:    "https://confluent.cloud/login/sso/company",
			isCCloud: true,
			errMsg:   errors.NewErrorWithSuggestions("there is no need to pass the `--url` flag if you are logging in to Confluent Cloud", "Log in to Confluent Cloud with `confluent login`.").Error(),
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

func newLoginCmd(auth *ccloudv1mock.Auth, userInterface *ccloudv1mock.UserInterface, isCloud bool, req *require.Assertions, authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager, loginOrganizationManager pauth.LoginOrganizationManager) (*cobra.Command, *config.Config) {
	config.SetTempHomeDir()
	cfg := config.New()
	var ccloudClientFactory *climock.CCloudClientFactory
	var mdsClientManager *climock.MDSClientManager
	var prerunner pcmd.PreRunner

	if !isCloud {
		mdsClient := climock.NewMdsClientMock(testToken1)
		mdsClientManager = &climock.MDSClientManager{
			GetMDSClientFunc: func(_, _, _, _ string, _ bool) (*mdsv1.APIClient, error) {
				return mdsClient, nil
			},
		}
		prerunner = climock.NewPreRunnerMock(nil, nil, mdsClient, nil, cfg)
	} else {
		ccloudClientFactory = climock.NewCCloudClientFactoryMock(auth, userInterface, req)
		prerunner = climock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, nil, nil, cfg)
	}

	loginCmd := New(cfg, prerunner, ccloudClientFactory, mdsClientManager, loginCredentialsManager, loginOrganizationManager, authTokenHandler)
	loginCmd.Flags().Bool("unsafe-trace", false, "")
	return loginCmd, cfg
}

func newLogoutCmd(auth *ccloudv1mock.Auth, userInterface *ccloudv1mock.UserInterface, isCloud bool, req *require.Assertions, authTokenHandler pauth.AuthTokenHandler, cfg *config.Config) *cobra.Command {
	var prerunner pcmd.PreRunner

	if isCloud {
		ccloudClientFactory := climock.NewCCloudClientFactoryMock(auth, userInterface, req)
		prerunner = climock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, nil, nil, cfg)
	} else {
		mdsClient := climock.NewMdsClientMock(testToken1)
		prerunner = climock.NewPreRunnerMock(nil, nil, mdsClient, nil, cfg)
	}
	return logout.New(cfg, prerunner, authTokenHandler)
}

func verifyLoggedInState(t *testing.T, cfg *config.Config, isCloud bool, organizationId string) {
	req := require.New(t)
	ctx := cfg.Context()
	req.NotNil(ctx)
	if organizationId == organizationId1 || organizationId == "" {
		req.Equal(testToken1, ctx.GetAuthToken())
	} else if organizationId == organizationId2 {
		req.Equal(testToken2, ctx.GetAuthToken())
	}
	contextName := fmt.Sprintf("login-%s-%s", promptUser, ctx.GetPlatformServer())
	credName := fmt.Sprintf("username-%s", ctx.Credential.Username)
	req.Contains(cfg.Platforms, ctx.GetPlatform().GetName())
	req.Equal(ctx.Platform, cfg.Platforms[ctx.GetPlatform().GetName()])
	req.Contains(cfg.Credentials, credName)
	req.Equal(promptUser, cfg.Credentials[credName].Username)
	req.Contains(cfg.Contexts, contextName)
	req.Equal(ctx.GetPlatform(), cfg.Contexts[contextName].GetPlatform())
	req.Equal(ctx.Credential, cfg.Contexts[contextName].Credential)
	if isCloud {
		// MDS doesn't set some things like cfg.Auth.User since e.g. an MDS user != an ccloudv1 User
		req.Equal(&ccloudv1.User{Id: 23, Email: promptUser, FirstName: "Cody"}, ctx.GetUser())
		req.Equal(organizationId, ctx.GetCurrentOrganization())
	} else {
		req.Equal("http://localhost:8090", ctx.GetPlatformServer())
	}
}

func verifyLoggedOutState(t *testing.T, cfg *config.Config, loggedOutContext string) {
	req := require.New(t)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}
