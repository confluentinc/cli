package sso

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

var (
	ssoProviderCallbackEndpoint = "/cli_callback"
	ssoProviderCallbackLocalURL = fmt.Sprintf("http://127.0.0.1:%d", port) + ssoProviderCallbackEndpoint

	ssoConfigs = map[string]ssoConfig{
		"devel": {
			ssoProviderDomain:     "login.confluent-dev.io/oauth",
			ssoProviderIdentifier: "https://confluent-dev.auth0.com/api/v2/",
			ssoProviderScope:      "email%20openid%20offline_access",
		},
		"devel-us-gov": {
			ssoProviderDomain: "confluent-devel-us-gov.oktapreview.com/oauth2/v1",
			ssoProviderScope:  "openid+profile+email+offline_access",
		},
		"infra-us-gov": {
			ssoProviderDomain: "confluent-infra-us-gov.okta.com/oauth2/v1",
			ssoProviderScope:  "openid+profile+email+offline_access",
		},
		"prod": {
			ssoProviderDomain:     "login.confluent.io/oauth",
			ssoProviderIdentifier: "https://confluent.auth0.com/api/v2/",
			ssoProviderScope:      "email%20openid%20offline_access",
		},
		"prod-us-gov": {
			ssoProviderDomain: "confluent-prod-us-gov.okta.com/oauth2/v1",
			ssoProviderScope:  "openid+profile+email+offline_access",
		},
		"stag": {
			ssoProviderDomain:     "login-stag.confluent-dev.io/oauth",
			ssoProviderIdentifier: "https://confluent-stag.auth0.com/api/v2/",
			ssoProviderScope:      "email%20openid%20offline_access",
		},
		"test": {
			ssoProviderDomain:     "test.com/oauth",
			ssoProviderIdentifier: "https://test.auth0.com/api/v2/",
			ssoProviderScope:      "email%20openid%20offline_access",
		},
	}
)

type ssoConfig struct {
	ssoProviderDomain     string
	ssoProviderIdentifier string
	ssoProviderScope      string
}

/*
authState holds auth related codes and hashes and urls that are used by both the browser based SSO auth
and non browser based auth mechanisms
*/
type authState struct {
	CodeVerifier                  string
	CodeChallenge                 string
	SSOProviderAuthenticationCode string
	SSOProviderIDToken            string
	SSOProviderRefreshToken       string
	SSOProviderState              string
	SSOProviderHost               string
	SSOProviderClientID           string
	SSOProviderCallbackUrl        string
	SSOProviderIdentifier         string
	SSOProviderScope              string
}

// InitState generates various auth0 related codes and hashes
// and tweaks certain variables for internal development and testing of the CLIs
// auth0 server / SSO integration.
func newState(authURL string, noBrowser bool) (*authState, error) {
	authURL = strings.TrimSuffix(authURL, "/")

	if authURL == "" {
		authURL = "https://confluent.cloud"
	}

	var env string
	if authURL == "https://confluent.cloud" {
		env = "prod"
	} else if authURL == "https://devel.cpdev.cloud" {
		env = "devel"
	} else if authURL == "https://stag.cpdev.cloud" {
		env = "stag"
	} else if authURL == "https://confluentgov.com" {
		env = "prod-us-gov"
	} else if authURL == "https://infra.confluentgov-internal.com" {
		env = "infra-us-gov"
	} else if authURL == "https://devel.confluentgov-internal.com" {
		env = "devel-us-gov"
	} else if authURL == testserver.TestCloudUrl.String() {
		env = "test"
	} else {
		return nil, fmt.Errorf("unrecognized auth url: %s", authURL)
	}

	state := &authState{}
	state.SSOProviderCallbackUrl = authURL + ssoProviderCallbackEndpoint
	state.SSOProviderHost = "https://" + ssoConfigs[env].ssoProviderDomain
	state.SSOProviderClientID = GetAuth0CCloudClientIdFromBaseUrl(authURL)
	state.SSOProviderIdentifier = ssoConfigs[env].ssoProviderIdentifier
	state.SSOProviderScope = ssoConfigs[env].ssoProviderScope

	if !noBrowser {
		// if we're not using the no browser flow, the callback will always be localhost regardless of environment
		state.SSOProviderCallbackUrl = ssoProviderCallbackLocalURL
	}

	if err := state.generateCodes(); err != nil {
		return nil, err
	}

	return state, nil
}

