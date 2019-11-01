package sso

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	ssoProviderDomain               = "login.confluent.io"
	ssoProviderDomainDevel          = "login.confluent-dev.io"
	ssoProviderClientID             = "hPbGZM8G55HSaUsaaieiiAprnJaEc9rH"
	ssoProviderClientIDDevel        = "XKlqgOEo39iyonTl3Yv3IHWIXGKDP3fA"
	ssoProviderCallbackLocalURL     = "http://127.0.0.1:26635/cli_callback"
	ssoProviderCallbackCCloudURL    = "https://" + ssoProviderDomain + "/cli_callback" // used in the --no-browser sso flow
	ssoProviderCallbackCCloudDevURL = "https://" + ssoProviderDomainDevel + "/cli_callback"
	ssoProviderIdentifier           = "https://confluent.auth0.com/api/v2/"
	ssoProviderIdentifierDevel      = "https://confluent-dev.auth0.com/api/v2/"
)

/*
authState holds auth related codes and hashes and urls that are used by both the browser based SSO auth
and non browser based auth mechanisms
*/
type authState struct {
	CodeVerifier                  string
	CodeChallenge                 string
	SSOProviderAuthenticationCode string
	SSOProviderIDToken            string
	SSOProviderState              string
	SSOProviderHost               string
	SSOProviderClientID           string
	SSOProviderCallbackUrl        string
	SSOProviderIdentifier         string
}

// InitState generates various auth0 related codes and hashes
// and tweaks certain variables for internal development and testing of the CLI's
// auth0 server / SSO integration.
func newState(config *config.Config) (*authState, error) {
	env := "prod"
	if strings.Contains(config.AuthURL, "devel.cpdev.cloud") || strings.Contains(config.AuthURL, "priv.cpdev.cloud") {
		env = "devel"
	}
	if strings.Contains(config.AuthURL, "stag.cpdev.cloud") {
		env = "stag"
	}

	state := &authState{}
	err := state.generateCodes()
	if err != nil {
		return nil, err
	}

	if env == "devel" || env == "stag" {
		state.SSOProviderHost = "https://" + ssoProviderDomainDevel
		state.SSOProviderClientID = ssoProviderClientIDDevel
		state.SSOProviderIdentifier = ssoProviderIdentifierDevel

		if config.NoBrowser {
			state.SSOProviderCallbackUrl = ssoProviderCallbackCCloudDevURL
		} else {
			state.SSOProviderCallbackUrl = ssoProviderCallbackLocalURL
		}
	} else {
		state.SSOProviderHost = "https://" + ssoProviderDomain
		state.SSOProviderClientID = ssoProviderClientID
		state.SSOProviderIdentifier = ssoProviderIdentifier

		if config.NoBrowser {
			state.SSOProviderCallbackUrl = ssoProviderCallbackCCloudURL
		} else {
			state.SSOProviderCallbackUrl = ssoProviderCallbackLocalURL
		}
	}
	return state, nil
}

// generateCodes makes code variables for use with the Authorization Code + PKCE flow
func (s *authState) generateCodes() error {
	randomBytes := make([]byte, 32)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return errors.Wrap(err, "unable to generate random bytes for SSO provider state")
	}

	s.SSOProviderState = base64.RawURLEncoding.EncodeToString(randomBytes)

	_, err = rand.Read(randomBytes)
	if err != nil {
		return errors.Wrap(err, "unable to generate random bytes for code verifier")
	}

	s.CodeVerifier = base64.RawURLEncoding.EncodeToString(randomBytes)

	hasher := sha256.New()
	_, err = hasher.Write([]byte(s.CodeVerifier))
	if err != nil {
		return errors.Wrap(err, "unable to compute hash for code challenge")
	}
	s.CodeChallenge = base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	return nil
}

// GetOAuthToken exchanges the obtained authorization code for an auth0/ID token from the SSO provider
func (s *authState) getOAuthToken() error {
	url := s.SSOProviderHost + "/oauth/token"
	payload := strings.NewReader("grant_type=authorization_code" +
		"&client_id=" + s.SSOProviderClientID +
		"&code_verifier=" + s.CodeVerifier +
		"&code=" + s.SSOProviderAuthenticationCode +
		"&redirect_uri=" + s.SSOProviderCallbackUrl)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return errors.Wrap(err, "failed to construct oauth token request")
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to get oauth token")
	}

	defer res.Body.Close()
	responseBody, _ := ioutil.ReadAll(res.Body)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(responseBody), &data)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal response body in oauth token request")
	}

	token, ok := data["id_token"]
	if ok {
		s.SSOProviderIDToken = token.(string)
	} else {
		return errors.New("oauth token response body did not contain id_token field")
	}

	return nil
}

func (s *authState) getAuthorizationCodeUrl(ssoProviderConnectionName string) string {
	url := s.SSOProviderHost + "/authorize?" +
		"response_type=code" +
		"&code_challenge=" + s.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + s.SSOProviderClientID +
		"&redirect_uri=" + s.SSOProviderCallbackUrl +
		"&scope=email%20openid" +
		"&audience=" + s.SSOProviderIdentifier +
		"&state=" + s.SSOProviderState
	if ssoProviderConnectionName != "" {
		url += "&connection=" + ssoProviderConnectionName
	}

	return url
}
