package sso

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStateDev(t *testing.T) {
	state, err := newState("https://devel.cpdev.cloud", false)
	require.NoError(t, err)
	// randomly generated
	require.True(t, len(state.CodeVerifier) > 10)
	require.True(t, len(state.CodeChallenge) > 10)
	require.True(t, len(state.SSOProviderState) > 10)
	// make sure we didn't so something dumb generating the hashes
	require.True(t,
		(state.CodeVerifier != state.CodeChallenge) &&
			(state.CodeVerifier != state.SSOProviderState) &&
			(state.CodeChallenge != state.SSOProviderState))
	// dev configs
	require.Equal(t, state.SSOProviderHost, "https://login.confluent-dev.io")
	require.Equal(t, state.SSOProviderClientID, "XKlqgOEo39iyonTl3Yv3IHWIXGKDP3fA")
	require.Equal(t, state.SSOProviderCallbackUrl, "http://127.0.0.1:26635/cli_callback")
	require.Equal(t, state.SSOProviderIdentifier, "https://confluent-dev.auth0.com/api/v2/")
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)

	// check stag configs
	stateStag, err := newState("https://stag.cpdev.cloud", false)
	require.NoError(t, err)
	// configs for devel and staging are the same
	require.Equal(t, state.SSOProviderHost, stateStag.SSOProviderHost)
	require.Equal(t, state.SSOProviderClientID, stateStag.SSOProviderClientID)
	require.Equal(t, state.SSOProviderCallbackUrl, stateStag.SSOProviderCallbackUrl)
	require.Equal(t, state.SSOProviderIdentifier, stateStag.SSOProviderIdentifier)
	require.Empty(t, stateStag.SSOProviderAuthenticationCode)
	require.Empty(t, stateStag.SSOProviderIDToken)
}

func TestNewStateDevNoBrowser(t *testing.T) {
	state, err := newState("https://devel.cpdev.cloud", true)
	require.NoError(t, err)
	// randomly generated
	require.True(t, len(state.CodeVerifier) > 10)
	require.True(t, len(state.CodeChallenge) > 10)
	require.True(t, len(state.SSOProviderState) > 10)
	// make sure we didn't so something dumb generating the hashes
	require.True(t,
		(state.CodeVerifier != state.CodeChallenge) &&
			(state.CodeVerifier != state.SSOProviderState) &&
			(state.CodeChallenge != state.SSOProviderState))

	// dev configs
	require.Equal(t, state.SSOProviderHost, "https://login.confluent-dev.io")
	require.Equal(t, state.SSOProviderClientID, "XKlqgOEo39iyonTl3Yv3IHWIXGKDP3fA")
	require.Equal(t, state.SSOProviderCallbackUrl, "https://devel.cpdev.cloud/cli_callback")
	require.Equal(t, state.SSOProviderIdentifier, "https://confluent-dev.auth0.com/api/v2/")
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)

	// check stag configs
	stateStag, err := newState("https://stag.cpdev.cloud", true)
	require.NoError(t, err)
	// configs for devel and staging are the same except the callback url
	require.Equal(t, state.SSOProviderHost, stateStag.SSOProviderHost)
	require.Equal(t, state.SSOProviderClientID, stateStag.SSOProviderClientID)
	require.Equal(t, "https://stag.cpdev.cloud/cli_callback", stateStag.SSOProviderCallbackUrl)
	require.Equal(t, state.SSOProviderIdentifier, stateStag.SSOProviderIdentifier)
	require.Empty(t, stateStag.SSOProviderAuthenticationCode)
	require.Empty(t, stateStag.SSOProviderIDToken)

	// check cpd configs
	stateCpd, err := newState("https://aware-monkfish.gcp.priv.cpdev.cloud", true)
	require.NoError(t, err)
	// configs for cpd and devel are the same except the callback url
	require.Equal(t, state.SSOProviderHost, stateCpd.SSOProviderHost)
	require.Equal(t, state.SSOProviderClientID, stateCpd.SSOProviderClientID)
	require.Equal(t, "https://aware-monkfish.gcp.priv.cpdev.cloud/cli_callback", stateCpd.SSOProviderCallbackUrl)
	require.Equal(t, state.SSOProviderIdentifier, stateCpd.SSOProviderIdentifier)
	require.Empty(t, stateCpd.SSOProviderAuthenticationCode)
	require.Empty(t, stateCpd.SSOProviderIDToken)
}

