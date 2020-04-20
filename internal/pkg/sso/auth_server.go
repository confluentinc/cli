package sso

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cc-utils/local"
)

/*
authServer is an HTTP server embedded in the CLI to serve callback requests for SSO logins.
The server runs in a goroutine / in the background.
*/
type authServer struct {
	server *http.Server
	wg     *sync.WaitGroup
	bgErr  error
	State  *authState
}

func newServer(state *authState) *authServer {
	return &authServer{State: state}
}

// Start begins the server including attempting to bind the desired TCP port
func (s *authServer) startServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/cli_callback", s.callbackHandler)

	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 26635}) // confl

	if err != nil {
		return errors.Wrap(err, "unable to start HTTP server")
	}

	s.wg = &sync.WaitGroup{}
	s.server = &http.Server{Handler: mux}

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

// GetAuthorizationCode takes the code verifier/challenge and gets an authorization code from the SSO provider
func (s *authServer) awaitAuthorizationCode(timeout time.Duration) error {
	// Wait until flow is finished / callback is called (or timeout...)
	go func() {
		time.Sleep(timeout)
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
func (s *authServer) callbackHandler(rw http.ResponseWriter, request *http.Request) {
	states, ok := request.URL.Query()["state"]
	if !(ok && states[0] == s.State.SSOProviderState) {
		s.bgErr = errors.New("authentication callback URL either did not contain a state parameter in query string, or the state parameter was invalid; login will fail")
	}

	rawCallbackFile, err := local.Asset("assets/sso_callback.html")
	if err != nil {
		s.bgErr = errors.New("could not read callback page template")
		fmt.Fprintf(rw, "could not read callback page template, see CLI terminal for more details")
	}
	fmt.Fprintf(rw, string(rawCallbackFile))

	codes, ok := request.URL.Query()["code"]
	if ok {
		s.State.SSOProviderAuthenticationCode = codes[0]
	} else {
		s.bgErr = errors.New("authentication callback URL did not contain code parameter in query string; login will fail")
	}

	s.wg.Done()
}
