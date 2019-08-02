package auth_server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/pkg/browser"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	Auth0Domain      = "confluent-dev.auth0.com"
	Auth0ClientID    = "yKJfeHs2o7PdEhxDmPIqflWNE6cPieqm"
	Auth0CallbackURL = "http://127.0.0.1:26635/callback"
	Auth0Identifier  = "https://confluent-dev.auth0.com/api/v2/"
	Auth0State       = "ConfluentCLI"
)

/*
AuthServer is an HTTP server embedded in the CLI to serve callback requests for Auth0 SSO logins.
The server runs in a goroutine / in the background.
*/
type AuthServer struct {
	CodeVerifier  string
	CodeChallenge string
}

// GenerateCodes makes code variables for use with Auth0
func (s *AuthServer) GenerateCodes() error {
	randomBytes := make([]byte, 32)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return errors.Wrap(err, "Unable to generate random bytes for code verifier")
	}

	s.CodeVerifier = base64.URLEncoding.EncodeToString(randomBytes)

	hasher := sha256.New()
	_, err = hasher.Write([]byte(s.CodeVerifier))
	if err != nil {
		return errors.Wrap(err, "Unable to compute hash for code challenge")
	}
	s.CodeChallenge = base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return nil
}

// Start begins the server including attempting to bind the desired TCP port
func (s *AuthServer) Start() error {
	err := s.GenerateCodes()
	if err != nil {
		return err
	}

	http.HandleFunc("/callback", s.CallbackHandler)

	server := &http.Server{}
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 26635}) // confl

	if err != nil {
		return errors.Wrap(err, "Unable to start HTTP server")
	}

	go server.Serve(listener)

	return nil
}

// GetAuthorizationCode takes the code verifier/challenge and gets an authorization code from Auth0
func (s *AuthServer) GetAuthorizationCode(auth0ConnectionName string) error {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
		  return http.ErrUseLastResponse
	  } }
	url := "https://" + Auth0Domain + "/authorize?" +
		"response_type=token&" +
		"code_challenge=" + s.CodeChallenge + "&" +
		"code_challenge_method=S256&" +
		"client_id=" + Auth0ClientID + "&" +
		"redirect_uri=" + Auth0CallbackURL + "&" +
		"scope=profile%20email&" +
		"audience=" + Auth0Identifier + "&" +
		"state=" + Auth0State
	if auth0ConnectionName != "" {
		url += "&connection=" + auth0ConnectionName
	}
	fmt.Println(url)
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	redirectUrl, _ := resp.Location()
	fmt.Println(redirectUrl)

	browser.OpenURL(redirectUrl.String())

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(body))
	log.Println(resp.StatusCode)
	sl, e := resp.Location()
	if e != nil {
		log.Println(sl)
	}

	return nil
}

// CallbackHandler serves the route /callback
func (s *AuthServer) CallbackHandler(rw http.ResponseWriter, request *http.Request) {
	/* ... */
}
