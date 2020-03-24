package auth

import (
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	authMock "github.com/confluentinc/cli/internal/pkg/auth/mock"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	testUtils "github.com/confluentinc/cli/test"
)

var (
	netrcFilePath           = "test_files/netrc"
	outputFileMds     = "test_files/output-mds"
	outputFileCcloudLogin = "test_files/output-ccloud-login"
	outputFileCcloudSSO = "test_files/output-ccloud-sso"
	mdsContext        = "mds-context"
	ccloudLoginContext = "ccloud-login"
	ccloudSSOContext = "ccloud-sso"
	netrcUser               = "jamal@jj"
	netrcPassword           = "12345"
	mockConfigUser          = "mock-user"
	mockConfigPassword      = "mock-password"
)

func TestNetRCCredentialReader(t *testing.T) {
	tests := []struct {
		name        string
		want        []string
		cliName     string
		isSSO       bool
		contextName string
		wantErr     bool
		file        string
	}{
		{
			name:        "mds context",
			want:        []string{netrcUser, netrcPassword},
			contextName: mdsContext,
			cliName:     "confluent",
			file:        netrcFilePath,
		},
		{
			name:        "ccloud login context",
			want:        []string{netrcUser, netrcPassword},
			contextName: ccloudLoginContext,
			cliName:     "ccloud",
			file:        netrcFilePath,
		},
		{
			name:        "ccloud sso context",
			want:        []string{netrcUser, netrcPassword},
			contextName: ccloudSSOContext,
			cliName:     "ccloud",
			isSSO:       true,
			file:        netrcFilePath,
		},
		{
			name:        "No file error",
			contextName: mdsContext,
			cliName:     "confluent",
			wantErr:     true,
			file:        "wrong-file",
		},
		{
			name:        "Context doesn't exist",
			contextName: "non-existing-context",
			cliName:     "ccloud",
			wantErr:     true,
			file:        netrcFilePath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netrcHandler := netrcHandler{fileName: tt.file}
			var username, password string
			var err error
			if username, password, err = netrcHandler.getNetrcCredentials(tt.cliName, tt.isSSO, tt.contextName); (err != nil) != tt.wantErr {
				t.Errorf("getNetrcCredentials error = %+v, wantErr %+v", err, tt.wantErr)
			}
			if len(tt.want) != 0 && !t.Failed() && username != tt.want[0] {
				t.Errorf("getNetrcCredenials username = %+v, want %+v", username, tt.want[0])
			}
			if len(tt.want) == 2 && !t.Failed() && password != tt.want[1] {
				t.Errorf("getNetrcCredenials password = %+v, want %+v", password, tt.want[1])
			}
		})
	}
}

func TestNetrcWriter(t *testing.T) {
	tests := []struct {
		name        string
		wantFile    string
		cliName     string
		isSSO       bool
		contextName string
		wantErr     bool
	}{
		{
			name:        "mds context",
			wantFile:    outputFileMds,
			contextName: mdsContext,
			cliName:     "confluent",
		},
		{
			name:        "ccloud login context",
			wantFile:    outputFileCcloudLogin,
			contextName: ccloudLoginContext,
			cliName:     "ccloud",
		},
		{
			name:        "ccloud sso context",
			wantFile:    outputFileCcloudSSO,
			contextName: ccloudSSOContext,
			cliName:     "ccloud",
			isSSO:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, _ := ioutil.TempFile("", "tempNetrc.json")
			netrcHandler := netrcHandler{fileName:tempFile.Name()}
			err := netrcHandler.WriteNetrcCredentials(tt.cliName, tt.isSSO, tt.contextName, netrcUser, netrcPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteNetrcCredentials error = %+v, wantErr %+v", err, tt.wantErr)
			}
			gotBytes, err := ioutil.ReadFile(tempFile.Name())
			require.NoError(t, err)
			got := testUtils.NormalizeNewLines(string(gotBytes))

			wantBytes, err := ioutil.ReadFile(tt.wantFile)
			require.NoError(t, err)
			want := testUtils.NormalizeNewLines(string(wantBytes))

			if got != want {
				t.Errorf("WriteNetrcCredentials = \n%s\n want = \n%s\n", got, want)
			}
			_ = os.Remove(tempFile.Name())
		})
	}
}

