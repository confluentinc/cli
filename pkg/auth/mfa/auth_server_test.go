package mfa

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/auth/sso"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

func TestServerTimeout(t *testing.T) {
	state, err := newState("https://devel.cpdev.cloud", "test+mfa@confluent.io")
	require.NoError(t, err)
	server := newServer(state)

	require.NoError(t, server.startServer())

	err = server.awaitAuthorizationCode(1 * time.Second)
	require.Error(t, err)
	require.Equal(t, err.Error(), errors.BrowserAuthTimedOutErrorMsg)
	errors.VerifyErrorAndSuggestions(require.New(t), err, errors.BrowserAuthTimedOutErrorMsg, errors.BrowserAuthTimedOutSuggestions)

	//SSO Server
	stateSSO, err := sso.NewState("https://devel.cpdev.cloud", false)
	require.NoError(t, err)
	serverSSO := sso.NewServer(stateSSO)

	require.NoError(t, serverSSO.StartServer())

	err = serverSSO.AwaitAuthorizationCode(1 * time.Second)
	require.Error(t, err)
	require.Equal(t, err.Error(), errors.BrowserAuthTimedOutErrorMsg)
	errors.VerifyErrorAndSuggestions(require.New(t), err, errors.BrowserAuthTimedOutErrorMsg, errors.BrowserAuthTimedOutSuggestions)

}

func TestCallback(t *testing.T) {
	state, err := newState("https://devel.cpdev.cloud", "test+mfa@confluent.io")
	require.NoError(t, err)
	server := newServer(state)

	require.NoError(t, server.startServer())

	state.MFAProviderCallbackUrl = "http://127.0.0.1:26635/cli_callback"
	url := state.MFAProviderCallbackUrl
	mockCode := "uhlU7Fvq5NwLwBwk"
	mockUri := url + "?code=" + mockCode + "&state=" + state.MFAProviderState

	ch := make(chan bool)
	go func() {
		<-ch
		// send mock request to server's callback endpoint
		req, err := http.NewRequest(http.MethodGet, mockUri, nil)
		require.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
	}()

	go func() {
		// trigger the callback function after waiting a sec
		time.Sleep(500)
		close(ch)
	}()
	authCodeError := server.awaitAuthorizationCode(3 * time.Second)
	require.NoError(t, authCodeError)
	require.Equal(t, state.MFAProviderAuthenticationCode, "uhlU7Fvq5NwLwBwk")

	//SSO Callback
	stateSSO, err := sso.NewState("https://devel.cpdev.cloud", false)
	require.NoError(t, err)
	serverSSO := sso.NewServer(stateSSO)

	require.NoError(t, serverSSO.StartServer())

	stateSSO.SSOProviderCallbackUrl = "http://127.0.0.1:26635/cli_callback"
	urlSSO := stateSSO.SSOProviderCallbackUrl
	mockCodeSSO := "uhlU7Fvq5NwLwBwk"
	mockUriSSO := urlSSO + "?code=" + mockCodeSSO + "&state=" + stateSSO.SSOProviderState

	c := make(chan bool)
	go func() {
		<-c
		// send mock request to server's callback endpoint
		req, err := http.NewRequest(http.MethodGet, mockUriSSO, nil)
		require.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
	}()

	go func() {
		// trigger the callback function after waiting a sec
		time.Sleep(500)
		close(c)
	}()
	authCodeErrorSSO := serverSSO.AwaitAuthorizationCode(3 * time.Second)
	require.NoError(t, authCodeErrorSSO)
	require.Equal(t, stateSSO.SSOProviderAuthenticationCode, "uhlU7Fvq5NwLwBwk")
}
