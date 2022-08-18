package cmd_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	sdkMock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/update/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	expiredAuthTokenForDevCloud = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJvcmdhbml6YXRpb25JZCI6MT" +
		"U5NCwidXNlcklkIjoxNTM3MiwiZXhwIjoxNTc0NzIwODgzLCJqdGkiOiJkMzFlYjc2OC0zNzIzLTQ4MTEtYjg3" +
		"Zi1lMTQ2YTQyYmMyMjciLCJpYXQiOjE1NzQ3MTcyODMsImlzcyI6IkNvbmZsdWVudCIsInN1YiI6IjE1MzcyIn" +
		"0.r9o6HEaacidXV899sjYDajCfVd_Tczyfk541jzidw8r0TRGz74RxL2UFK0aGyR4tNrJRSOJlYHSEBNMV7" +
		"J1sEzdGj_mYbvdAL8feH3Sj0uOf1BSKEdhOLsZbQRPn1TnUwUI0ujxjvY3V4l9unXjdRcNceQx1RcAIm8JEo" +
		"Vjpgsb5MRQWYHlTTEwJe5MVY-dZZEsq40YzlchmFi8LVYCxY3rtwEtINbFJx7K-0rW-GJWyek2zRMiUDtmXI" +
		"o8C87TmR9JfLAhLGYKYB-sMnX1FWQs1GSEf9CBGerhZ6T4wwTu_GCVEqg_kDZpGxx1V3nTr0K_lHN8QxFHoJA" +
		"ccbtRHKFuQZaXkJjhsq4i6q9OV-wgL_G7y003Z-hRiBvdBPoEqecXOfI6HKYbzfv9P9N2p0UnfPF2fZWitcmd" +
		"55IgHZ15zwDkFqixoV1hY_tG7dWtQNZIlPDabgm5UH0mc7GS2dh9Z5spZTvqH8xZ0SFF6T5-iFqpJjm6wkzMd6" +
		"1u9UuWTTTNG-Nr_8abS0cYfChZIXde3D1so2KhG4r6uAB1onlNWK4Gq2Lc9uT_r2tKcGDqyZWFPvVtAepr8duW" +
		"ts27QsDs7BvMnwSkUjGv6scSJZWX1fMZbXh7zd0Khg_13dWshAyE935n46T4S7VJm9JhZLEwUcoOPOhWmVcJn5xSJ-YQ"
	validAuthToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiO" +
		"jE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1w" +
		"bGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE"
	jwtWithNoExp = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwia" +
		"WF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
)

var (
	mockLoginCredentialsManager = &cliMock.MockLoginCredentialsManager{
		GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetPrerunCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
		GetCloudCredentialsFromPromptFunc: func(_ *cobra.Command, orgResourceId string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
	}
	mockAuthTokenHandler = &cliMock.MockAuthTokenHandler{
		GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, _ *pauth.Credentials, _ bool, _ string) (string, string, error) {
			return "", "", nil
		},
		GetConfluentTokenFunc: func(_ *mds.APIClient, _ *pauth.Credentials) (string, error) {
			return "", nil
		},
	}
)

func getPreRunBase() *pcmd.PreRun {
	return &pcmd.PreRun{
		Config:  v1.AuthenticatedCloudConfigMock(),
		Version: pmock.NewVersionMock(),
		UpdateClient: &mock.Client{
			CheckForUpdatesFunc: func(_, _ string, _ bool) (string, string, error) {
				return "", "", nil
			},
		},
		FlagResolver: &pcmd.FlagResolverImpl{
			Prompt: &form.RealPrompt{},
			Out:    os.Stdout,
		},
		CCloudClientFactory: &cliMock.MockCCloudClientFactory{
			JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloud.Client {
				return &ccloud.Client{}
			},
			AnonHTTPClientFactoryFunc: func(baseURL string) *ccloud.Client {
				return &ccloud.Client{}
			},
		},
		MDSClientManager: &cliMock.MockMDSClientManager{
			GetMDSClientFunc: func(url, caCertPath string) (client *mds.APIClient, e error) {
				return &mds.APIClient{}, nil
			},
		},
		LoginCredentialsManager: mockLoginCredentialsManager,
		JWTValidator:            pcmd.NewJWTValidator(),
		AuthTokenHandler:        mockAuthTokenHandler,
	}
}