func TestUpateSSOToken(t *testing.T) {
	initialAuthToken := "initial-auth"
	finalAuthToken := "final-auth"
	mockCCloud := &authMock.MockCCloudTokenHandler{
		GetUserSSOFunc: func(client *ccloud.Client, email string) (user *orgv1.User, e error) {
			return &orgv1.User{}, nil
		},
		RefreshSSOTokenFunc: func(client *ccloud.Client, refreshToken, url string) (s string, e error) {
			require.Equal(t, refreshToken, mockConfigPassword)
			return finalAuthToken, nil
		},
	}

	netrcHandler := &netrcHandler{fileName: netrcFilePath}

	updateTokenHandler := UpdateTokenHandlerImpl{
		ccloudTokenHandler: mockCCloud,
		netrcHandler:       netrcHandler,
	}

	cfg := v3.AuthenticatedCloudConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateCCloudAuthToken(ctx, "userAgent", log.New())
	require.NoError(t, err)

	require.True(t, mockCCloud.GetUserSSOCalled())
	require.True(t, mockCCloud.RefreshSSOTokenCalled())

	require.Equal(t, ctx.State.AuthToken, finalAuthToken)
}

func TestUpdateCloudLoginCredentialsToken(t *testing.T) {
	initialAuthToken := "initial-auth"
	finalAuthToken := "final-auth"

	mockCCloudTokenHandler := &authMock.MockCCloudTokenHandler{
		GetUserSSOFunc: func(client *ccloud.Client, email string) (user *orgv1.User, e error) {
			return nil, nil
		},
		GetCredentialsTokenFunc: func(client *ccloud.Client, email, password string) (s string, e error) {
			require.Equal(t, email, mockConfigUser)
			require.Equal(t, password, mockConfigPassword)
			return finalAuthToken, nil
		},
	}
	netrcHandler := &netrcHandler{fileName: netrcFilePath}

	updateTokenHandler := UpdateTokenHandlerImpl{
		ccloudTokenHandler: mockCCloudTokenHandler,
		netrcHandler:       netrcHandler,
	}

	cfg := v3.AuthenticatedCloudConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateCCloudAuthToken(ctx, "userAgent", log.New())
	require.NoError(t, err)

	require.True(t, mockCCloudTokenHandler.GetUserSSOCalled())
	require.True(t, mockCCloudTokenHandler.GetCredentialsTokenCalled())

	require.Equal(t, ctx.State.AuthToken, finalAuthToken)
}

func TestUpdateConfluent(t *testing.T) {
	initialAuthToken := "initial-auth"
	finalAuthToken := "final-auth"
	mockConfluentTokenHandler := &authMock.MockConfluentTokenHandler{
		GetAuthTokenFunc: func(mdsClient *mds.APIClient, email, password string) (s string, e error) {
			require.Equal(t, email, mockConfigUser)
			require.Equal(t, password, mockConfigPassword)
			return finalAuthToken, nil
		},
	}
	netrcHandler := &netrcHandler{fileName: netrcFilePath}

	updateTokenHandler := UpdateTokenHandlerImpl{
		confluentTokenHandler: mockConfluentTokenHandler,
		netrcHandler:          netrcHandler,
	}

	cfg := v3.AuthenticatedConfluentConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateConfluentAuthToken(ctx, log.New())
	require.NoError(t, err)

	require.True(t, mockConfluentTokenHandler.GetAuthTokenCalled())
	require.Equal(t, ctx.State.AuthToken, finalAuthToken)
}

func TestFailedCCloudUpdate(t *testing.T) {
	initialAuthToken := "initial-auth"

	mockCCloudTokenHandler := &authMock.MockCCloudTokenHandler{
		GetUserSSOFunc: func(client *ccloud.Client, email string) (user *orgv1.User, e error) {
			return nil, nil
		},
		GetCredentialsTokenFunc: func(client *ccloud.Client, email, password string) (s string, e error) {
			return "", errors.Errorf("Failed to get auth token")
		},
	}
	netrcHandler := &netrcHandler{fileName: netrcFilePath}

	updateTokenHandler := UpdateTokenHandlerImpl{
		ccloudTokenHandler: mockCCloudTokenHandler,
		netrcHandler:       netrcHandler,
	}

	cfg := v3.AuthenticatedCloudConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateCCloudAuthToken(ctx, "userAgent", log.New())
	require.Error(t, err)

	require.True(t, mockCCloudTokenHandler.GetUserSSOCalled())
	require.True(t, mockCCloudTokenHandler.GetCredentialsTokenCalled())

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
}

func TestFailedConfluentUpdate(t *testing.T) {
	initialAuthToken := "initial-auth"
	mockConfluentTokenHandler := &authMock.MockConfluentTokenHandler{
		GetAuthTokenFunc: func(mdsClient *mds.APIClient, email, password string) (s string, e error) {
			return "", errors.Errorf("Failed to get auth token")
		},
	}
	netrcHandler := &netrcHandler{fileName: netrcFilePath}

	updateTokenHandler := UpdateTokenHandlerImpl{
		confluentTokenHandler: mockConfluentTokenHandler,
		netrcHandler:          netrcHandler,
	}

	cfg := v3.AuthenticatedConfluentConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateConfluentAuthToken(ctx, log.New())
	require.Error(t, err)

	require.True(t, mockConfluentTokenHandler.GetAuthTokenCalled())
	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
}
