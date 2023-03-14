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

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	mdsmock "github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"

	"github.com/confluentinc/cli/internal/cmd/logout"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
	climock "github.com/confluentinc/cli/mock"
)

const (
	envUser        = "env-user"
	envPassword    = "env-password"
	testToken1     = "org-1-y0ur.jwt.T0kEn"
	testToken2     = "org-2-y0ur.jwt.T0kEn"
	promptUser     = "prompt-user@confluent.io"
	promptPassword = " prompt-password "
	netrcFile      = "netrc-file"
	ccloudURL      = "https://confluent.cloud"
	org1Id         = "o-001"
	org2Id         = "o-002"
	refreshToken   = "refreshToken"
)

var (
	envCreds = &pauth.Credentials{
		Username: envUser,
		Password: envPassword,
	}
	mockAuth = &ccloudv1mock.Auth{
		UserFunc: func(_ context.Context) (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     "",
					FirstName: "",
				},
				Organization: &ccloudv1.Organization{ResourceId: org1Id},
				Accounts:     []*ccloudv1.Account{{Id: "a-595", Name: "Default"}},
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
		GetSsoCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromConfigFunc: func(_ *v1.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromKeychainFunc: func(_ *v1.Config, _ bool, _, _ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		SetCloudClientFunc: func(arg0 *ccloudv1.Client) {
		},
	}
	LoginOrganizationManager = &climock.LoginOrganizationManager{
		GetLoginOrganizationFromArgsFunc: func(cmd *cobra.Command) func() (string, error) {
			return pauth.NewLoginOrganizationManagerImpl().GetLoginOrganizationFromArgs(cmd)
		},
		GetLoginOrganizationFromEnvVarFunc: func(cmd *cobra.Command) func() (string, error) {
			return pauth.NewLoginOrganizationManagerImpl().GetLoginOrganizationFromEnvVar(cmd)
		},
		GetDefaultLoginOrganizationFunc: func() func() (string, error) {
			return pauth.NewLoginOrganizationManagerImpl().GetDefaultLoginOrganization()
		},
	}
	AuthTokenHandler = &climock.AuthTokenHandler{
		GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, credentials *pauth.Credentials, noBrowser bool, orgResourceId string) (string, string, error) {
			if orgResourceId == "" || orgResourceId == org1Id {
				return testToken1, refreshToken, nil
			} else if orgResourceId == org2Id {
				return testToken2, refreshToken, nil
			} else {
				return "", "", &ccloudv1.Error{Message: "invalid user", Code: http.StatusUnauthorized}
			}
		},
		GetConfluentTokenFunc: func(mdsClient *mds.APIClient, credentials *pauth.Credentials) (s string, e error) {
			return testToken1, nil
		},
	}
	mockNetrcHandler = &pmock.NetrcHandler{
		GetFileNameFunc: func() string { return netrcFile },
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
	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ context.Context, _ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken1}, nil
		},
		UserFunc: func(_ context.Context) (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     envUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: org1Id},
				Accounts:     []*ccloudv1.Account{{Id: "a-595", Name: "Default"}},
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
		GetSsoCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromConfigFunc: func(_ *v1.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
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
		GetCredentialsFromKeychainFunc: func(_ *v1.Config, _ bool, _, _ string) func() (*pauth.Credentials, error) {
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
		UserFunc: func(ctx context.Context) (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: org2Id},
				Accounts:     []*ccloudv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	userInterface := &ccloudv1mock.UserInterface{}
	type test struct {
		setEnv     bool
		setDefault bool
	}
	tests := []*test{
		{setEnv: true},
		{setDefault: true},
	}

	for _, tt := range tests {
		loginOrganizationManager := &climock.LoginOrganizationManager{
			GetLoginOrganizationFromArgsFunc: LoginOrganizationManager.GetLoginOrganizationFromArgsFunc,
			GetLoginOrganizationFromEnvVarFunc: func(cmd *cobra.Command) func() (string, error) {
				if tt.setEnv {
					return func() (string, error) { return org2Id, nil }
				} else {
					return LoginOrganizationManager.GetLoginOrganizationFromEnvVarFunc(cmd)
				}
			},
			GetDefaultLoginOrganizationFunc: func() func() (string, error) {
				if tt.setDefault {
					return func() (string, error) { return org2Id, nil }
				} else {
					return LoginOrganizationManager.GetDefaultLoginOrganization()
				}
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
		verifyLoggedInState(t, cfg, true, org2Id)
	}
}

func TestLoginSuccess(t *testing.T) {
	req := require.New(t)
	org2 := false
	auth := &ccloudv1mock.Auth{
		LoginFunc: func(_ context.Context, _ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken1}, nil
		},
		UserFunc: func(_ context.Context) (*ccloudv1.GetMeReply, error) {
			org := org1Id
			if org2 {
				org = org2Id
			}
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: org},
				Accounts:     []*ccloudv1.Account{{Id: "a-595", Name: "Default"}},
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
			args: []string{"--url=http://localhost:8090"},
		},
		{
			isCloud: true,
			orgId:   org1Id,
		},
		{
			isCloud: true,
			orgId:   org2Id,
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
			org2 = s.orgId == org2Id
			s.args = append(s.args, "--organization-id="+s.orgId)
		}
		loginCmd, cfg := newLoginCmd(auth, userInterface, s.isCloud, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
		output, err := pcmd.ExecuteCommand(loginCmd, s.args...)
		req.NoError(err)
		req.NotContains(output, fmt.Sprintf(errors.LoggedInAsMsg, promptUser))
		if s.isCloud && s.orgId == "" {
			s.orgId = org1Id // org1Id treated as default org
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				GetSsoCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCredentialsFromConfigFunc: func(_ *v1.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCredentialsFromKeychainFunc: func(_ *v1.Config, _ bool, _, _ string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				SetCloudClientFunc: func(_ *ccloudv1.Client) {},
			}
			if tt.setNetrcUser {
				loginCredentialsManager.GetCredentialsFromNetrcFunc = func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return netrcCreds, nil
					}
				}
			}
			if tt.isCloud {
				if tt.setEnvVar {
					loginCredentialsManager.GetCloudCredentialsFromEnvVarFunc = func(orgResourceId string) func() (*pauth.Credentials, error) {
						return func() (*pauth.Credentials, error) {
							return envCreds, nil
						}
					}
				}
			} else {
				if tt.setEnvVar {
					loginCredentialsManager.GetOnPremCredentialsFromEnvVarFunc = func() func() (*pauth.Credentials, error) {
						return func() (*pauth.Credentials, error) {
							return envCreds, nil
						}
					}
				}
			}

			loginCmd, _ := newLoginCmd(mockAuth, mockUserInterface, tt.isCloud, req, mockNetrcHandler, AuthTokenHandler, loginCredentialsManager, LoginOrganizationManager)
			var loginArgs []string
			if !tt.isCloud {
				loginArgs = []string{"--url=http://localhost:8090"}
			}
			output, err := pcmd.ExecuteCommand(loginCmd, loginArgs...)
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
				SetCloudClientFunc: func(arg0 *ccloudv1.Client) {
				},
			}
			loginCmd, _ := newLoginCmd(mockAuth, mockUserInterface, tt.isCloud, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
			loginArgs := []string{"--prompt"}
			if !tt.isCloud {
				loginArgs = append(loginArgs, "--url=http://localhost:8090")
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
				return nil, errors.New("DO NOT RETURN THIS ERR")
			}
		},
		GetSsoCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, errors.New("DO NOT RETURN THIS ERR")
			}
		},
		GetCredentialsFromConfigFunc: func(_ *v1.Config, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, errors.New("DO NOT RETURN THIS ERR")
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, &ccloudv1.InvalidLoginError{}
			}
		},
		GetCredentialsFromKeychainFunc: func(_ *v1.Config, _ bool, _, _ string) func() (*pauth.Credentials, error) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(pauth.ConfluentPlatformCACertPath, "testcert.pem")
			}
			cfg := v1.New()
			var expectedCaCert string
			if tt.setEnv {
				expectedCaCert = tt.envCertPath
			} else {
				expectedCaCert = tt.caCertPathFlag
			}
			loginCmd := getNewLoginCommandForSelfSignedCertTest(req, cfg, expectedCaCert)
			_, err := pcmd.ExecuteCommand(loginCmd, "--url=http://localhost:8090", fmt.Sprintf("--ca-cert-path=%s", tt.caCertPathFlag))
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
			_, err := pcmd.ExecuteCommand(loginCmd, args...)
			req.NoError(err)

			ctx := cfg.Context()
			// ensure the right CaCertPath is stored in Config and context name is the right name
			req.Equal(tt.expectedCaCertPath, ctx.Platform.CaCertPath)
			req.Equal(legacyContextName, ctx.Name)
		})
	}
}

