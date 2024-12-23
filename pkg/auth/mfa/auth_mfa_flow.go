package mfa

import (
	"fmt"
	"github.com/pkg/browser"
	"strings"
	"time"
)

func Login(authURL, email string) (string, string, error) {
	state, err := newState(authURL, email)
	if err != nil {
		return "", "", err
	}

	server := newServer(state)
	if err := server.startServer(); err != nil {
		return "", "", fmt.Errorf("unable to start HTTP server: %w", err)
	}
	// Get authorization code for making subsequent token request
	url := state.getAuthorizationCodeUrl()
	if err := browser.OpenURL(url); err != nil {
		return "", "", fmt.Errorf("unable to open web browser for authorization[][][]: %w", err)
	}

	if err := server.awaitAuthorizationCode(300 * time.Second); err != nil {
		return "", "", err
	}

	// Exchange authorization code for OAuth token from provider
	if err := state.getOAuthToken(); err != nil {
		return "", "", err
	}

	return state.MFAProviderIDToken, state.MFAProviderRefreshToken, nil
}

func RefreshTokens(authURL, refreshToken, email string) (string, string, error) {
	state, err := newState(authURL, email)
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
