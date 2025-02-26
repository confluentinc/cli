package sso

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/confluentinc/cli/v4/pkg/output"
)

func Login(authURL string, noBrowser bool, connectionName string) (string, string, error) {
	state, err := NewState(authURL, noBrowser)
	if err != nil {
		return "", "", err
	}

	isOkta := IsOkta(authURL)

	if noBrowser {
		// no browser flag does not need to launch the server
		// it prints the url and has the user copy this into their browser instead
		output.Println(false, "Navigate to the following link in your browser to authenticate:")
		output.Println(false, state.getAuthorizationCodeUrl(connectionName, isOkta))
		output.Println(false, "")
		output.Println(false, "After authenticating in your browser, paste the code here:")

		// wait for the user to paste the code
		// the code should come in the format {state}/{auth0_auth_code}
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()
		split := strings.SplitAfterN(input, "/", 2)
		if len(split) < 2 {
			return "", "", fmt.Errorf("pasted input had invalid format")
		}
		auth0State := strings.Replace(split[0], "/", "", 1)
		if !(auth0State == state.SSOProviderState) {
			return "", "", fmt.Errorf("authentication code either did not contain a state parameter or the state parameter was invalid; login will fail")
		}

		state.SSOProviderAuthenticationCode = split[1]
	} else {
		// we need to start a background HTTP server to support the authorization code flow with PKCE
		// described at https://auth0.com/docs/flows/guides/auth-code-pkce/call-api-auth-code-pkce
		server := NewServer(state)
		if err := server.StartServer(); err != nil {
			return "", "", fmt.Errorf("unable to start HTTP server: %w", err)
		}

		// Get authorization code for making subsequent token request
		url := state.getAuthorizationCodeUrl(connectionName, isOkta)
		if err := browser.OpenURL(url); err != nil {
			return "", "", fmt.Errorf("unable to open web browser for authorization: %w", err)
		}

		if err := server.AwaitAuthorizationCode(30 * time.Second); err != nil {
			return "", "", err
		}
	}

	// Exchange authorization code for OAuth token from SSO provider
	if err := state.getOAuthToken(); err != nil {
		return "", "", err
	}

	return state.SSOProviderIDToken, state.SSOProviderRefreshToken, nil
}

func RefreshTokens(authURL, refreshToken string) (string, string, error) {
	state, err := NewState(authURL, false)
	if err != nil {
		return "", "", err
	}
	state.SSOProviderRefreshToken = refreshToken

	if err := state.refreshOAuthToken(); err != nil {
		return "", "", err
	}

	return state.SSOProviderIDToken, state.SSOProviderRefreshToken, nil
}
