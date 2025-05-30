package sso

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStateDev(t *testing.T) {
	state, err := NewState("https://devel.cpdev.cloud/", false)
	require.NoError(t, err)
	require.True(t, len(state.CodeVerifier) > 10)
	require.True(t, len(state.CodeChallenge) > 10)
	require.True(t, len(state.SSOProviderState) > 10)
	require.True(t,
		(state.CodeVerifier != state.CodeChallenge) &&
			(state.CodeVerifier != state.SSOProviderState) &&
			(state.CodeChallenge != state.SSOProviderState))
	require.Equal(t, "https://login.confluent-dev.io/oauth", state.SSOProviderHost)
	require.Equal(t, "sPhOuMMVRSFFR7HfB606KLxf1RAU4SXg", state.SSOProviderClientID)
	require.Equal(t, "http://127.0.0.1:26635/cli_callback", state.SSOProviderCallbackUrl)
	require.Equal(t, "https://confluent-dev.auth0.com/api/v2/", state.SSOProviderIdentifier)
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)

	stateStag, err := NewState("https://stag.cpdev.cloud/", false)
	require.NoError(t, err)
	require.Equal(t, "https://login-stag.confluent-dev.io/oauth", stateStag.SSOProviderHost)
	require.Equal(t, "8RxQmZEYtEDah4MTIIzl4hGGeFwdJS6w", stateStag.SSOProviderClientID)
	require.Equal(t, "http://127.0.0.1:26635/cli_callback", stateStag.SSOProviderCallbackUrl)
	require.Equal(t, "https://confluent-stag.auth0.com/api/v2/", stateStag.SSOProviderIdentifier)
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)
}

func TestNewStateDevNoBrowser(t *testing.T) {
	state, err := NewState("https://devel.cpdev.cloud", true)
	require.NoError(t, err)
	require.True(t, len(state.CodeVerifier) > 10)
	require.True(t, len(state.CodeChallenge) > 10)
	require.True(t, len(state.SSOProviderState) > 10)
	require.True(t,
		(state.CodeVerifier != state.CodeChallenge) &&
			(state.CodeVerifier != state.SSOProviderState) &&
			(state.CodeChallenge != state.SSOProviderState))

	require.Equal(t, "https://login.confluent-dev.io/oauth", state.SSOProviderHost)
	require.Equal(t, "sPhOuMMVRSFFR7HfB606KLxf1RAU4SXg", state.SSOProviderClientID)
	require.Equal(t, "https://devel.cpdev.cloud/cli_callback", state.SSOProviderCallbackUrl)
	require.Equal(t, "https://confluent-dev.auth0.com/api/v2/", state.SSOProviderIdentifier)
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)

	stateStag, err := NewState("https://stag.cpdev.cloud", true)
	require.NoError(t, err)
	require.Equal(t, "https://login-stag.confluent-dev.io/oauth", stateStag.SSOProviderHost)
	require.Equal(t, "8RxQmZEYtEDah4MTIIzl4hGGeFwdJS6w", stateStag.SSOProviderClientID)
	require.Equal(t, "https://stag.cpdev.cloud/cli_callback", stateStag.SSOProviderCallbackUrl)
	require.Equal(t, "https://confluent-stag.auth0.com/api/v2/", stateStag.SSOProviderIdentifier)
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)
}

func TestNewStateProd(t *testing.T) {
	state, err := NewState("https://confluent.cloud", false)
	require.NoError(t, err)
	require.True(t, len(state.CodeVerifier) > 10)
	require.True(t, len(state.CodeChallenge) > 10)
	require.True(t, len(state.SSOProviderState) > 10)
	require.True(t,
		(state.CodeVerifier != state.CodeChallenge) &&
			(state.CodeVerifier != state.SSOProviderState) &&
			(state.CodeChallenge != state.SSOProviderState))
	require.Equal(t, state.SSOProviderHost, "https://login.confluent.io/oauth")
	require.Equal(t, state.SSOProviderClientID, "oX2nvSKl5jvBKVgwehZfvR4K8RhsZIEs")
	require.Equal(t, state.SSOProviderCallbackUrl, "http://127.0.0.1:26635/cli_callback")
	require.Equal(t, state.SSOProviderIdentifier, "https://confluent.auth0.com/api/v2/")
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)
}

