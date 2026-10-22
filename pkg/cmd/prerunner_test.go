package cmd_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	climock "github.com/confluentinc/cli/v4/mock"
	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/log"
	pmock "github.com/confluentinc/cli/v4/pkg/mock"
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

func getPreRunBase(ctrl *gomock.Controller) *pcmd.PreRun {
	loginCredentialsManager := climock.NewMockLoginCredentialsManager(ctrl)
	loginCredentialsManager.EXPECT().GetCloudCredentialsFromEnvVar(gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return nil, nil
	}).AnyTimes()
	loginCredentialsManager.EXPECT().GetPrerunCredentialsFromConfig(gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return nil, nil
	}).AnyTimes()
	loginCredentialsManager.EXPECT().GetCloudCredentialsFromPrompt(gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return nil, nil
	}).AnyTimes()
	loginCredentialsManager.EXPECT().GetCredentialsFromKeychain(gomock.Any(), gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return nil, nil
	}).AnyTimes()
	loginCredentialsManager.EXPECT().GetCredentialsFromConfig(gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return nil, nil
	}).AnyTimes()

	authTokenHandler := climock.NewMockAuthTokenHandler(ctrl)
	authTokenHandler.EXPECT().GetCCloudTokens(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", "", nil).AnyTimes()
	authTokenHandler.EXPECT().GetConfluentToken(gomock.Any(), gomock.Any(), gomock.Any()).Return("", "", nil).AnyTimes()

	ccloudClientFactory := climock.NewMockCCloudClientFactory(ctrl)
	ccloudClientFactory.EXPECT().JwtHTTPClientFactory(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ccloudv1.Client{}).AnyTimes()
	ccloudClientFactory.EXPECT().AnonHTTPClientFactory(gomock.Any()).Return(&ccloudv1.Client{}).AnyTimes()

	mdsClientManager := climock.NewMockMDSClientManager(ctrl)
	mdsClientManager.EXPECT().GetMDSClient(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mdsv1.APIClient{}, nil).AnyTimes()

	return &pcmd.PreRun{
		Config:                  config.AuthenticatedCloudConfigMock(),
		Version:                 pmock.NewVersionMock(),
		CCloudClientFactory:     ccloudClientFactory,
		MDSClientManager:        mdsClientManager,
		LoginCredentialsManager: loginCredentialsManager,
		JWTValidator:            jwt.NewValidator(),
		AuthTokenHandler:        authTokenHandler,
	}
}

func TestPreRunAnonymousSetLoggingLevel(t *testing.T) {
	cfg := &config.Config{IsTest: true, Contexts: map[string]*config.Context{}}
	featureflags.Init(cfg)

	tests := map[string]log.Level{
		"":      log.ERROR,
		"-v":    log.WARN,
		"-vv":   log.INFO,
		"-vvv":  log.DEBUG,
		"-vvvv": log.TRACE,
	}

	for flags, level := range tests {
		ctrl := gomock.NewController(t)
		r := getPreRunBase(ctrl)

		cmd := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().CountP("verbose", "v", "Increase verbosity")
		cmd.Flags().Bool("unsafe-trace", false, "")
		c := pcmd.NewAnonymousCLICommand(cmd, r)

		_, err := pcmd.ExecuteCommand(c.Command, "help", flags)
		require.NoError(t, err)

		require.Equal(t, level, log.CliLogger.Level)
	}
}

func TestPreRunTokenExpires(t *testing.T) {
	cfg := config.AuthenticatedCloudConfigMock()
	cfg.Context().State.AuthToken = expiredAuthTokenForDevCloud

	ctrl := gomock.NewController(t)
	r := getPreRunBase(ctrl)
	r.Config = cfg

	root := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
	rootCmd := pcmd.NewAuthenticatedCLICommand(root, r)
	root.Flags().CountP("verbose", "v", "Increase verbosity")
	root.Flags().Bool("unsafe-trace", false, "")

	_, err := pcmd.ExecuteCommand(rootCmd.Command)
	require.Error(t, err)

	// Check auth is nil for now, until there is a better to create a fake logged in user and check if it's logged out
	require.Nil(t, cfg.Context().State.Auth)
}