func TestPreRun_Anonymous_SetLoggingLevel(t *testing.T) {
	featureflags.Init(nil, true)

	tests := map[string]log.Level{
		"":      log.ERROR,
		"-v":    log.WARN,
		"-vv":   log.INFO,
		"-vvv":  log.DEBUG,
		"-vvvv": log.TRACE,
	}

	for flags, level := range tests {
		r := getPreRunBase()

		cmd := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().CountP("verbose", "v", "Increase verbosity")
		c := pcmd.NewAnonymousCLICommand(cmd, r)

		_, err := pcmd.ExecuteCommand(c.Command, "help", flags)
		require.NoError(t, err)

		require.Equal(t, level, log.CliLogger.Level)
	}
}

func TestPreRun_HasAPIKey_SetupLoggingAndCheckForUpdates(t *testing.T) {
	calledAnonymous := false

	r := getPreRunBase()
	r.UpdateClient = &mock.Client{
		CheckForUpdatesFunc: func(_, _ string, _ bool) (string, string, error) {
			calledAnonymous = true
			return "", "", nil
		},
	}

	root := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
	root.Flags().CountP("verbose", "v", "Increase verbosity")
	rootCmd := pcmd.NewAnonymousCLICommand(root, r)
	args := strings.Split("help", " ")
	_, err := pcmd.ExecuteCommand(rootCmd.Command, args...)
	require.NoError(t, err)

	if !calledAnonymous {
		t.Errorf("PreRun.HasAPIKey() didn't call the Anonymous() helper to set logging level and updates")
	}
}

func TestPreRun_TokenExpires(t *testing.T) {
	cfg := v1.AuthenticatedCloudConfigMock()
	cfg.Context().State.AuthToken = expiredAuthTokenForDevCloud

	r := getPreRunBase()
	r.Config = cfg

	root := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {},
	}
	rootCmd := pcmd.NewAnonymousCLICommand(root, r)
	root.Flags().CountP("verbose", "v", "Increase verbosity")

	_, err := pcmd.ExecuteCommand(rootCmd.Command)
	require.NoError(t, err)

	// Check auth is nil for now, until there is a better to create a fake logged in user and check if it's logged out
	require.Nil(t, cfg.Context().State.Auth)
}

func Test_UpdateToken(t *testing.T) {
	tests := []struct {
		name      string
		isCloud   bool
		authToken string
	}{
		{
			name:      "ccloud expired token",
			isCloud:   true,
			authToken: expiredAuthTokenForDevCloud,
		},
		{
			name:      "ccloud empty token",
			isCloud:   true,
			authToken: "",
		},
		{
			name:      "ccloud invalid token",
			isCloud:   true,
			authToken: "jajajajaja",
		},
		{
			name:      "ccloud jwt with no exp claim",
			isCloud:   true,
			authToken: jwtWithNoExp,
		},
		{
			name:      "confluent expired token",
			authToken: expiredAuthTokenForDevCloud,
		},
		{
			name:      "confluent empty token",
			authToken: "",
		},
		{
			name:      "confluent invalid token",
			authToken: "jajajajaja",
		},
		{
			name:      "confluent jwt with no exp claim",
			authToken: jwtWithNoExp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *v1.Config
			if tt.isCloud {
				cfg = v1.AuthenticatedCloudConfigMock()
			} else {
				cfg = v1.AuthenticatedOnPremConfigMock()
			}

			cfg.Context().State.AuthToken = tt.authToken

			mockLoginCredentialsManager := &cliMock.MockLoginCredentialsManager{
				GetPrerunCredentialsFromConfigFunc: func(cfg *v1.Config) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetCredentialsFromNetrcFunc: func(cmd *cobra.Command, filterParams netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return &pauth.Credentials{Username: "username", Password: "password"}, nil
					}
				},
			}

			r := getPreRunBase()
			r.Config = cfg
			r.LoginCredentialsManager = mockLoginCredentialsManager

			root := &cobra.Command{
				Run: func(cmd *cobra.Command, args []string) {},
			}
			rootCmd := pcmd.NewAnonymousCLICommand(root, r)
			root.Flags().CountP("verbose", "v", "Increase verbosity")

			_, err := pcmd.ExecuteCommand(rootCmd.Command)
			require.NoError(t, err)
			require.True(t, mockLoginCredentialsManager.GetCredentialsFromNetrcCalled())
		})
	}
}