func TestNewStateProdNoBrowser(t *testing.T) {
	for _, authURL := range []string{"", "https://confluent.cloud"} {
		state, err := NewState(authURL, true)
		require.NoError(t, err)

		require.True(t, len(state.CodeVerifier) > 10)
		require.True(t, len(state.CodeChallenge) > 10)
		require.True(t, len(state.SSOProviderState) > 10)

		require.True(t,
			(state.CodeVerifier != state.CodeChallenge) &&
				(state.CodeVerifier != state.SSOProviderState) &&
				(state.CodeChallenge != state.SSOProviderState))

		require.Equal(t, "https://login.confluent.io/oauth", state.SSOProviderHost)
		require.Equal(t, "oX2nvSKl5jvBKVgwehZfvR4K8RhsZIEs", state.SSOProviderClientID)
		require.Equal(t, "https://confluent.cloud/cli_callback", state.SSOProviderCallbackUrl)
		require.Equal(t, "https://confluent.auth0.com/api/v2/", state.SSOProviderIdentifier)
		require.Empty(t, state.SSOProviderAuthenticationCode)
		require.Empty(t, state.SSOProviderIDToken)
	}
}

func TestNewStateInvalidUrl(t *testing.T) {
	state, err := NewState("Invalid url", true)
	require.NoError(t, err)
	require.NotNil(t, state)
}

func TestGetAuthorizationUrl(t *testing.T) {
	state, err := NewState("https://devel.cpdev.cloud", false)
	require.NoError(t, err)

	authCodeUrlDevel := state.getAuthorizationCodeUrl("foo", false)
	expectedUri := "/authorize?" +
		"response_type=code" +
		"&code_challenge=" + state.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + state.SSOProviderClientID +
		"&redirect_uri=" + state.SSOProviderCallbackUrl +
		"&scope=email%20openid%20offline_access" +
		"&state=" + state.SSOProviderState +
		"&audience=" + state.SSOProviderIdentifier +
		"&connection=foo"
	expectedUrl := state.SSOProviderHost + expectedUri
	require.Equal(t, authCodeUrlDevel, expectedUrl)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, req.URL.String(), expectedUri)
		_, e := rw.Write([]byte(`OK`))
		require.NoError(t, e)
	}))
	defer server.Close()
	require.NotEmpty(t, server.URL)
}

func TestGetOAuthToken(t *testing.T) {
	mockRefreshToken := "foo"

	state, err := NewState("https://devel.cpdev.cloud", false)
	require.NoError(t, err)

	expectedUri := "/token"
	expectedPayload := "grant_type=authorization_code" +
		"&client_id=" + state.SSOProviderClientID +
		"&code_verifier=" + state.CodeVerifier +
		"&code=" + state.SSOProviderAuthenticationCode +
		"&redirect_uri=" + state.SSOProviderCallbackUrl

	mockIDToken := "foobar"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, req.URL.String(), expectedUri)
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.True(t, len(body) > 0)
		require.Equal(t, expectedPayload, string(body))

		_, err = rw.Write([]byte(fmt.Sprintf(`{"id_token": "%s", "refresh_token": "%s"}`, mockIDToken, mockRefreshToken)))
		require.NoError(t, err)
	}))
	defer server.Close()
	serverPort := strings.Split(server.URL, ":")[2]

	state.SSOProviderHost = "http://127.0.0.1:" + serverPort

	err = state.getOAuthToken()
	require.NoError(t, err)

	require.Equal(t, mockIDToken, state.SSOProviderIDToken)
}

func TestRefreshOAuthToken(t *testing.T) {
	mockRefreshToken1 := "foo"
	mockRefreshToken2 := "bar"

	state, err := NewState("https://devel.cpdev.cloud", false)
	require.NoError(t, err)
	state.SSOProviderRefreshToken = mockRefreshToken1

	expectedUri := "/token"
	expectedPayload := "grant_type=refresh_token" +
		"&client_id=" + state.SSOProviderClientID +
		"&refresh_token=" + state.SSOProviderRefreshToken +
		"&redirect_uri=" + state.SSOProviderCallbackUrl

	mockIDToken := "foobar"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, req.URL.String(), expectedUri)
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.True(t, len(body) > 0)
		require.Equal(t, expectedPayload, string(body))

		_, err = rw.Write([]byte(fmt.Sprintf(`{"id_token": "%s", "refresh_token": "%s"}`, mockIDToken, mockRefreshToken2)))
		require.NoError(t, err)
	}))
	defer server.Close()
	serverPort := strings.Split(server.URL, ":")[2]
	state.SSOProviderHost = "http://127.0.0.1:" + serverPort

	err = state.refreshOAuthToken()
	require.NoError(t, err)

	require.Equal(t, mockIDToken, state.SSOProviderIDToken)
	require.Equal(t, mockRefreshToken2, state.SSOProviderRefreshToken)
}
