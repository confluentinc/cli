package mfa

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/cli/v4/pkg/auth/sso"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
	"io"
	"net/http"
	"strings"
)

type authState struct {
	CodeVerifier                  string
	CodeChallenge                 string
	MFAProviderAuthenticationCode string
	MFAProviderIDToken            string
	MFAProviderRefreshToken       string
	MFAProviderState              string
	email                         string
	MFAProviderCallbackUrl        string
	MFAProviderHost               string
	MFAProviderClientID           string
	MFAProviderScope              string
}

func newState(authUrl, email string) (*authState, error) {
	if authUrl == "" {
		authUrl = "https://confluent.cloud"
	}

	env := sso.GetCCloudEnvFromBaseUrl(authUrl)

	state := &authState{
		MFAProviderHost:        "https://" + sso.SsoConfigs[env].SsoProviderDomain,
		MFAProviderCallbackUrl: sso.SsoProviderCallbackLocalURL,
		MFAProviderClientID:    sso.GetAuth0CCloudClientIdFromBaseUrl(authUrl),
		email:                  email,
		MFAProviderScope:       sso.SsoConfigs[env].SsoProviderScope,
	}
	if err := state.generateCodes(); err != nil {
		return nil, err
	}
	return state, nil
}

func (s *authState) generateCodes() error {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("unable to generate random bytes for SSO provider state: %w", err)
	}

	s.MFAProviderState = base64.RawURLEncoding.EncodeToString(randomBytes)

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

func (s *authState) getAuthorizationCodeUrl() string {
	url := s.MFAProviderHost + "/authorize?challenge_mfa=true" +
		"&response_type=code" +
		"&email=" + s.email +
		"&from_cli=true" +
		"&code_challenge=" + s.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + s.MFAProviderClientID +
		"&redirect_uri=" + s.MFAProviderCallbackUrl +
		"&scope=" + s.MFAProviderScope +
		"&state=" + s.MFAProviderState

	return url
}

func (s *authState) saveOAuthTokenResponse(data map[string]any) error {
	if token, ok := data["id_token"]; ok {
		s.MFAProviderIDToken = token.(string)
	} else {
		return fmt.Errorf("Incorrect token added. Please try to login again.")
	}

	if token, ok := data["refresh_token"]; ok {
		s.MFAProviderRefreshToken = token.(string)
	} else {
		return fmt.Errorf(errors.FmtMissingOauthFieldErrorMsg, "refresh_token")
	}

	return nil
}

func (s *authState) getOAuthToken() error {
	payload := strings.NewReader("grant_type=authorization_code" +
		"&client_id=" + s.MFAProviderClientID +
		"&code_verifier=" + s.CodeVerifier +
		"&code=" + s.MFAProviderAuthenticationCode +
		"&redirect_uri=" + s.MFAProviderCallbackUrl)

	data, err := s.getOAuthTokenResponse(payload)
	if err != nil {
		return err
	}

	return s.saveOAuthTokenResponse(data)
}

func (s *authState) getOAuthTokenResponse(payload *strings.Reader) (map[string]any, error) {
	url := s.MFAProviderHost + "/token"
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