func getNewLoginCommandForSelfSignedCertTest(req *require.Assertions, cfg *v1.Config, expectedCaCertPath string) *cobra.Command {
	mdsConfig := mds.NewConfiguration()
	mdsClient := mds.NewAPIClient(mdsConfig)

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
		GetTokenFunc: func(ctx context.Context) (mds.AuthenticationResponse, *http.Response, error) {
			req.NotEqual(http.DefaultClient, mdsClient)
			transport, ok := mdsClient.GetConfig().HTTPClient.Transport.(*http.Transport)
			req.True(ok)
			req.NotEqual(http.DefaultTransport, transport)
			found := false
			for _, actualSubject := range transport.TLSClientConfig.RootCAs.Subjects() { //nolint:staticcheck
				if bytes.Equal(cert.RawSubject, actualSubject) {
					found = true
					break
				}
			}
			req.True(found, "Certificate not found in client.")
			return mds.AuthenticationResponse{
				AuthToken: testToken1,
				TokenType: "JWT",
				ExpiresIn: 100,
			}, nil, nil
		},
	}
	mdsClientManager := &climock.MDSClientManager{
		GetMDSClientFunc: func(_, caCertPath string, _ bool) (*mds.APIClient, error) {
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
		LoginFunc: func(_ context.Context, _ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken1}, nil
		},
		UserFunc: func(_ context.Context) (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: org1Id},
				Accounts:     []*ccloudv1.Account{{Id: "a-595", Name: "Default"}},
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
			args: []string{
				"--url=http://localhost:8090",
			},
		},
	}

	activeApiKey := "bo"
	kafkaCluster := &v1.KafkaClusterConfig{
		ID:        "lkc-0000",
		Name:      "bob",
		Bootstrap: "http://bobby",
		APIKeys: map[string]*v1.APIKeyPair{
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
		verifyLoggedInState(t, cfg, s.isCloud, org1Id)

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
		verifyLoggedInState(t, cfg, s.isCloud, org1Id)

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
			errMsg:   errors.NewErrorWithSuggestions(errors.UnneccessaryUrlFlagForCloudLoginErrorMsg, errors.UnneccessaryUrlFlagForCloudLoginSuggestions).Error(),
		},
		{
			urlIn:    "https://confluent.cloud/login/sso/company",
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

func newLoginCmd(auth *ccloudv1mock.Auth, userInterface *ccloudv1mock.UserInterface, isCloud bool, req *require.Assertions, netrcHandler netrc.NetrcHandler, authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager, loginOrganizationManager pauth.LoginOrganizationManager) (*cobra.Command, *v1.Config) {
	cfg := v1.New()
	var mdsClient *mds.APIClient
	if !isCloud {
		mdsConfig := mds.NewConfiguration()
		mdsClient = mds.NewAPIClient(mdsConfig)
		mdsClient.TokensAndAuthenticationApi = &mdsmock.TokensAndAuthenticationApi{
			GetTokenFunc: func(ctx context.Context) (mds.AuthenticationResponse, *http.Response, error) {
				return mds.AuthenticationResponse{
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
				GetFreeTrialInfoFunc: func(_ context.Context, orgId int32) ([]*ccloudv1.GrowthPromoCodeClaim, error) {
					var claims []*ccloudv1.GrowthPromoCodeClaim
					return claims, nil
				},
			}, Auth: auth, User: userInterface}
		},
	}
	mdsClientManager := &climock.MDSClientManager{
		GetMDSClientFunc: func(_, _ string, _ bool) (*mds.APIClient, error) {
			return mdsClient, nil
		},
	}
	prerunner := climock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, mdsClient, nil, cfg)
	loginCmd := New(cfg, prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, loginOrganizationManager, authTokenHandler)
	loginCmd.Flags().Bool("unsafe-trace", false, "")
	return loginCmd, cfg
}

func newLogoutCmd(cfg *v1.Config, netrcHandler netrc.NetrcHandler) (*cobra.Command, *v1.Config) {
	logoutCmd := logout.New(cfg, climock.NewPreRunnerMock(nil, nil, nil, nil, cfg), netrcHandler)
	return logoutCmd, cfg
}

func verifyLoggedInState(t *testing.T, cfg *v1.Config, isCloud bool, orgResourceId string) {
	req := require.New(t)
	ctx := cfg.Context()
	req.NotNil(ctx)
	if orgResourceId == org1Id || orgResourceId == "" {
		req.Equal(testToken1, ctx.GetAuthToken())
	} else if orgResourceId == org2Id {
		req.Equal(testToken2, ctx.GetAuthToken())
	}
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
		// MDS doesn't set some things like cfg.Auth.User since e.g. an MDS user != an ccloudv1 User
		req.Equal(&ccloudv1.User{Id: 23, Email: promptUser, FirstName: "Cody"}, ctx.GetUser())
		req.Equal(orgResourceId, ctx.GetOrganization().GetResourceId())
		req.Equal(orgResourceId, ctx.Config.GetLastUsedOrgId())
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