func TestUpdateToken(t *testing.T) {
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var cfg *config.Config
			if test.isCloud {
				cfg = config.AuthenticatedCloudConfigMock()
			} else {
				cfg = config.AuthenticatedOnPremConfigMock()
			}

			cfg.Context().State.AuthToken = test.authToken

			ctrl := gomock.NewController(t)
			r := getPreRunBase(ctrl)
			r.Config = cfg

			loginCredentialsManager := climock.NewMockLoginCredentialsManager(ctrl)
			loginCredentialsManager.EXPECT().GetPrerunCredentialsFromConfig(gomock.Any()).Return(func() (*pauth.Credentials, error) {
				return nil, nil
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetCredentialsFromKeychain(gomock.Any(), gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
				return nil, nil
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetCredentialsFromConfig(gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
				return &pauth.Credentials{Username: "username", Password: "password"}, nil
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetCloudCredentialsFromEnvVar(gomock.Any()).Return(func() (*pauth.Credentials, error) {
				return nil, nil
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetOnPremSsoCredentialsFromConfig(gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
				return nil, nil
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetOnPremPrerunCredentialsFromEnvVar().Return(func() (*pauth.Credentials, error) {
				return nil, nil
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetOnPremCredentialsFromEnvVar().Return(func() (*pauth.Credentials, error) {
				return nil, nil
			}).AnyTimes()
			r.LoginCredentialsManager = loginCredentialsManager

			root := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
			var rootCmd *pcmd.AuthenticatedCLICommand
			if test.isCloud {
				rootCmd = pcmd.NewAuthenticatedCLICommand(root, r)
			} else {
				rootCmd = pcmd.NewAuthenticatedWithMDSCLICommand(root, r)
			}
			root.Flags().CountP("verbose", "v", "Increase verbosity")
			root.Flags().Bool("unsafe-trace", false, "")

			_, err := pcmd.ExecuteCommand(rootCmd.Command, "--unsafe-trace")
			require.Error(t, err)
		})
	}
}

func TestPrerunAutoLogin(t *testing.T) {
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
		name            string
		isCloud         bool
		envVarChecked   bool
		keychainChecked bool
		configChecked   bool
		wantErr         bool
		envVarReturn    credentialsFuncReturnValues
		keychainReturn  credentialsFuncReturnValues
		configReturn    credentialsFuncReturnValues
	}{
		{
			name:            "CCloud no env var credentials but successful login from keychain",
			isCloud:         true,
			envVarReturn:    credentialsFuncReturnValues{nil, nil},
			keychainReturn:  credentialsFuncReturnValues{ccloudCreds, nil},
			configReturn:    credentialsFuncReturnValues{ccloudCreds, nil},
			envVarChecked:   true,
			keychainChecked: true,
		},
		{
			name:            "CCloud no env var credentials no keychain but successful login from config",
			isCloud:         true,
			envVarReturn:    credentialsFuncReturnValues{nil, nil},
			keychainReturn:  credentialsFuncReturnValues{nil, nil},
			configReturn:    credentialsFuncReturnValues{ccloudCreds, nil},
			envVarChecked:   true,
			keychainChecked: true,
			configChecked:   true,
		},
		{
			name:          "CCloud successful login from env var",
			isCloud:       true,
			envVarReturn:  credentialsFuncReturnValues{ccloudCreds, nil},
			configReturn:  credentialsFuncReturnValues{ccloudCreds, nil},
			envVarChecked: true,
			configChecked: false,
		},
		{
			name:          "Confluent successful login from env var",
			envVarReturn:  credentialsFuncReturnValues{confluentCreds, nil},
			configReturn:  credentialsFuncReturnValues{confluentCreds, nil},
			envVarChecked: true,
			configChecked: false,
		},
		{
			name:            "CCloud env var failed but config succeeds",
			isCloud:         true,
			envVarReturn:    credentialsFuncReturnValues{nil, fmt.Errorf("ENV VAR FAILED")},
			configReturn:    credentialsFuncReturnValues{ccloudCreds, nil},
			envVarChecked:   true,
			keychainChecked: true,
			configChecked:   true,
		},
		{
			name:            "CCloud failed non-interactive login",
			isCloud:         true,
			envVarReturn:    credentialsFuncReturnValues{nil, fmt.Errorf("ENV VAR FAILED")},
			envVarChecked:   true,
			keychainChecked: true,
			configChecked:   true,
			wantErr:         true,
		},
		{
			name:          "Confluent failed non-interactive login",
			envVarReturn:  credentialsFuncReturnValues{nil, fmt.Errorf("ENV VAR FAILED")},
			configReturn:  credentialsFuncReturnValues{nil, fmt.Errorf("CONFIG FAILED")},
			envVarChecked: true,
			wantErr:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var cfg *config.Config
			if test.isCloud {
				cfg = config.AuthenticatedCloudConfigMock()
			} else {
				cfg = config.AuthenticatedOnPremConfigMock()
			}
			err := pauth.PersistLogout(cfg)
			require.NoError(t, err)

			ctrl := gomock.NewController(t)
			r := getPreRunBase(ctrl)
			r.Config = cfg

			ccloudClientFactory := climock.NewMockCCloudClientFactory(ctrl)
			ccloudClientFactory.EXPECT().JwtHTTPClientFactory(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ccloudv1.Client{Auth: &ccloudv1mock.Auth{
				UserFunc: func() (*ccloudv1.GetMeReply, error) {
					return &ccloudv1.GetMeReply{
						User:         &ccloudv1.User{Id: 23},
						Organization: &ccloudv1.Organization{ResourceId: "o-123"},
						Accounts:     []*ccloudv1.Account{{Id: "env-596", Name: "Default"}},
					}, nil
				},
			}}).AnyTimes()
			ccloudClientFactory.EXPECT().AnonHTTPClientFactory(gomock.Any()).Return(&ccloudv1.Client{}).AnyTimes()
			r.CCloudClientFactory = ccloudClientFactory

			authTokenHandler := climock.NewMockAuthTokenHandler(ctrl)
			authTokenHandler.EXPECT().GetCCloudTokens(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(validAuthToken, "", nil).AnyTimes()
			authTokenHandler.EXPECT().GetConfluentToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(validAuthToken, "", nil).AnyTimes()
			r.AuthTokenHandler = authTokenHandler

			var ccloudEnvVarCalled bool
			var ccloudConfigCalled bool
			var ccloudKeychainCalled bool
			var confluentEnvVarCalled bool
			loginCredentialsManager := climock.NewMockLoginCredentialsManager(ctrl)
			loginCredentialsManager.EXPECT().GetCloudCredentialsFromEnvVar(gomock.Any()).Return(func() (*pauth.Credentials, error) {
				ccloudEnvVarCalled = true
				return test.envVarReturn.creds, test.envVarReturn.err
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetPrerunCredentialsFromConfig(gomock.Any()).Return(func() (*pauth.Credentials, error) {
				ccloudConfigCalled = true
				return test.configReturn.creds, test.configReturn.err
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetOnPremPrerunCredentialsFromEnvVar().Return(func() (*pauth.Credentials, error) {
				confluentEnvVarCalled = true
				return test.envVarReturn.creds, test.envVarReturn.err
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetCredentialsFromConfig(gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
				return nil, nil
			}).AnyTimes()
			loginCredentialsManager.EXPECT().GetCredentialsFromKeychain(gomock.Any(), gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
				ccloudKeychainCalled = true
				return test.keychainReturn.creds, test.keychainReturn.err
			}).AnyTimes()
			r.LoginCredentialsManager = loginCredentialsManager

			root := &cobra.Command{
				Run: func(cmd *cobra.Command, args []string) {},
			}
			var rootCmd *pcmd.AuthenticatedCLICommand
			if test.isCloud {
				rootCmd = pcmd.NewAuthenticatedCLICommand(root, r)
			} else {
				rootCmd = pcmd.NewAuthenticatedWithMDSCLICommand(root, r)
			}
			root.Flags().CountP("verbose", "v", "Increase verbosity")
			root.Flags().Bool("unsafe-trace", false, "")

			out, err := pcmd.ExecuteCommand(rootCmd.Command)

			if test.isCloud {
				require.Equal(t, test.envVarChecked, ccloudEnvVarCalled)
				require.Equal(t, test.configChecked, ccloudConfigCalled)
				require.Equal(t, test.keychainChecked, ccloudKeychainCalled)
				require.False(t, confluentEnvVarCalled)
			} else {
				require.Equal(t, test.envVarChecked, confluentEnvVarCalled)
				require.Equal(t, test.keychainChecked, ccloudKeychainCalled)
				require.False(t, ccloudEnvVarCalled)
			}

			if !test.wantErr {
				require.NoError(t, err)
				require.NotContains(t, out, "Successful auto log in with non-interactive credentials.\n")
				require.NotContains(t, out, fmt.Sprintf(errors.LoggedInAsMsg, username))
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), errors.NotLoggedInErrorMsg)
			}
		})
	}
}

func TestPrerunReLoginToLastOrgUsed(t *testing.T) {
	ccloudCreds := &pauth.Credentials{
		Username: "username",
		Password: "password",
	}
	ctrl := gomock.NewController(t)
	r := getPreRunBase(ctrl)

	ccloudClientFactory := climock.NewMockCCloudClientFactory(ctrl)
	ccloudClientFactory.EXPECT().JwtHTTPClientFactory(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ccloudv1.Client{Auth: &ccloudv1mock.Auth{
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User:         &ccloudv1.User{Id: 23},
				Organization: &ccloudv1.Organization{ResourceId: "o-123"},
				Accounts:     []*ccloudv1.Account{{Id: "env-596", Name: "Default"}},
			}, nil
		},
	}}).AnyTimes()
	ccloudClientFactory.EXPECT().AnonHTTPClientFactory(gomock.Any()).Return(&ccloudv1.Client{}).AnyTimes()
	r.CCloudClientFactory = ccloudClientFactory

	authTokenHandler := climock.NewMockAuthTokenHandler(ctrl)
	authTokenHandler.EXPECT().GetCCloudTokens(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ pauth.CCloudClientFactory, _ string, _ *pauth.Credentials, _ bool, organizationId string) (string, string, error) {
		require.Equal(t, "o-555", organizationId)
		return validAuthToken, "", nil
	}).AnyTimes()
	r.AuthTokenHandler = authTokenHandler

	loginCredentialsManager := climock.NewMockLoginCredentialsManager(ctrl)
	loginCredentialsManager.EXPECT().GetCloudCredentialsFromEnvVar(gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return ccloudCreds, nil
	}).AnyTimes()
	loginCredentialsManager.EXPECT().GetPrerunCredentialsFromConfig(gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return nil, nil
	}).AnyTimes()
	loginCredentialsManager.EXPECT().GetCredentialsFromConfig(gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return nil, nil
	}).AnyTimes()
	loginCredentialsManager.EXPECT().GetCredentialsFromKeychain(gomock.Any(), gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
		return nil, nil
	}).AnyTimes()
	r.LoginCredentialsManager = loginCredentialsManager

	cfg := config.AuthenticatedToOrgCloudConfigMock(555, "o-555")
	cfg.Context().Platform = &config.Platform{Name: "confluent.cloud", Server: "https://confluent.cloud"}
	err := cfg.Context().DeleteUserAuth()
	require.NoError(t, err)
	r.Config = cfg

	root := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
	rootCmd := pcmd.NewAuthenticatedCLICommand(root, r)
	root.Flags().CountP("verbose", "v", "Increase verbosity")
	root.Flags().Bool("unsafe-trace", false, "")

	_, err = pcmd.ExecuteCommand(rootCmd.Command)
	require.NoError(t, err)
}