func TestPrerun_AutoLogin(t *testing.T) {
	type credentialsFuncReturnValues struct {
		creds *pauth.Credentials
		err   error
	}

	username := "csreesangkom"

	ccloudCreds := &pauth.Credentials{
		Username: username,
		Password: "csreepassword",
	}
	confluentCreds := &pauth.Credentials{
		Username:       username,
		Password:       "csreepassword",
		PrerunLoginURL: "http://localhost:8090",
	}
	tests := []struct {
		name          string
		isCloud       bool
		envVarChecked bool
		netrcChecked  bool
		wantErr       bool
		envVarReturn  credentialsFuncReturnValues
		netrcReturn   credentialsFuncReturnValues
	}{
		{
			name:          "CCloud no env var credentials but successful login from netrc",
			isCloud:       true,
			envVarReturn:  credentialsFuncReturnValues{nil, nil},
			netrcReturn:   credentialsFuncReturnValues{ccloudCreds, nil},
			envVarChecked: true,
			netrcChecked:  true,
		},
		{
			name:          "Confluent no env var credentials but successful login from netrc",
			envVarReturn:  credentialsFuncReturnValues{nil, nil},
			netrcReturn:   credentialsFuncReturnValues{confluentCreds, nil},
			envVarChecked: true,
			netrcChecked:  true,
		},
		{
			name:          "CCloud successful login from env var",
			isCloud:       true,
			envVarReturn:  credentialsFuncReturnValues{ccloudCreds, nil},
			netrcReturn:   credentialsFuncReturnValues{ccloudCreds, nil},
			envVarChecked: true,
			netrcChecked:  false,
		},
		{
			name:          "Confluent successful login from env var",
			envVarReturn:  credentialsFuncReturnValues{confluentCreds, nil},
			netrcReturn:   credentialsFuncReturnValues{confluentCreds, nil},
			envVarChecked: true,
			netrcChecked:  false,
		},
		{
			name:          "CCloud env var failed but netrc succeeds",
			isCloud:       true,
			envVarReturn:  credentialsFuncReturnValues{nil, errors.New("ENV VAR FAILED")},
			netrcReturn:   credentialsFuncReturnValues{ccloudCreds, nil},
			envVarChecked: true,
			netrcChecked:  true,
		},
		{
			name:          "Confluent env var failed but netrc succeeds",
			envVarReturn:  credentialsFuncReturnValues{nil, errors.New("ENV VAR FAILED")},
			netrcReturn:   credentialsFuncReturnValues{confluentCreds, nil},
			envVarChecked: true,
			netrcChecked:  true,
		},
		{
			name:          "CCloud failed non-interactive login",
			isCloud:       true,
			envVarReturn:  credentialsFuncReturnValues{nil, errors.New("ENV VAR FAILED")},
			netrcReturn:   credentialsFuncReturnValues{nil, errors.New("NETRC FAILED")},
			envVarChecked: true,
			netrcChecked:  true,
			wantErr:       true,
		},
		{
			name:          "Confluent failed non-interactive login",
			envVarReturn:  credentialsFuncReturnValues{nil, errors.New("ENV VAR FAILED")},
			netrcReturn:   credentialsFuncReturnValues{nil, errors.New("NETRC FAILED")},
			envVarChecked: true,
			netrcChecked:  true,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *v1.Config
			if tt.isCloud {
				cfg = v1.AuthenticatedCloudConfigMock()
			} else {
				cfg = v1.AuthenticatedOnPremConfigMock()
			}
			err := pauth.PersistLogoutToConfig(cfg)
			require.NoError(t, err)

			r := getPreRunBase()
			r.Config = cfg
			r.CCloudClientFactory = &cliMock.MockCCloudClientFactory{
				JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloud.Client {
					return &ccloud.Client{Auth: &sdkMock.Auth{
						UserFunc: func(_ context.Context) (*flowv1.GetMeReply, error) {
							return &flowv1.GetMeReply{
								User: &orgv1.User{
									Id:        23,
									Email:     "",
									FirstName: "",
								},
								Organization: &orgv1.Organization{ResourceId: "o-123"},
								Accounts:     []*orgv1.Account{{Id: "a-595", Name: "Default"}},
							}, nil
						},
					}}
				},
				AnonHTTPClientFactoryFunc: func(baseURL string) *ccloud.Client {
					return &ccloud.Client{}
				},
			}
			r.AuthTokenHandler = &cliMock.MockAuthTokenHandler{
				GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, _ *pauth.Credentials, _ bool, _ string) (string, string, error) {
					return validAuthToken, "", nil
				},
				GetConfluentTokenFunc: func(mdsClient *mds.APIClient, credentials *pauth.Credentials) (s string, e error) {
					return validAuthToken, nil
				},
			}

			var ccloudEnvVarCalled bool
			var ccloudNetrcCalled bool
			var confluentEnvVarCalled bool
			var confluentNetrcCalled bool
			r.LoginCredentialsManager = &cliMock.MockLoginCredentialsManager{
				GetCloudCredentialsFromEnvVarFunc: func(orgResourceId string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						ccloudEnvVarCalled = true
						return tt.envVarReturn.creds, tt.envVarReturn.err
					}
				},
				GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						ccloudNetrcCalled = true
						return tt.netrcReturn.creds, tt.netrcReturn.err
					}
				},
				GetPrerunCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						return nil, nil
					}
				},
				GetOnPremPrerunCredentialsFromEnvVarFunc: func() func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						confluentEnvVarCalled = true
						return tt.envVarReturn.creds, tt.envVarReturn.err
					}
				},
				GetOnPremPrerunCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (credentials *pauth.Credentials, e error) {
						confluentNetrcCalled = true
						return tt.netrcReturn.creds, tt.netrcReturn.err
					}
				},
			}

			root := &cobra.Command{
				Run: func(cmd *cobra.Command, args []string) {},
			}
			var rootCmd *pcmd.AuthenticatedCLICommand
			if tt.isCloud {
				rootCmd = pcmd.NewAuthenticatedCLICommand(root, r)
			} else {
				rootCmd = pcmd.NewAuthenticatedWithMDSCLICommand(root, r)
			}
			root.Flags().CountP("verbose", "v", "Increase verbosity")

			out, err := pcmd.ExecuteCommand(rootCmd.Command)

			if tt.isCloud {
				require.Equal(t, tt.envVarChecked, ccloudEnvVarCalled)
				require.Equal(t, tt.netrcChecked, ccloudNetrcCalled)
				require.False(t, confluentEnvVarCalled)
				require.False(t, confluentNetrcCalled)
			} else {
				require.Equal(t, tt.envVarChecked, confluentEnvVarCalled)
				require.Equal(t, tt.netrcChecked, confluentNetrcCalled)
				require.False(t, ccloudEnvVarCalled)
				require.False(t, ccloudNetrcCalled)
			}

			if !tt.wantErr {
				require.NoError(t, err)
				require.NotContains(t, out, errors.AutoLoginMsg)
				require.NotContains(t, out, fmt.Sprintf(errors.LoggedInAsMsg, username))
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), errors.NotLoggedInErrorMsg)
			}
		})
	}
}

