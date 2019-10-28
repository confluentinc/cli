package sso

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestServerTimeout(t *testing.T) {
	configDevel := &config.Config{AuthURL: "https://devel.cpdev.cloud"}
	state, err := newState(configDevel)
	require.NoError(t, err)
	server := newServer(state)

	require.NoError(t, server.startServer())

	err = server.awaitAuthorizationCode(1 * time.Second)
	require.Error(t, err)
	require.Equal(t, err.Error(), "timed out while waiting for browser authentication to occur; please try logging in again")
}

func TestCallback(t *testing.T) {
	configDevel := &config.Config{AuthURL: "https://devel.cpdev.cloud"}
	state, err := newState(configDevel)
	require.NoError(t, err)
	server := newServer(state)

	require.NoError(t, server.startServer())

	state.SSOProviderCallbackUrl = "http://127.0.0.1:26635/cli_callback"
	url := state.SSOProviderCallbackUrl
	mockCode := "uhlU7Fvq5NwLwBwk"
	mockUri := url + "?code="+mockCode+"&state="+state.SSOProviderState

	ch := make(chan bool)
	go func() {
		_ = <- ch
		// send mock request to server's callback endpoint
		req, err := http.NewRequest("GET", mockUri, nil)
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
	require.Equal(t, state.SSOProviderAuthenticationCode, "uhlU7Fvq5NwLwBwk")
}
