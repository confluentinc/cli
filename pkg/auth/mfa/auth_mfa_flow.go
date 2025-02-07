package mfa

import (
	"bufio"
	"fmt"
	"github.com/confluentinc/cli/v4/pkg/output"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/confluentinc/cli/v4/pkg/auth/sso"
)

func Login(authURL, email, connectionName string, noBrowser bool) (string, string, error) {
	state, err := newState(authURL, email, noBrowser)
	if err != nil {
		return "", "", err
	}

	isOkta := sso.IsOkta(authURL)

	if noBrowser {
		// no browser flag does not need to launch the server
		// it prints the url and has the user copy this into their browser instead
		output.Println(false, "Navigate to the following link in your browser to authenticate:")
		output.Println(false, state.getAuthorizationCodeUrl(isOkta, connectionName))
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
		if !(auth0State == state.MFAProviderState) {
			return "", "", fmt.Errorf("authentication code either did not contain a state parameter or the state parameter was invalid; login will fail")
		}

		state.MFAProviderAuthenticationCode = split[1]
	} else {

		server := newServer(state)
		if err := server.startServer(); err != nil {
			return "", "", fmt.Errorf("unable to start HTTP server: %w", err)
		}
		// Get authorization code for making subsequent token request
		url := state.getAuthorizationCodeUrl(isOkta, connectionName)
		if err := browser.OpenURL(url); err != nil {
			return "", "", fmt.Errorf("unable to open web browser for authorization: %w", err)
		}
		if err := server.awaitAuthorizationCode(300 * time.Second); err != nil {
			return "", "", err
		}
	}

	// Exchange authorization code for OAuth token from provider
	if err := state.getOAuthToken(); err != nil {
		return "", "", err
	}

	return state.MFAProviderIDToken, state.MFAProviderRefreshToken, nil
}

func RefreshTokens(authURL, refreshToken, email string) (string, string, error) {
	state, err := newState(authURL, email, false)
	if err != nil {
		return "", "", err
	}
	state.MFAProviderRefreshToken = refreshToken

	if err := state.refreshOAuthToken(); err != nil {
		return "", "", err
	}

	return state.MFAProviderIDToken, state.MFAProviderRefreshToken, nil
}
func (s *authState) refreshOAuthToken() error {
	payload := strings.NewReader("grant_type=refresh_token" +
		"&client_id=" + s.MFAProviderClientID +
		"&refresh_token=" + s.MFAProviderRefreshToken +
		"&redirect_uri=" + s.MFAProviderCallbackUrl)

	data, err := s.getOAuthTokenResponse(payload)
	if err != nil {
		return err
	}

	return s.saveOAuthTokenResponse(data)
}