func TestPrerun_ReLoginToLastOrgUsed(t *testing.T) {
	ccloudCreds := &pauth.Credentials{
		Username: "username",
		Password: "password",
	}
	r := getPreRunBase()
	r.CCloudClientFactory = &cliMock.MockCCloudClientFactory{
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloud.Client {
			return &ccloud.Client{Auth: &sdkMock.Auth{
				UserFunc: func(ctx context.Context) (*flowv1.GetMeReply, error) {
					return &flowv1.GetMeReply{
						User: &orgv1.User{
							Id:        23,
							Email:     "",
							FirstName: "",
						},
						Organization: &orgv1.Organization{ResourceId: "o-123"},
						Accounts:     []*orgv1.Account{{Id: "a-595", Name: "Default"}},
					}, nil
				},
			}}
		},
		AnonHTTPClientFactoryFunc: func(baseURL string) *ccloud.Client {
			return &ccloud.Client{}
		},
	}
	r.AuthTokenHandler = &cliMock.MockAuthTokenHandler{
		GetCCloudTokensFunc: func(_ pauth.CCloudClientFactory, _ string, _ *pauth.Credentials, _ bool, orgResourceId string) (s string, s2 string, e error) {
			require.Equal(t, "o-555", orgResourceId) // validate correct org id is used
			return validAuthToken, "", nil
		},
	}
	r.LoginCredentialsManager = &cliMock.MockLoginCredentialsManager{
		GetCredentialsFromNetrcFunc: mockLoginCredentialsManager.GetCredentialsFromNetrcFunc,
		GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return ccloudCreds, nil
			}
		},
		GetPrerunCredentialsFromConfigFunc: func(_ *v1.Config) func() (*pauth.Credentials, error) {
			return func() (*pauth.Credentials, error) {
				return nil, nil
			}
		},
	}

	cfg := v1.AuthenticatedToOrgCloudConfigMock(555, "o-555")
	err := cfg.Context().DeleteUserAuth()
	require.NoError(t, err)
	r.Config = cfg

	root := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {},
	}
	rootCmd := pcmd.NewAuthenticatedCLICommand(root, r)
	root.Flags().CountP("verbose", "v", "Increase verbosity")

	_, err = pcmd.ExecuteCommand(rootCmd.Command)
	require.NoError(t, err)
}

