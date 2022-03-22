package sso

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

var (
	ssoProviderCallbackEndpoint = "/cli_callback"
	ssoProviderCallbackLocalURL = fmt.Sprintf("http://127.0.0.1:%d", port) + ssoProviderCallbackEndpoint

	ssoConfigs = map[string]ssoConfig{
		"cpd": {
			ssoProviderDomain:     "login-cpd.confluent-dev.io",
			ssoProviderIdentifier: "https://confluent-cpd.auth0.com/api/v2/",
		},
		"devel": {
			ssoProviderDomain:     "login.confluent-dev.io",
			ssoProviderIdentifier: "https://confluent-dev.auth0.com/api/v2/",
		},
		"stag": {
			ssoProviderDomain:     "login-stag.confluent-dev.io",
			ssoProviderIdentifier: "https://confluent-stag.auth0.com/api/v2/",
		},
		"prod": {
			ssoProviderDomain:     "login.confluent.io",
			ssoProviderIdentifier: "https://confluent.auth0.com/api/v2/",
		},
	}
)

type ssoConfig struct {
	ssoProviderDomain     string
	ssoProviderIdentifier string
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
}

// InitState generates various auth0 related codes and hashes
// and tweaks certain variables for internal development and testing of the CLIs
// auth0 server / SSO integration.
func newState(authURL string, noBrowser bool) (*authState, error) {
	if authURL == "" {
		authURL = "https://confluent.cloud"
	}

	var env string
	if authURL == "https://confluent.cloud" {
		env = "prod"
	} else if strings.HasSuffix(authURL, "priv.cpdev.cloud") {
		env = "cpd"
	} else if authURL == "https://devel.cpdev.cloud" {
		env = "devel"
	} else if authURL == "https://stag.cpdev.cloud" {
		env = "stag"
	} else {
		return nil, fmt.Errorf("unrecognized auth url: %s", authURL)
	}

	state := &authState{}
	state.SSOProviderCallbackUrl = authURL + ssoProviderCallbackEndpoint
	state.SSOProviderHost = "https://" + ssoConfigs[env].ssoProviderDomain
	state.SSOProviderClientID = GetAuth0CCloudClientIdFromBaseUrl(authURL)
	state.SSOProviderIdentifier = ssoConfigs[env].ssoProviderIdentifier

	if !noBrowser {
		// if we're not using the no browser flow, the callback will always be localhost regardless of environment
		state.SSOProviderCallbackUrl = ssoProviderCallbackLocalURL
	}

	err := state.generateCodes()
	if err != nil {
		return nil, err
	}

	return state, nil
}

// generateCodes makes code variables for use with the Authorization Code + PKCE flow
func (s *authState) generateCodes() error {
	randomBytes := make([]byte, 32)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return errors.Wrap(err, errors.GenerateRandomSSOProviderErrorMsg)
	}

	s.SSOProviderState = base64.RawURLEncoding.EncodeToString(randomBytes)

	_, err = rand.Read(randomBytes)
	if err != nil {
		return errors.Wrap(err, errors.GenerateRandomCodeVerifierErrorMsg)
	}

	s.CodeVerifier = base64.RawURLEncoding.EncodeToString(randomBytes)

	hasher := sha256.New()
	_, err = hasher.Write([]byte(s.CodeVerifier))
	if err != nil {
		return errors.Wrap(err, errors.ComputeHashErrorMsg)
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

func (s *authState) saveOAuthTokenResponse(data map[string]interface{}) error {
	if token, ok := data["id_token"]; ok {
		s.SSOProviderIDToken = token.(string)
	} else {
		return errors.Errorf(errors.FmtMissingOAuthFieldErrorMsg, "id_token")
	}

	if token, ok := data["refresh_token"]; ok {
		s.SSOProviderRefreshToken = token.(string)
	} else {
		return errors.Errorf(errors.FmtMissingOAuthFieldErrorMsg, "refresh_token")
	}

	return nil
}

func (s *authState) getOAuthTokenResponse(payload *strings.Reader) (map[string]interface{}, error) {
	url := s.SSOProviderHost + "/oauth/token"
	log.CliLogger.Debugf("Oauth token request URL: %s", url)
	log.CliLogger.Debug("Oauth token request payload: ", payload)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, errors.Wrap(err, errors.ConstructOAuthRequestErrorMsg)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get oauth token")
	}
	defer res.Body.Close()
	responseBody, _ := ioutil.ReadAll(res.Body)
	var data map[string]interface{}
	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		log.CliLogger.Debugf("Failed oauth token response body: %s", responseBody)
		return nil, errors.Wrap(err, errors.UnmarshalOAuthTokenErrorMsg)
	}
	return data, nil
}

func (s *authState) getAuthorizationCodeUrl(ssoProviderConnectionName string) string {
	url := s.SSOProviderHost + "/authorize?" +
		"response_type=code" +
		"&code_challenge=" + s.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + s.SSOProviderClientID +
		"&redirect_uri=" + s.SSOProviderCallbackUrl +
		"&scope=email%20openid%20offline_access" +
		"&audience=" + s.SSOProviderIdentifier +
		"&state=" + s.SSOProviderState
	if ssoProviderConnectionName != "" {
		url += "&connection=" + ssoProviderConnectionName
	}
	return url
}
