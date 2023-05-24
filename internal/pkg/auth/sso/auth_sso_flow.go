package sso

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func Login(authURL string, noBrowser bool, connectionName string) (string, string, error) {
	state, err := newState(authURL, noBrowser)
	if err != nil {
		return "", "", err
	}

	isOkta := IsOkta(authURL)

	if noBrowser {
		// no browser flag does not need to launch the server
		// it prints the url and has the user copy this into their browser instead
		url := state.getAuthorizationCodeUrl(connectionName, isOkta)
		fmt.Printf(errors.NoBrowserSSOInstructionsMsg, url)

		// wait for the user to paste the code
		// the code should come in the format {state}/{auth0_auth_code}
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()
		split := strings.SplitAfterN(input, "/", 2)
		if len(split) < 2 {
			return "", "", errors.New(errors.PastedInputErrorMsg)
		}
		auth0State := strings.Replace(split[0], "/", "", 1)
		if !(auth0State == state.SSOProviderState) {
			return "", "", errors.New(errors.LoginFailedStateParamErrorMsg)
		}

		state.SSOProviderAuthenticationCode = split[1]
	} else {
		// we need to start a background HTTP server to support the authorization code flow with PKCE
		// described at https://auth0.com/docs/flows/guides/auth-code-pkce/call-api-auth-code-pkce
		server := newServer(state)
		if err := server.startServer(); err != nil {
			return "", "", errors.Wrap(err, errors.StartHTTPServerErrorMsg)
		}

		// Get authorization code for making subsequent token request
		if err := browser.OpenURL(state.getAuthorizationCodeUrl(connectionName, isOkta)); err != nil {
			return "", "", errors.Wrap(err, errors.OpenWebBrowserErrorMsg)
		}

		if err = server.awaitAuthorizationCode(30 * time.Second); err != nil {
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
	state, err := newState(authURL, false)
	if err != nil {
		return "", "", err
	}
	state.SSOProviderRefreshToken = refreshToken

	if err := state.refreshOAuthToken(); err != nil {
		return "", "", err
	}

	return state.SSOProviderIDToken, state.SSOProviderRefreshToken, nil
}
