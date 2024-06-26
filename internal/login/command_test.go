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
	"slices"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	mdsmock "github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"

	"github.com/confluentinc/cli/v3/internal/logout"
	climock "github.com/confluentinc/cli/v3/mock"
	pauth "github.com/confluentinc/cli/v3/pkg/auth"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	pmock "github.com/confluentinc/cli/v3/pkg/mock"
	"github.com/confluentinc/cli/v3/pkg/netrc"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

const (
	envUser         = "env-user"
	envPassword     = "env-password"
	testToken1      = "org-1-y0ur.jwt.T0kEn"
	testToken2      = "org-2-y0ur.jwt.T0kEn"
	promptUser      = "prompt-user@confluent.io"
	promptPassword  = " prompt-password "
	netrcFile       = "netrc-file"
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
		GetSsoCredentialsFromConfigFunc: func(_ *config.Config, _ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromConfigFunc: func(_ *config.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetOnPremSsoCredentialsFunc: func(_ string, _ string, _ bool) func() (*pauth.Credentials, error) {
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
	mockNetrcHandler = &pmock.NetrcHandler{
		GetFileNameFunc: func() string {
			return netrcFile
		},
		RemoveNetrcCredentialsFunc: func(_ bool, _ string) (string, error) {
			return "", nil
		},
		CheckCredentialExistFunc: func(_ bool, _ string) (bool, error) {
			return false, nil
		},
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
		GetSsoCredentialsFromConfigFunc: func(_ *config.Config, _ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromConfigFunc: func(_ *config.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
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
	loginCmd, cfg := newLoginCmd(auth, userInterface, true, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)

	output, err := pcmd.ExecuteCommand(loginCmd)
	req.NoError(err)
	req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, envUser))

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
	loginCmd, cfg := newLoginCmd(auth, userInterface, true, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, loginOrganizationManager)

	output, err := pcmd.ExecuteCommand(loginCmd)
	req.NoError(err)
	req.Empty("", output)

	ctx := cfg.Context()
	req.NotNil(ctx)
	req.Equal(pauth.GenerateContextName(promptUser, ccloudURL, ""), ctx.Name)

	req.Equal(testToken2, ctx.GetAuthToken())
	req.Equal(&ccloudv1.User{Id: 23, Email: promptUser, FirstName: "Cody"}, ctx.GetUser())
	verifyLoggedInState(t, cfg, true, organizationId2)
}

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)
	org2 := false
	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken1}, nil
		},
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			org := organizationId1
			if org2 {
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
		if s.orgId != "" {
			org2 = s.orgId == organizationId2
			s.args = append(s.args, "--organization-id", s.orgId)
		}
		loginCmd, cfg := newLoginCmd(auth, userInterface, s.isCloud, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
		output, err := pcmd.ExecuteCommand(loginCmd, s.args...)
		req.NoError(err)
		req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		if s.isCloud && s.orgId == "" {
			s.orgId = organizationId1 // org1Id treated as default org
		}
		verifyLoggedInState(t, cfg, s.isCloud, s.orgId)
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
				GetSsoCredentialsFromConfigFunc: func(_ *config.Config, _ string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCredentialsFromConfigFunc: func(_ *config.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetOnPremSsoCredentialsFunc: func(_ string, _ string, _ bool) func() (*pauth.Credentials, error) {
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
			if test.setNetrcUser {
				loginCredentialsManager.GetCredentialsFromNetrcFunc = func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return netrcCreds, nil
					}
				}
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

			loginCmd, _ := newLoginCmd(mockAuth, mockUserInterface, test.isCloud, req, mockNetrcHandler, AuthTokenHandler, loginCredentialsManager, LoginOrganizationManager)
			var loginArgs []string
			if !test.isCloud {
				loginArgs = []string{"--url", "http://localhost:8090"}
			}
			output, err := pcmd.ExecuteCommand(loginCmd, loginArgs...)
			req.NoError(err)
			req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, test.wantUser))
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
				GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return wrongCreds, nil
					}
				},
				SetCloudClientFunc: func(_ *ccloudv1.Client) {},
			}
			loginCmd, _ := newLoginCmd(mockAuth, mockUserInterface, test.isCloud, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
			loginArgs := []string{"--prompt"}
			if !test.isCloud {
				loginArgs = append(loginArgs, "--url", "http://localhost:8090")
			}
			output, err := pcmd.ExecuteCommand(loginCmd, loginArgs...)
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
	mockLoginCredentialsManager := &climock.LoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, fmt.Errorf("DO NOT RETURN THIS ERR")
			}
		},
		GetSsoCredentialsFromConfigFunc: func(_ *config.Config, _ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, fmt.Errorf("DO NOT RETURN THIS ERR")
			}
		},
		GetCredentialsFromConfigFunc: func(_ *config.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, fmt.Errorf("DO NOT RETURN THIS ERR")
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
	loginCmd, _ := newLoginCmd(mockAuth, mockUserInterface, true, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setEnv {
				os.Setenv(pauth.ConfluentPlatformCACertPath, "testcert.pem")
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
			_, err := pcmd.ExecuteCommand(loginCmd, "--url", "http://localhost:8090", "--ca-cert-path", test.caCertPathFlag)
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
			expectedCaCertPath: originalCaCertPath,
		},
		{
			name:              "reset ca-cert-path",
			useCaCertPathFlag: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			legacyContextName := "login-prompt-user@confluent.io-http://localhost:8090"
			cfg := config.AuthenticatedConfigMockWithContextName(legacyContextName)
			cfg.Contexts[legacyContextName].Platform.CaCertPath = originalCaCertPath

			loginCmd := getNewLoginCommandForSelfSignedCertTest(req, cfg, test.expectedCaCertPath)
			args := []string{"--url", "http://localhost:8090"}
			if test.useCaCertPathFlag {
				args = append(args, "--ca-cert-path", "")
			}
			fmt.Println(args)
			_, err := pcmd.ExecuteCommand(loginCmd, args...)
			req.NoError(err)

			ctx := cfg.Context()
			// ensure the right CaCertPath is stored in Config and context name is the right name
			req.Equal(test.expectedCaCertPath, ctx.Platform.CaCertPath)
			req.Equal(legacyContextName, ctx.Name)
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
	priv, err := rsa.GenerateKey(rand.Reader, 512)
	req.NoError(err, "Couldn't generate private key")
	certBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)
	req.NoError(err, "Couldn't generate certificate from private key")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certReader := bytes.NewReader(pemBytes)

	cert, err := x509.ParseCertificate(certBytes)
	req.NoError(err, "Couldn't reparse certificate")
	mdsClient.TokensAndAuthenticationApi = &mdsmock.TokensAndAuthenticationApi{
		GetTokenFunc: func(ctx context.Context) (mdsv1.AuthenticationResponse, *http.Response, error) {
			req.NotEqual(http.DefaultClient, mdsClient)
			transport, ok := mdsClient.GetConfig().HTTPClient.Transport.(*http.Transport)
			req.True(ok)
			req.NotEqual(http.DefaultTransport, transport)
			found := slices.ContainsFunc(transport.TLSClientConfig.RootCAs.Subjects(), func(subject []byte) bool { //nolint
				return bytes.Equal(cert.RawSubject, subject)
			})
			req.True(found, "Certificate not found in client.")
			return mdsv1.AuthenticationResponse{
				AuthToken: testToken1,
				TokenType: "JWT",
				ExpiresIn: 100,
			}, nil, nil
		},
	}
	mdsClientManager := &climock.MDSClientManager{
		GetMDSClientFunc: func(_, caCertPath string, _ bool) (*mdsv1.APIClient, error) {
			// ensure the right caCertPath is used
			req.Contains(caCertPath, expectedCaCertPath)
			mdsClient.GetConfig().HTTPClient, err = utils.SelfSignedCertClient(certReader, tls.Certificate{})
			if err != nil {
				return nil, err
			}
			return mdsClient, nil
		},
	}
	loginCmd := New(cfg, prerunner, nil, mdsClientManager, mockNetrcHandler, mockLoginCredentialsManager, LoginOrganizationManager, AuthTokenHandler)
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
		loginCmd, cfg := newLoginCmd(auth, userInterface, s.isCloud, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)

		// Login to the CLI control plane
		_, err := pcmd.ExecuteCommand(loginCmd, s.args...)
		req.NoError(err)
		verifyLoggedInState(t, cfg, s.isCloud, organizationId1)

		// Set kafka related states for the logged in context
		ctx := cfg.Context()
		ctx.KafkaClusterContext.AddKafkaClusterConfig(kafkaCluster)
		ctx.KafkaClusterContext.SetActiveKafkaCluster(kafkaCluster.ID)

		// Executing logout
		logoutCmd, _ := newLogoutCmd(cfg, mockNetrcHandler)
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

func newLoginCmd(auth *ccloudv1mock.Auth, userInterface *ccloudv1mock.UserInterface, isCloud bool, req *require.Assertions, netrcHandler netrc.NetrcHandler, authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager, loginOrganizationManager pauth.LoginOrganizationManager) (*cobra.Command, *config.Config) {
	config.SetTempHomeDir()
	cfg := config.New()
	var mdsClient *mdsv1.APIClient
	if !isCloud {
		mdsConfig := mdsv1.NewConfiguration()
		mdsClient = mdsv1.NewAPIClient(mdsConfig)
		mdsClient.TokensAndAuthenticationApi = &mdsmock.TokensAndAuthenticationApi{
			GetTokenFunc: func(ctx context.Context) (mdsv1.AuthenticationResponse, *http.Response, error) {
				return mdsv1.AuthenticationResponse{
					AuthToken: testToken1,
					TokenType: "JWT",
					ExpiresIn: 100,
				}, nil, nil
			},
		}
	}
	ccloudClientFactory := &climock.CCloudClientFactory{
		AnonHTTPClientFactoryFunc: func(baseURL string) *ccloudv1.Client {
			req.Equal("https://confluent.cloud", baseURL)
			return &ccloudv1.Client{Params: &ccloudv1.Params{HttpClient: new(http.Client)}, Auth: auth, User: userInterface}
		},
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloudv1.Client {
			return &ccloudv1.Client{Growth: &ccloudv1mock.Growth{
				GetFreeTrialInfoFunc: func(_ int32) ([]*ccloudv1.GrowthPromoCodeClaim, error) {
					return []*ccloudv1.GrowthPromoCodeClaim{}, nil
				},
			}, Auth: auth, User: userInterface}
		},
	}
	mdsClientManager := &climock.MDSClientManager{
		GetMDSClientFunc: func(_, _ string, _ bool) (*mdsv1.APIClient, error) {
			return mdsClient, nil
		},
	}
	prerunner := climock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, mdsClient, nil, cfg)
	loginCmd := New(cfg, prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, loginOrganizationManager, authTokenHandler)
	loginCmd.Flags().Bool("unsafe-trace", false, "")
	return loginCmd, cfg
}

func newLogoutCmd(cfg *config.Config, netrcHandler netrc.NetrcHandler) (*cobra.Command, *config.Config) {
	logoutCmd := logout.New(cfg, climock.NewPreRunnerMock(nil, nil, nil, nil, cfg), netrcHandler)
	return logoutCmd, cfg
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