// generateCodes makes code variables for use with the Authorization Code + PKCE flow
func (s *authState) generateCodes() error {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("unable to generate random bytes for SSO provider state: %w", err)
	}

	s.SSOProviderState = base64.RawURLEncoding.EncodeToString(randomBytes)

	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("unable to generate random bytes for code verifier: %w", err)
	}

	s.CodeVerifier = base64.RawURLEncoding.EncodeToString(randomBytes)

	hasher := sha256.New()
	if _, err := hasher.Write([]byte(s.CodeVerifier)); err != nil {
		return fmt.Errorf("unable to compute hash for code challenge: %w", err)
	}
	s.CodeChallenge = base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	return nil
}

// getOAuthToken exchanges the obtained authorization code for an auth0/ID token from the SSO provider
func (s *authState) getOAuthToken() error {
	payload := strings.NewReader("grant_type=authorization_code" +
		"&client_id=" + s.SSOProviderClientID +
		"&code_verifier=" + s.CodeVerifier +
		"&code=" + s.SSOProviderAuthenticationCode +
		"&redirect_uri=" + s.SSOProviderCallbackUrl)

	data, err := s.getOAuthTokenResponse(payload)
	if err != nil {
		return err
	}

	return s.saveOAuthTokenResponse(data)
}

// refreshOAuthToken exchanges the refresh token for an auth0/ID token from the SSO provider
func (s *authState) refreshOAuthToken() error {
	payload := strings.NewReader("grant_type=refresh_token" +
		"&client_id=" + s.SSOProviderClientID +
		"&refresh_token=" + s.SSOProviderRefreshToken +
		"&redirect_uri=" + s.SSOProviderCallbackUrl)

	data, err := s.getOAuthTokenResponse(payload)
	if err != nil {
		return err
	}

	return s.saveOAuthTokenResponse(data)
}

func (s *authState) saveOAuthTokenResponse(data map[string]any) error {
	if token, ok := data["id_token"]; ok {
		s.SSOProviderIDToken = token.(string)
	} else {
		return fmt.Errorf(errors.FmtMissingOauthFieldErrorMsg, "id_token")
	}

	if token, ok := data["refresh_token"]; ok {
		s.SSOProviderRefreshToken = token.(string)
	} else {
		return fmt.Errorf(errors.FmtMissingOauthFieldErrorMsg, "refresh_token")
	}

	return nil
}

func (s *authState) getOAuthTokenResponse(payload *strings.Reader) (map[string]any, error) {
	url := s.SSOProviderHost + "/token"
	log.CliLogger.Debugf("OAuth token request URL: %s", url)
	log.CliLogger.Debug("OAuth token request payload: ", payload)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to construct oauth token re, errquest: %w", err)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth token: %w", err)
	}
	defer res.Body.Close()
	errorResponseBody, _ := io.ReadAll(res.Body)
	var data map[string]any
	if err := json.Unmarshal(errorResponseBody, &data); err != nil {
		log.CliLogger.Debugf("Failed oauth token response body: %s", errorResponseBody)
		return nil, fmt.Errorf("failed to unmarshal response body in oauth token request: %w", err)
	}
	return data, nil
}

func (s *authState) getAuthorizationCodeUrl(ssoProviderConnectionName string, isOkta bool) string {
	url := s.SSOProviderHost + "/authorize?" +
		"response_type=code" +
		"&code_challenge=" + s.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + s.SSOProviderClientID +
		"&redirect_uri=" + s.SSOProviderCallbackUrl +
		"&scope=" + s.SSOProviderScope +
		"&state=" + s.SSOProviderState

	if s.SSOProviderIdentifier != "" {
		url += "&audience=" + s.SSOProviderIdentifier
	}

	if isOkta {
		url += "&idp=" + ssoProviderConnectionName
	} else {
		url += "&connection=" + ssoProviderConnectionName
	}

	return url
}
