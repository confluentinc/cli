package mfa

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
	state, err := newState("https://devel.cpdev.cloud/", "test+mfa@conluent.io")
	require.NoError(t, err)
	// randomly generated
	require.True(t, len(state.CodeVerifier) > 10)
	require.True(t, len(state.CodeChallenge) > 10)
	require.True(t, len(state.MFAProviderState) > 10)
	// make sure we didn't so something dumb generating the hashes
	require.True(t,
		(state.CodeVerifier != state.CodeChallenge) &&
			(state.CodeVerifier != state.MFAProviderState) &&
			(state.CodeChallenge != state.MFAProviderState))
	// dev configs
	require.Equal(t, "https://login.confluent-dev.io/oauth", state.MFAProviderHost)
	require.Equal(t, "sPhOuMMVRSFFR7HfB606KLxf1RAU4SXg", state.MFAProviderClientID)
	require.Equal(t, "http://127.0.0.1:26635/cli_callback", state.MFAProviderCallbackUrl)
	require.Equal(t, "https://confluent-dev.auth0.com/api/v2/", state.MFAProviderIdentifier)
	require.Empty(t, state.MFAProviderAuthenticationCode)
	require.Empty(t, state.MFAProviderIDToken)

	// check stag configs
	stateStag, err := newState("https://stag.cpdev.cloud/", "test+mfa@conluent.io")
	require.NoError(t, err)
	require.Equal(t, "https://login-stag.confluent-dev.io/oauth", stateStag.MFAProviderHost)
	require.Equal(t, "8RxQmZEYtEDah4MTIIzl4hGGeFwdJS6w", stateStag.MFAProviderClientID)
	require.Equal(t, "http://127.0.0.1:26635/cli_callback", stateStag.MFAProviderCallbackUrl)
	require.Equal(t, "https://confluent-stag.auth0.com/api/v2/", stateStag.MFAProviderIdentifier)
	require.Empty(t, state.MFAProviderAuthenticationCode)
	require.Empty(t, state.MFAProviderIDToken)
}

func TestNewStateInvalidUrl(t *testing.T) {
	state, err := newState("Invalid url", "xyz.com")
	require.NoError(t, err)
	require.NotNil(t, state)
}

func TestGetAuthorizationUrl(t *testing.T) {
	state, err := newState("https://devel.cpdev.cloud", "test+mfa@confluent.io")
	require.NoError(t, err)

	// test get auth code url
	authCodeUrlDevel := state.getAuthorizationCodeUrl(false, "connection1")
	expectedUri := "/authorize?challenge_mfa=true" +
		"&response_type=code" +
		"&email=" + encodeEmail(state.Email) +
		"&from_cli=true&mfa_from_cli=true" +
		"&code_challenge=" + state.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + state.MFAProviderClientID +
		"&redirect_uri=" + state.MFAProviderCallbackUrl +
		"&scope=" + state.MFAProviderScope +
		"&state=" + state.MFAProviderState +
		"&audience=" + state.MFAProviderIdentifier +
		"&connection=connection1"
	expectedUrl := state.MFAProviderHost + expectedUri
	require.Equal(t, authCodeUrlDevel, expectedUrl)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		require.Equal(t, req.URL.String(), expectedUri)
		// Send response to be tested
		_, e := rw.Write([]byte(`OK`))
		require.NoError(t, e)
	}))
	defer server.Close()
	require.NotEmpty(t, server.URL)
}

func TestGetOAuthToken(t *testing.T) {
	mockRefreshToken := "foo"

	state, err := newState("https://devel.cpdev.cloud", "test+mfa@confluent.io")
	require.NoError(t, err)

	expectedUri := "/token"
	expectedPayload := "grant_type=authorization_code&from_cli=true&mfa_from_cli=true" +
		"&client_id=" + state.MFAProviderClientID +
		"&code_verifier=" + state.CodeVerifier +
		"&code=" + state.MFAProviderAuthenticationCode +
		"&redirect_uri=" + state.MFAProviderCallbackUrl

	mockIDToken := "foobar"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		require.Equal(t, req.URL.String(), expectedUri)
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.True(t, len(body) > 0)
		require.Equal(t, expectedPayload, string(body))

		// mock response
		_, err = rw.Write([]byte(fmt.Sprintf(`{"id_token": "%s", "refresh_token": "%s"}`, mockIDToken, mockRefreshToken)))
		require.NoError(t, err)
	}))
	defer server.Close()
	serverPort := strings.Split(server.URL, ":")[2]

	// mock auth0 endpoint with local test server
	state.MFAProviderHost = "http://127.0.0.1:" + serverPort

	err = state.getOAuthToken()
	require.NoError(t, err)

	require.Equal(t, mockIDToken, state.MFAProviderIDToken)
}

func TestRefreshOAuthToken(t *testing.T) {
	mockRefreshToken1 := "foo"
	mockRefreshToken2 := "bar"

	state, err := newState("https://devel.cpdev.cloud", "test+mfa@confluent.io")
	require.NoError(t, err)
	state.MFAProviderRefreshToken = mockRefreshToken1

	expectedUri := "/token"
	expectedPayload := "grant_type=refresh_token" +
		"&client_id=" + state.MFAProviderClientID +
		"&refresh_token=" + state.MFAProviderRefreshToken +
		"&redirect_uri=" + state.MFAProviderCallbackUrl

	mockIDToken := "foobar"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		require.Equal(t, req.URL.String(), expectedUri)
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.True(t, len(body) > 0)
		require.Equal(t, expectedPayload, string(body))

		// mock response
		_, err = rw.Write([]byte(fmt.Sprintf(`{"id_token": "%s", "refresh_token": "%s"}`, mockIDToken, mockRefreshToken2)))
		require.NoError(t, err)
	}))
	defer server.Close()
	serverPort := strings.Split(server.URL, ":")[2]

	// mock auth0 endpoint with local test server
	state.MFAProviderHost = "http://127.0.0.1:" + serverPort

	err = state.refreshOAuthToken()
	require.NoError(t, err)

	require.Equal(t, mockIDToken, state.MFAProviderIDToken)
	require.Equal(t, mockRefreshToken2, state.MFAProviderRefreshToken)
}
