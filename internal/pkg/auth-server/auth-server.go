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

	server = &http.Server{}
	wg     = &sync.WaitGroup{}
	bgErr  = *new(error)
)

/*
AuthServer is an HTTP server embedded in the CLI to serve callback requests for Auth0 SSO logins.
The server runs in a goroutine / in the background.
*/
type AuthServer struct {
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
		return errors.Wrap(err, "Unable to generate random bytes for code verifier")
	}

	s.CodeVerifier = base64.RawURLEncoding.EncodeToString(randomBytes)

	hasher := sha256.New()
	_, err = hasher.Write([]byte(s.CodeVerifier))
	if err != nil {
		return errors.Wrap(err, "Unable to compute hash for code challenge")
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
		return errors.Wrap(err, "Unable to start HTTP server")
	}

	wg.Add(1)
	go server.Serve(listener)

	return nil
}

// GetAuthorizationCode takes the code verifier/challenge and gets an authorization code from Auth0
func (s *AuthServer) GetAuthorizationCode(auth0ConnectionName string) error {
	url := "https://" + Auth0Domain + "/authorize?" +
		"response_type=code&" +
		"code_challenge=" + s.CodeChallenge + "&" +
		"code_challenge_method=S256&" +
		"client_id=" + Auth0ClientID + "&" +
		"redirect_uri=" + Auth0CallbackURL + "&" +
		"scope=openid&" +
		"audience=" + Auth0Identifier + "&" +
		"state=" + Auth0State
	if auth0ConnectionName != "" {
		url += "&connection=" + auth0ConnectionName
	}

	browser.OpenURL(url)

	// Wait until flow is finished / callback is called (or timeout...)
	go func() {
		time.Sleep(30 * time.Second)
		bgErr = errors.New("Timed out while waiting for browser authentication to occur.  Please try logging in again.")
		server.Close()
		wg.Done()
	}()
	wg.Wait()

	defer server.Shutdown(context.Background())

	return bgErr
}

// CallbackHandler serves the route /callback
func (s *AuthServer) CallbackHandler(rw http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(rw, "CLI Authentication successful!  You may close this tab/window.")

	codes, ok := request.URL.Query()["code"]
	if ok {
		s.Auth0AuthenticationCode = codes[0]
	} else {
		bgErr = errors.New("Authentication callback URL did not contain code parameter in query string.  Login will fail.")
	}

	wg.Done()
}

// GetAuth0Token exchanges the obtained Auth0 authorization code for an auth/ID token from Auth0
func (s *AuthServer) GetAuth0Token() error {
	url := "https://" + Auth0Domain + "/oauth/token"
	payload := strings.NewReader("grant_type=authorization_code&" +
		"client_id=" + Auth0ClientID + "&" +
		"code_verifier=" + s.CodeVerifier + "&" +
		"code=" + s.Auth0AuthenticationCode + "&" +
		"redirect_uri=" + Auth0CallbackURL)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	responseBody, _ := ioutil.ReadAll(res.Body)

	var data map[string]interface{}
	err := json.Unmarshal([]byte(responseBody), &data)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal response body in GetAuth0Token request")
	}

	token, ok := data["id_token"]
	if ok {
		s.Auth0IDToken = token.(string)
	} else {
		return errors.New("GetAuth0Token response body did not contain id_token field")
	}

	return nil
}
