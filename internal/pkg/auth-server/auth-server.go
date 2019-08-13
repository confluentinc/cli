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
	Auth0Domain      = "login.confluent.io"
	Auth0ClientID    = "Z1Pnpscwhl5WgcEdhP3ec2O307D6HfKg"
	Auth0CallbackURL = "http://127.0.0.1:26635/callback"
	Auth0Identifier  = "https://confluent.auth0.com/api/v2/"
	Auth0State       = "ConfluentCLI"

	bgErr = *new(error)
)

/*
AuthServer is an HTTP server embedded in the CLI to serve callback requests for Auth0 SSO logins.
The server runs in a goroutine / in the background.
*/
type AuthServer struct {
	server                  *http.Server
	wg                      *sync.WaitGroup
	CodeVerifier            string
	CodeChallenge           string
	Auth0AuthenticationCode string
	Auth0IDToken            string
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

// Start begins the server including attempting to bind the desired TCP port
func (s *AuthServer) Start(env string) error {
	if env == "devel" || env == "stag" {
		Auth0Domain = "login.confluent-dev.io"
		Auth0ClientID = "yKJfeHs2o7PdEhxDmPIqflWNE6cPieqm"
		Auth0Identifier = "https://confluent-dev.auth0.com/api/v2/"
	}

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
	url := "https://" + Auth0Domain + "/authorize?" +
		"response_type=code" +
		"&code_challenge=" + s.CodeChallenge +
		"&code_challenge_method=S256" +
		"&client_id=" + Auth0ClientID +
		"&redirect_uri=" + Auth0CallbackURL +
		"&scope=email%20openid" +
		"&audience=" + Auth0Identifier +
		"&state=" + Auth0State
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
		bgErr = errors.New("timed out while waiting for browser authentication to occur; please try logging in again")
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

	return bgErr
}

// CallbackHandler serves the route /callback
func (s *AuthServer) CallbackHandler(rw http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(rw, "CLI Authentication successful!  You may close this tab/window.")

	codes, ok := request.URL.Query()["code"]
	if ok {
		s.Auth0AuthenticationCode = codes[0]
	} else {
		bgErr = errors.New("authentication callback URL did not contain code parameter in query string, login will fail")
	}

	s.wg.Done()
}

// GetAuth0Token exchanges the obtained Auth0 authorization code for an auth/ID token from Auth0
func (s *AuthServer) GetAuth0Token() error {
	url := "https://" + Auth0Domain + "/oauth/token"
	payload := strings.NewReader("grant_type=authorization_code" +
		"&client_id=" + Auth0ClientID +
		"&code_verifier=" + s.CodeVerifier +
		"&code=" + s.Auth0AuthenticationCode +
		"&redirect_uri=" + Auth0CallbackURL)
	req, _ := http.NewRequest("POST", url, payload)
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
		s.Auth0IDToken = token.(string)
	} else {
		return errors.New("oauth token response body did not contain id_token field")
	}

	return nil
}
