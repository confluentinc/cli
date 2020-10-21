package auth

import (
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"testing"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"

	authMock "github.com/confluentinc/cli/internal/pkg/auth/mock"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

const (
	netrcFilePath         = "test_files/netrc"
	mockConfigUser        = "mock-user"
	mockConfigPassword    = "mock-password"
)

func TestUpateSSOToken(t *testing.T) {
	initialAuthToken := "initial-auth"
	finalAuthToken := "final-auth"
	mockCCloud := &authMock.MockCCloudTokenHandler{
		GetUserSSOFunc: func(client *ccloud.Client, email string) (user *orgv1.User, e error) {
			return &orgv1.User{}, nil
		},
		RefreshSSOTokenFunc: func(client *ccloud.Client, refreshToken, url string, logger *log.Logger) (s string, e error) {
			require.Equal(t, refreshToken, mockConfigPassword)
			return finalAuthToken, nil
		},
	}

	netrcHandler := netrc.NewNetrcHandler(netrcFilePath)

	updateTokenHandler := UpdateTokenHandlerImpl{
		ccloudTokenHandler: mockCCloud,
		netrcHandler:       netrcHandler,
	}

	cfg := v3.AuthenticatedCloudConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateCCloudAuthTokenUsingNetrcCredentials(ctx, "userAgent", log.New())
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
	netrcHandler := netrc.NewNetrcHandler(netrcFilePath)

	updateTokenHandler := UpdateTokenHandlerImpl{
		ccloudTokenHandler: mockCCloudTokenHandler,
		netrcHandler:       netrcHandler,
	}

	cfg := v3.AuthenticatedCloudConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateCCloudAuthTokenUsingNetrcCredentials(ctx, "userAgent", log.New())
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
	netrcHandler := netrc.NewNetrcHandler(netrcFilePath)

	updateTokenHandler := UpdateTokenHandlerImpl{
		confluentTokenHandler: mockConfluentTokenHandler,
		netrcHandler:          netrcHandler,
	}

	cfg := v3.AuthenticatedConfluentConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateConfluentAuthTokenUsingNetrcCredentials(ctx, log.New())
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
	netrcHandler := netrc.NewNetrcHandler(netrcFilePath)

	updateTokenHandler := UpdateTokenHandlerImpl{
		ccloudTokenHandler: mockCCloudTokenHandler,
		netrcHandler:       netrcHandler,
	}

	cfg := v3.AuthenticatedCloudConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateCCloudAuthTokenUsingNetrcCredentials(ctx, "userAgent", log.New())
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
	netrcHandler := netrc.NewNetrcHandler(netrcFilePath)

	updateTokenHandler := UpdateTokenHandlerImpl{
		confluentTokenHandler: mockConfluentTokenHandler,
		netrcHandler:          netrcHandler,
	}

	cfg := v3.AuthenticatedConfluentConfigMock()
	ctx := cfg.Context()
	ctx.State.AuthToken = initialAuthToken

	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
	err := updateTokenHandler.UpdateConfluentAuthTokenUsingNetrcCredentials(ctx, log.New())
	require.Error(t, err)

	require.True(t, mockConfluentTokenHandler.GetAuthTokenCalled())
	require.Equal(t, ctx.State.AuthToken, initialAuthToken)
}
