package sso

import (
	"bufio"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/pkg/browser"
	"os"
	"strings"
	"time"
)

func Login(config *config.Config, auth0ConnectionName string) (idToken string, err error) {
	state, err := newState(config)
	if err != nil {
		return "", err
	}

	if config.NoBrowser {
		// no browser flag does not need to launch the server
		// it prints the url and has the user copy this into their browser instead
		url := state.getAuthorizationCodeUrl(auth0ConnectionName)
		fmt.Println("Navigate to the following link in your browser to authenticate:")
		fmt.Printf("%s", url)
		fmt.Println()
		fmt.Println()
		fmt.Println("After authenticating in your browser, paste the code here:")

		// wait for the user to paste the code
		// the code should come in the format {state}/{auth0_auth_code}
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()
		split := strings.SplitAfterN(input, "/", 2)
		if len(split) < 2 {
			return "", errors.New("Pasted input had invalid format")
		}
		auth0State := strings.Replace(split[0], "/", "", 1)
		if !(auth0State == state.SSOProviderState) {
			return "", errors.New("authentication code either did not contain a state parameter or the state parameter was invalid; login will fail")
		}

		state.SSOProviderAuthenticationCode = split[1]
	} else {
		// we need to start a background HTTP server to support the authorization code flow with PKCE
		// described at https://auth0.com/docs/flows/guides/auth-code-pkce/call-api-auth-code-pkce
		server := newServer(state)
		err = server.startServer()
		if err != nil {
			return "", err
		}

		// Get authorization code for making subsequent token request
		err := browser.OpenURL(state.getAuthorizationCodeUrl(auth0ConnectionName))
		if err != nil {
			return "", errors.Wrap(err, "unable to open web browser for authorization")
		}

		err = server.awaitAuthorizationCode(30 * time.Second)
		if err != nil {
			return "", err
		}
	}

	// Exchange authorization code for OAuth token from SSO provider
	err = state.getOAuthToken()
	if err != nil {
		return "", err
	}

	return state.SSOProviderIDToken, nil
}
