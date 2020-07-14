package sso

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func Login(authURL string, noBrowser bool, auth0ConnectionName string, logger *log.Logger) (idToken string, refreshToken string, err error) {
	state, err := newState(authURL, noBrowser, logger)
	if err != nil {
		return "", "", err
	}

	if noBrowser {
		// no browser flag does not need to launch the server
		// it prints the url and has the user copy this into their browser instead
		url := state.getAuthorizationCodeUrl(auth0ConnectionName)
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
		err = server.startServer()
		if err != nil {
			return "", "", err
		}

		// Get authorization code for making subsequent token request
		err := browser.OpenURL(state.getAuthorizationCodeUrl(auth0ConnectionName))
		if err != nil {
			return "", "", errors.Wrap(err, errors.OpenWebBrowserErrorMsg)
		}

		err = server.awaitAuthorizationCode(30 * time.Second)
		if err != nil {
			return "", "", err
		}
	}

	// Exchange authorization code for OAuth token from SSO provider
	err = state.getOAuthToken()
	if err != nil {
		return "", "", err
	}

	return state.SSOProviderIDToken, state.SSOProviderRefreshToken, nil
}

func GetNewIDTokenFromRefreshToken(authURL string, refreshToken string, logger *log.Logger) (idToken string, err error) {
	state, err := newState(authURL, false, logger)
	if err != nil {
		return "", err
	}
	state.SSOProviderRefreshToken = refreshToken
	err = state.refreshOAuthToken()
	if err != nil {
		return "", err
	}
	return state.SSOProviderIDToken, err
}