func TestNewStateProd(t *testing.T) {
	state, err := newState("https://confluent.cloud", false)
	require.NoError(t, err)
	// randomly generated
	require.True(t, len(state.CodeVerifier) > 10)
	require.True(t, len(state.CodeChallenge) > 10)
	require.True(t, len(state.SSOProviderState) > 10)
	// make sure we didn't so something dumb generating the hashes
	require.True(t,
		(state.CodeVerifier != state.CodeChallenge) &&
			(state.CodeVerifier != state.SSOProviderState) &&
			(state.CodeChallenge != state.SSOProviderState))
	require.Equal(t, state.SSOProviderHost, "https://login.confluent.io")
	require.Equal(t, state.SSOProviderClientID, "hPbGZM8G55HSaUsaaieiiAprnJaEc9rH")
	require.Equal(t, state.SSOProviderCallbackUrl, "http://127.0.0.1:26635/cli_callback")
	require.Equal(t, state.SSOProviderIdentifier, "https://confluent.auth0.com/api/v2/")
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)
}

func TestNewStateProdNoBrowser(t *testing.T) {
	state, err := newState("https://confluent.cloud", true)
	require.NoError(t, err)
	// randomly generated
	require.True(t, len(state.CodeVerifier) > 10)
	require.True(t, len(state.CodeChallenge) > 10)
	require.True(t, len(state.SSOProviderState) > 10)
	// make sure we didn't so something dumb generating the hashes
	require.True(t,
		(state.CodeVerifier != state.CodeChallenge) &&
			(state.CodeVerifier != state.SSOProviderState) &&
			(state.CodeChallenge != state.SSOProviderState))

	require.Equal(t, state.SSOProviderHost, "https://login.confluent.io")
	require.Equal(t, state.SSOProviderClientID, "hPbGZM8G55HSaUsaaieiiAprnJaEc9rH")
	require.Equal(t, state.SSOProviderCallbackUrl, "https://confluent.cloud/cli_callback")
	require.Equal(t, state.SSOProviderIdentifier, "https://confluent.auth0.com/api/v2/")
	require.Empty(t, state.SSOProviderAuthenticationCode)
	require.Empty(t, state.SSOProviderIDToken)
}

func TestGetAuthorizationUrl(t *testing.T) {
	state, _ := newState("https://devel.cpdev.cloud", false)

	// test get auth code url
	authCodeUrlDevel := state.getAuthorizationCodeUrl("foo")
	expectedUri := "/authorize?" +
		"response_type=code" +
		"&code_challenge=" + state.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + state.SSOProviderClientID +
		"&redirect_uri=" + state.SSOProviderCallbackUrl +
		"&scope=email%20openid" +
		"&audience=" + state.SSOProviderIdentifier +
		"&state=" + state.SSOProviderState +
		"&connection=foo"
	expectedUrl := state.SSOProviderHost + expectedUri
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
	state, _ := newState("https://devel.cpdev.cloud", false)

	expectedUri := "/oauth/token"
	expectedPayload := "grant_type=authorization_code" +
		"&client_id=" + state.SSOProviderClientID +
		"&code_verifier=" + state.CodeVerifier +
		"&code=" + state.SSOProviderAuthenticationCode +
		"&redirect_uri=" + state.SSOProviderCallbackUrl

	mockIDToken := "foobar"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		require.Equal(t, req.URL.String(), expectedUri)
		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		require.True(t, len(body) > 0)
		require.Equal(t, expectedPayload, string(body))

		// mock response
		_, err = rw.Write([]byte(`{"id_token": "` + mockIDToken + `"}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	serverPort := strings.Split(server.URL, ":")[2]

	// mock auth0 endpoint with local test server
	state.SSOProviderHost = "http://127.0.0.1:" + serverPort

	err := state.getOAuthToken()
	require.NoError(t, err)

	require.Equal(t, mockIDToken, state.SSOProviderIDToken)
}
