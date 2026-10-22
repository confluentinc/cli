package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/config"
)

// newAccessTokensTestContext returns an authenticated cloud context whose
// platform server points at the given test server, so the /api/access_tokens
// exchange is served by the handler under test.
func newAccessTokensTestContext(t *testing.T, server *httptest.Server) *config.Context {
	t.Helper()
	ctx := config.AuthenticatedCloudConfigMock().Context()
	ctx.Platform.Server = server.URL
	ctx.GetState().AuthToken = "session-token"
	return ctx
}

func newAccessTokensServer(t *testing.T, response map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/access_tokens", r.URL.Path)
		require.Equal(t, "Bearer session-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(response))
	}))
}

func TestGetAccessTokens_ReturnsBothTokens(t *testing.T) {
	server := newAccessTokensServer(t, map[string]string{
		"token":          "dataplane-token",
		"regional_token": "regional-token",
	})
	defer server.Close()
	ctx := newAccessTokensTestContext(t, server)

	dataplaneToken, err := GetDataplaneToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "dataplane-token", dataplaneToken)

	regionalToken, err := GetRegionalToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "regional-token", regionalToken)
}

func TestGetRegionalToken_MissingRegionalToken(t *testing.T) {
	// A response without regional_token must still satisfy existing
	// GetDataplaneToken callers, but GetRegionalToken must fail explicitly
	// rather than return an empty credential.
	server := newAccessTokensServer(t, map[string]string{"token": "dataplane-token"})
	defer server.Close()
	ctx := newAccessTokensTestContext(t, server)

	dataplaneToken, err := GetDataplaneToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "dataplane-token", dataplaneToken)

	regionalToken, err := GetRegionalToken(ctx)
	require.ErrorContains(t, err, "no regional token")
	require.Empty(t, regionalToken)
}

func TestGetAccessTokens_ErrorResponse(t *testing.T) {
	server := newAccessTokensServer(t, map[string]string{"error": "token exchange failed"})
	defer server.Close()
	ctx := newAccessTokensTestContext(t, server)

	_, err := GetDataplaneToken(ctx)
	require.EqualError(t, err, "token exchange failed")

	_, err = GetRegionalToken(ctx)
	require.EqualError(t, err, "token exchange failed")
}
