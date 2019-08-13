package auth_server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/browser"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	SSOProviderDomain      = "login.confluent.io"
	SSOProviderClientID    = "Z1Pnpscwhl5WgcEdhP3ec2O307D6HfKg"
	SSOProviderCallbackURL = "http://127.0.0.1:26635/callback"
	SSOProviderIdentifier  = "https://confluent.auth0.com/api/v2/"
	SSOProviderState       = "ConfluentCLI"
)

/*
AuthServer is an HTTP server embedded in the CLI to serve callback requests for Auth0 SSO logins.
The server runs in a goroutine / in the background.
*/
type AuthServer struct {
	server                        *http.Server
	wg                            *sync.WaitGroup
	bgErr                         error
	CodeVerifier                  string
	CodeChallenge                 string
	SSOProviderAuthenticationCode string
	SSOProviderIDToken            string
}

// GenerateCodes makes code variables for use with Auth0
func (s *AuthServer) GenerateCodes() error {
	randomBytes := make([]byte, 32)

	_, err := rand.Read(randomBytes)
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

// initializeInternalVariables is an internal function used to tweak
// certain variables for internal development and testing of the CLI's
// auth server / SSO integration.
func (s *AuthServer) initializeInternalVariables(authURL string) {
	// Auth configs change for Confluent internal development usage...
	env := "prod"
	if strings.Contains(authURL, "devel.cpdev.cloud") {
		env = "devel"
	}
	if strings.Contains(authURL, "stag.cpdev.cloud") {
		env = "stag"
	}

	if env == "devel" || env == "stag" {
		SSOProviderDomain = "login.confluent-dev.io"
		SSOProviderClientID = "yKJfeHs2o7PdEhxDmPIqflWNE6cPieqm"
		SSOProviderIdentifier = "https://confluent-dev.auth0.com/api/v2/"
	}
}

// Start begins the server including attempting to bind the desired TCP port
func (s *AuthServer) Start(authURL string) error {
	s.initializeInternalVariables(authURL)

	err := s.GenerateCodes()
	if err != nil {
		return err
	}

	http.HandleFunc("/callback", s.CallbackHandler)

	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 26635}) // confl

	if err != nil {
		return errors.Wrap(err, "unable to start HTTP server")
	}

	s.wg = &sync.WaitGroup{}
	s.server = &http.Server{}

	s.wg.Add(1)
	go func() {
		serverErr := s.server.Serve(listener)
		// Serve returns ErrServerClosed when the server is gracefully, intentionally closed:
		// https://go.googlesource.com/go/+/master/src/net/http/server.go#2854
		// So don't surface that error to the user.
		if serverErr != nil && serverErr.Error() != "http: Server closed" {
			fmt.Fprintf(os.Stderr, "CLI HTTP auth server encountered error while running: %s\n", serverErr.Error())
		}
	}()

	return nil
}

// GetAuthorizationCode takes the code verifier/challenge and gets an authorization code from Auth0
func (s *AuthServer) GetAuthorizationCode(auth0ConnectionName string) error {
	url := "https://" + SSOProviderDomain + "/authorize?" +
		"response_type=code" +
		"&code_challenge=" + s.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + SSOProviderClientID +
		"&redirect_uri=" + SSOProviderCallbackURL +
		"&scope=email%20openid" +
		"&audience=" + SSOProviderIdentifier +
		"&state=" + SSOProviderState
	if auth0ConnectionName != "" {
		url += "&connection=" + auth0ConnectionName
	}

	err := browser.OpenURL(url)
	if err != nil {
		return errors.Wrap(err, "unable to open web browser for authorization")
	}

	// Wait until flow is finished / callback is called (or timeout...)
	go func() {
		time.Sleep(30 * time.Second)
		s.bgErr = errors.New("timed out while waiting for browser authentication to occur; please try logging in again")
		s.server.Close()
		s.wg.Done()
	}()
	s.wg.Wait()

	defer func() {
		serverErr := s.server.Shutdown(context.Background())
		if serverErr != nil {
			fmt.Fprintf(os.Stderr, "CLI HTTP auth server encountered error while shutting down: %s\n", serverErr.Error())
		}
	}()

	return s.bgErr
}

// CallbackHandler serves the route /callback
func (s *AuthServer) CallbackHandler(rw http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(rw, "CLI Authentication successful!  You may close this tab/window.")

	codes, ok := request.URL.Query()["code"]
	if ok {
		s.SSOProviderAuthenticationCode = codes[0]
	} else {
		s.bgErr = errors.New("authentication callback URL did not contain code parameter in query string, login will fail")
	}

	s.wg.Done()
}

// GetAuth0Token exchanges the obtained Auth0 authorization code for an auth/ID token from Auth0
func (s *AuthServer) GetAuth0Token() error {
	url := "https://" + SSOProviderDomain + "/oauth/token"
	payload := strings.NewReader("grant_type=authorization_code" +
		"&client_id=" + SSOProviderClientID +
		"&code_verifier=" + s.CodeVerifier +
		"&code=" + s.SSOProviderAuthenticationCode +
		"&redirect_uri=" + SSOProviderCallbackURL)
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
