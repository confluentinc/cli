package auth

import (
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	authMock "github.com/confluentinc/cli/internal/pkg/auth/mock"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

var (
	netrcFilePath = "test_files/netrc"
	netrcContextName = "existing-context"
	netrcUser = "existing-user"
	netrcPassword = "existing-password"
	netrcUser2 = "mock-user"
	netrcPassword2 = "mock-password"
)

func TestNetRCReader(t *testing.T) {
	tests := []struct {
		name    string
		want    []string
		contextName string
		wantErr bool
		file    string
	}{
		{
			name: "Context exist",
			want: []string{netrcUser, netrcPassword},
			contextName: netrcContextName,
			file: netrcFilePath,
		},
		{
			name: "No file error",
			contextName: netrcContextName,
			wantErr: true,
			file: "wrong-file",
		},
		{
			name: "Context doesn't exist",
			contextName: "non-existing-context",
			wantErr: true,
			file: netrcFilePath,
		},
		{
			name: "Context exist with no password",
			want: []string{netrcUser, ""},
			contextName: "no-password-context",
			file: netrcFilePath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netrcHandler := netrcHandler{fileName:tt.file}
			var username, password string
			var err error
			if username, password, err = netrcHandler.getNetrcCredentials(tt.contextName); (err != nil) != tt.wantErr {
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

func TestUpateSSOToken(t *testing.T) {
	initialAuthToken := "initial-auth"
	finalAuthToken := "final-auth"
	mockCCloud := &authMock.MockCCloudTokenHandler{
		GetUserSSOFunc: func(client *ccloud.Client, email string) (user *orgv1.User, e error) {
			return &orgv1.User{}, nil
		},
		RefreshSSOTokenFunc: func(client *ccloud.Client, ctx *v3.Context, url string) (s string, e error) {
			return finalAuthToken, nil
		},
	}

	updateTokenHandler := UpdateTokenHandlerImpl{
		ccloudTokenHandler:    mockCCloud,
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

func TestUpateCloudLoginCredentialsToken(t *testing.T) {
	initialAuthToken := "initial-auth"
	finalAuthToken := "final-auth"

	mockCCloudTokenHandler := &authMock.MockCCloudTokenHandler{
		GetUserSSOFunc: func(client *ccloud.Client, email string) (user *orgv1.User, e error) {
			return nil, nil
		},
		GetCredentialsTokenFunc: func(client *ccloud.Client, email, password string) (s string, e error) {
			require.Equal(t, email, netrcUser2)
			require.Equal(t, password, netrcPassword2)
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
			require.Equal(t, email, netrcUser2)
			require.Equal(t, password, netrcPassword2)
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