func TestPrerunAutoLoginNotTriggeredIfLoggedIn(t *testing.T) {
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var cfg *config.Config
			if test.isCloud {
				cfg = config.AuthenticatedCloudConfigMock()
			} else {
				cfg = config.AuthenticatedOnPremConfigMock()
			}
			cfg.Context().State.AuthToken = validAuthToken
			cfg.Context().Platform.Server = "https://confluent.cloud"

			var envVarCalled bool
			ctrl := gomock.NewController(t)
			mockLoginCredentialsManager := climock.NewMockLoginCredentialsManager(ctrl)
			mockLoginCredentialsManager.EXPECT().GetCloudCredentialsFromEnvVar(gomock.Any()).Return(func() (*pauth.Credentials, error) {
				envVarCalled = true
				return nil, nil
			}).AnyTimes()
			mockLoginCredentialsManager.EXPECT().GetCredentialsFromKeychain(gomock.Any(), gomock.Any(), gomock.Any()).Return(func() (*pauth.Credentials, error) {
				return nil, nil
			}).AnyTimes()

			r := getPreRunBase(ctrl)
			r.Config = cfg
			r.LoginCredentialsManager = mockLoginCredentialsManager

			root := &cobra.Command{
				Run: func(cmd *cobra.Command, args []string) {},
			}
			var rootCmd *pcmd.AuthenticatedCLICommand
			if test.isCloud {
				rootCmd = pcmd.NewAuthenticatedCLICommand(root, r)
			} else {
				rootCmd = pcmd.NewAuthenticatedWithMDSCLICommand(root, r)
			}

			root.Flags().CountP("verbose", "v", "Increase verbosity")
			root.Flags().Bool("unsafe-trace", false, "")

			_, err := pcmd.ExecuteCommand(rootCmd.Command)
			require.NoError(t, err)
			require.False(t, envVarCalled)
		})
	}
}

func TestInitializeOnPremKafkaRest(t *testing.T) {
	cfg := config.AuthenticatedOnPremConfigMock()
	cfg.Context().State.AuthToken = validAuthToken
	ctrl := gomock.NewController(t)
	r := getPreRunBase(ctrl)
	r.Config = cfg
	cobraCmd := &cobra.Command{Use: "test"}
	cobraCmd.Flags().CountP("verbose", "v", "Increase verbosity")
	cobraCmd.Flags().Bool("unsafe-trace", false, "")
	c := pcmd.NewAuthenticatedCLICommand(cobraCmd, r)
	t.Run("InitializeOnPremKafkaRest_ValidMdsToken", func(t *testing.T) {
		err := r.InitializeOnPremKafkaRest(c)(c.Command, []string{})
		require.NoError(t, err)
		kafkaREST, err := c.GetKafkaREST(cobraCmd)
		require.NoError(t, err)
		auth, ok := kafkaREST.Context.Value(kafkarestv3.ContextAccessToken).(string)
		require.True(t, ok)
		require.Equal(t, validAuthToken, auth)
	})
	r.Config.Context().State.AuthToken = ""
	buf := new(bytes.Buffer)
	c.SetOut(buf)
}
