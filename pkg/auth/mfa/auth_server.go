package mfa

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

//go:embed mfa_callback.html
var mfaCallbackHTML string

// Ideally we would randomize this value, but Auth0 requires that we hardcode a single port.
const port = 26635

/*
authServer is an HTTP server embedded in the CLI to serve callback requests for SSO logins.
The server runs in a goroutine / in the background.
*/
type authServer struct {
	server *http.Server
	wait   chan bool
	bgErr  error
	State  *authState
}

func newServer(state *authState) *authServer {
	return &authServer{
		wait:  make(chan bool),
		State: state,
	}
}

// Start begins the server including attempting to bind the desired TCP port
func (s *authServer) startServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/cli_callback", s.callbackHandler)

	// A process can intercept the Auth0 callback by listening to 0.0.0.0:<port>. Verify that this is not the case.
	lis, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		return err
	}
	_ = lis.Close()

	lis, err = net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
	if err != nil {
		return err
	}

	s.server = &http.Server{Handler: mux}

	go func() {
		serverErr := s.server.Serve(lis)
		// Serve returns ErrServerClosed when the server is gracefully, intentionally closed:
		// https://go.googlesource.com/go/+/master/src/net/http/server.go#2854
		// So don't surface that error to the user.
		if serverErr != nil && serverErr.Error() != "http: Server closed" {
			output.ErrPrintf(false, "CLI HTTP auth server encountered error while running: %v\n", serverErr.Error())
		}
	}()

	return nil
}

// GetAuthorizationCode takes the code verifier/challenge and gets an authorization code from the MFA provider
func (s *authServer) awaitAuthorizationCode(timeout time.Duration) error {
	// Wait until flow is finished / callback is called (or timeout...)
	go func() {
		time.Sleep(timeout)
		s.bgErr = errors.NewErrorWithSuggestions(errors.BrowserAuthTimedOutErrorMsg, errors.BrowserAuthTimedOutSuggestions)
		s.server.Close()
		s.wait <- true
	}()
	<-s.wait

	defer func() {
		if err := s.server.Shutdown(context.Background()); err != nil {
			output.ErrPrintf(false, "CLI HTTP auth server encountered error while shutting down: %v\n", err)
		}
	}()

	return s.bgErr
}

// callbackHandler serves the route /callback
func (s *authServer) callbackHandler(w http.ResponseWriter, r *http.Request) {
	states, ok := r.URL.Query()["state"]
	if !(ok && states[0] == s.State.MFAProviderState) {
		s.bgErr = fmt.Errorf("authentication callback URL either did not contain a state parameter in query string, or the state parameter was invalid; login will fail")
	}

	fmt.Fprintln(w, mfaCallbackHTML)

	codes, ok := r.URL.Query()["code"]
	if ok {
		s.State.MFAProviderAuthenticationCode = codes[0]
	} else {
		s.bgErr = fmt.Errorf("authentication callback URL did not contain code parameter in query string; login will fail")
	}

	s.wait <- true
}