func TestPrerun_AutoLoginNotTriggeredIfLoggedIn(t *testing.T) {
	tests := []struct {
		name    string
		isCloud bool
	}{
		{
			name:    "ccloud logged in user",
			isCloud: true,
		},
		{
			name: "confluent logged in user",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *v1.Config
			if tt.isCloud {
				cfg = v1.AuthenticatedCloudConfigMock()
			} else {
				cfg = v1.AuthenticatedOnPremConfigMock()
			}
			cfg.Context().State.AuthToken = validAuthToken
			cfg.Context().Platform.Server = "https://confluent.cloud"

			var envVarCalled bool
			var netrcCalled bool
			mockLoginCredentialsManager := &cliMock.MockLoginCredentialsManager{
				GetCloudCredentialsFromEnvVarFunc: func(_ string) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						envVarCalled = true
						return nil, nil
					}
				},
				GetCredentialsFromNetrcFunc: func(_ *cobra.Command, filterParams netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
					return func() (*pauth.Credentials, error) {
						netrcCalled = true
						return nil, nil
					}
				},
			}

			r := getPreRunBase()
			r.Config = cfg
			r.LoginCredentialsManager = mockLoginCredentialsManager

			root := &cobra.Command{
				Run: func(cmd *cobra.Command, args []string) {},
			}
			var rootCmd *pcmd.AuthenticatedCLICommand
			if tt.isCloud {
				rootCmd = pcmd.NewAuthenticatedCLICommand(root, r)
			} else {
				rootCmd = pcmd.NewAuthenticatedWithMDSCLICommand(root, r)
			}

			root.Flags().CountP("verbose", "v", "Increase verbosity")

			_, err := pcmd.ExecuteCommand(rootCmd.Command)
			require.NoError(t, err)
			require.False(t, netrcCalled)
			require.False(t, envVarCalled)
		})
	}
}

