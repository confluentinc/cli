package sso

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	ssoProviderDomain                = "login.confluent.io"
	ssoProviderDomainDevel           = "login.confluent-dev.io"
	ssoProviderClientID              = "hPbGZM8G55HSaUsaaieiiAprnJaEc9rH"
	ssoProviderClientIDDevel         = "XKlqgOEo39iyonTl3Yv3IHWIXGKDP3fA"
	ssoProviderCallbackEndpoint      = "/cli_callback"
	ssoProviderCallbackLocalURL      = "http://127.0.0.1:26635" + ssoProviderCallbackEndpoint
	ssoProviderCallbackCCloudURL     = "https://confluent.cloud" + ssoProviderCallbackEndpoint // used in the --no-browser sso flow
	ssoProviderCallbackCCloudDevURL  = "https://devel.cpdev.cloud" + ssoProviderCallbackEndpoint
	ssoProviderCallbackCCloudStagURL = "https://stag.cpdev.cloud" + ssoProviderCallbackEndpoint
	ssoProviderIdentifier            = "https://confluent.auth0.com/api/v2/"
	ssoProviderIdentifierDevel       = "https://confluent-dev.auth0.com/api/v2/"
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
// and tweaks certain variables for internal development and testing of the CLIs
// auth0 server / SSO integration.
func newState(authURL string, noBrowser bool) (*authState, error) {
	env := "prod"
	if strings.Contains(authURL, "priv.cpdev.cloud") {
		env = "cpd"
	}
	if strings.Contains(authURL, "devel.cpdev.cloud") {
		env = "devel"
	}
	if strings.Contains(authURL, "stag.cpdev.cloud") {
		env = "stag"
	}

	state := &authState{}
	switch env {
	case "cpd":
		state.SSOProviderCallbackUrl = authURL + ssoProviderCallbackEndpoint // callback to the cpd cluster url that was passed in
		state.SSOProviderHost = "https://" + ssoProviderDomainDevel          // only one Auth0 account for cpd, dev and stag
		state.SSOProviderClientID = ssoProviderClientIDDevel
		state.SSOProviderIdentifier = ssoProviderIdentifierDevel
	case "devel":
		state.SSOProviderCallbackUrl = ssoProviderCallbackCCloudDevURL
		state.SSOProviderHost = "https://" + ssoProviderDomainDevel
		state.SSOProviderClientID = ssoProviderClientIDDevel
		state.SSOProviderIdentifier = ssoProviderIdentifierDevel
	case "stag":
		state.SSOProviderCallbackUrl = ssoProviderCallbackCCloudStagURL
		state.SSOProviderHost = "https://" + ssoProviderDomainDevel
		state.SSOProviderClientID = ssoProviderClientIDDevel
		state.SSOProviderIdentifier = ssoProviderIdentifierDevel
	case "prod":
		state.SSOProviderCallbackUrl = ssoProviderCallbackCCloudURL
		state.SSOProviderHost = "https://" + ssoProviderDomain
		state.SSOProviderClientID = ssoProviderClientID
		state.SSOProviderIdentifier = ssoProviderIdentifier
	}

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