func TestPreRun_HasAPIKeyCommand(t *testing.T) {
	userNameConfigLoggedIn := v1.AuthenticatedCloudConfigMock()
	userNameConfigLoggedIn.Context().State.AuthToken = validAuthToken

	userNameCfgCorruptedAuthToken := v1.AuthenticatedCloudConfigMock()
	userNameCfgCorruptedAuthToken.Context().State.AuthToken = "corrupted.auth.token"

	userNotLoggedIn := v1.UnauthenticatedCloudConfigMock()

	usernameClusterWithoutKeyOrSecret := v1.AuthenticatedCloudConfigMock()
	usernameClusterWithoutKeyOrSecret.Context().State.AuthToken = validAuthToken
	usernameClusterWithoutKeyOrSecret.Context().KafkaClusterContext.GetKafkaClusterConfig(v1.MockKafkaClusterId()).APIKey = ""

	usernameClusterWithStoredSecret := v1.AuthenticatedCloudConfigMock()
	usernameClusterWithStoredSecret.Context().State.AuthToken = validAuthToken
	usernameClusterWithStoredSecret.Context().KafkaClusterContext.GetKafkaClusterConfig(v1.MockKafkaClusterId()).APIKeys["miles"] = &v1.APIKeyPair{
		Key:    "miles",
		Secret: "secret",
	}
	usernameClusterWithoutSecret := v1.AuthenticatedCloudConfigMock()
	usernameClusterWithoutSecret.Context().State.AuthToken = validAuthToken
	tests := []struct {
		name           string
		config         *v1.Config
		errMsg         string
		suggestionsMsg string
		key            string
		secret         string
	}{
		{
			name:   "username logged in user",
			config: userNameConfigLoggedIn,
		},
		{
			name:   "not logged in user",
			config: userNotLoggedIn,
			errMsg: errors.NotLoggedInErrorMsg,
		},
		{
			name:   "api credential context",
			config: v1.APICredentialConfigMock(),
		},
		{
			name:   "api key and secret passed via flags",
			key:    "miles",
			secret: "shhhh",
			config: usernameClusterWithoutKeyOrSecret,
		},
		{
			name:   "api key passed via flag with stored secret",
			key:    "miles",
			config: usernameClusterWithStoredSecret,
		},
		{
			name:           "api key passed via flag without stored secret",
			key:            "miles",
			errMsg:         fmt.Sprintf(errors.NoAPISecretStoredOrPassedErrorMsg, "miles", v1.MockKafkaClusterId()),
			suggestionsMsg: fmt.Sprintf(errors.NoAPISecretStoredOrPassedSuggestions, "miles", v1.MockKafkaClusterId()),
			config:         usernameClusterWithoutSecret,
		},
		{
			name:           "just api secret passed via flag",
			secret:         "shhhh",
			config:         usernameClusterWithoutKeyOrSecret,
			errMsg:         errors.PassedSecretButNotKeyErrorMsg,
			suggestionsMsg: errors.PassedSecretButNotKeySuggestions,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := getPreRunBase()
			r.Config = tt.config

			root := &cobra.Command{
				Run: func(cmd *cobra.Command, args []string) {},
			}
			rootCmd := pcmd.NewHasAPIKeyCLICommand(root, r)
			root.Flags().CountP("verbose", "v", "Increase verbosity")
			root.Flags().String("api-key", "", "Kafka cluster API key.")
			root.Flags().String("api-secret", "", "API key secret.")
			root.Flags().String("cluster", "", "Kafka cluster ID.")

			_, err := pcmd.ExecuteCommand(rootCmd.Command, "--api-key", tt.key, "--api-secret", tt.secret)
			if tt.errMsg != "" {
				require.Error(t, err)
				require.Equal(t, tt.errMsg, err.Error())
				if tt.suggestionsMsg != "" {
					errors.VerifyErrorAndSuggestions(require.New(t), err, tt.errMsg, tt.suggestionsMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInitializeOnPremKafkaRest(t *testing.T) {
	cfg := v1.AuthenticatedOnPremConfigMock()
	cfg.Context().State.AuthToken = validAuthToken
	r := getPreRunBase()
	r.Config = cfg
	cobraCmd := &cobra.Command{Use: "test"}
	cobraCmd.Flags().CountP("verbose", "v", "Increase verbosity")
	cmd := pcmd.NewAuthenticatedCLICommand(cobraCmd, r)
	t.Run("InitializeOnPremKafkaRest_ValidMdsToken", func(t *testing.T) {
		err := r.InitializeOnPremKafkaRest(cmd)(cmd.Command, []string{})
		require.NoError(t, err)
		kafkaRest, err := cmd.GetKafkaREST()
		require.NoError(t, err)
		auth, ok := kafkaRest.Context.Value(krsdk.ContextAccessToken).(string)
		require.True(t, ok)
		require.Equal(t, validAuthToken, auth)
	})
	r.Config.Context().State.AuthToken = ""
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	t.Run("InitializeOnPremKafkaRest_InvalidMdsToken", func(t *testing.T) {
		mockLoginCredentialsManager := &cliMock.MockLoginCredentialsManager{
			GetOnPremPrerunCredentialsFromEnvVarFunc: func() func() (*pauth.Credentials, error) {
				return func() (*pauth.Credentials, error) {
					return nil, nil
				}
			},
			GetOnPremPrerunCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
				return func() (*pauth.Credentials, error) {
					return nil, nil
				}
			},
			GetPrerunCredentialsFromConfigFunc: func(cfg *v1.Config) func() (*pauth.Credentials, error) {
				return func() (*pauth.Credentials, error) {
					return nil, nil
				}
			},
			GetCredentialsFromNetrcFunc: func(_ *cobra.Command, _ netrc.NetrcMachineParams) func() (*pauth.Credentials, error) {
				return func() (*pauth.Credentials, error) {
					return nil, nil
				}
			},
		}
		r.LoginCredentialsManager = mockLoginCredentialsManager
		err := r.InitializeOnPremKafkaRest(cmd)(cmd.Command, []string{})
		require.NoError(t, err)
		kafkaRest, err := cmd.GetKafkaREST()
		require.Error(t, err)
		require.Nil(t, kafkaRest)
		require.Contains(t, buf.String(), errors.MDSTokenNotFoundMsg)
	})
}

func TestConvertToMetricsBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		inputUrl    string
		expectedUrl string
	}{
		{
			"test exact url",
			"https://api.telemetry.confluent.cloud/",
			"https://api.telemetry.confluent.cloud/",
		},
		{
			"test dev url",
			"https://devel.cpdev.cloud",
			"https://devel-sandbox-api.telemetry.aws.confluent.cloud/",
		},
		{
			"test cpd url",
			"https://nearby-asp.gcp.priv.cpdev.cloud",
			"https://devel-sandbox-api.telemetry.aws.confluent.cloud/",
		},
		{
			"test stag url",
			"https://stag.cpdev.cloud",
			"https://stag-sandbox-api.telemetry.aws.confluent.cloud/",
		},
		{
			"test prod url",
			"https://confluent.cloud",
			"https://api.telemetry.confluent.cloud/",
		},
		{
			"test prod url",
			"https://confluent.cloud/",
			"https://api.telemetry.confluent.cloud/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pcmd.ConvertToMetricsBaseURL(tt.inputUrl)
			if got != tt.expectedUrl {
				t.Errorf("got = %v, want %v", got, tt.expectedUrl)
			}
		})
	}
}
