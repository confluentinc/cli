package http

import (
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

const (
	loginPath = "/api/sessions"
	mePath = "/api/me"
)

var (
	ErrUnauthorized = fmt.Errorf("unauthorized")
)

// AuthService provides methods for authenticating to Confluent Control Plane
type AuthService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
}

// NewAuthService returns a new AuthService.
func NewAuthService(client *Client) *AuthService {
	return &AuthService{
		client: client.httpClient,
		logger: client.logger,
		sling: sling.New().Client(client.httpClient).Base(client.baseURL),
	}
}

func (a *AuthService) Login(username, password string) (string, error) {
	payload := map[string]string{"email": username, "password": password}
	req, err := a.sling.New().Post(loginPath).BodyJSON(payload).Request()
	if err != nil {
		return "", err
	}
	resp, err := a.client.Do(req)
	switch resp.StatusCode {
	case http.StatusNotFound:
		// For whatever reason, 404 is returned if credentials are bad
		return "", ErrUnauthorized
	case http.StatusOK:
		var token string
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "auth_token" {
				token = cookie.Value
				break
			}
		}
		if token == "" {
			return "", ErrUnauthorized
		}
		return token, nil
	}
	return "", ErrUnauthorized
}

func (a *AuthService) User() (*shared.AuthConfig, error) {
	me := &shared.AuthConfig{}
	confluentError := &ConfluentError{}
	_, err := a.sling.New().Get(mePath).Receive(me, confluentError)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch user info") // you just don't get me
	}
	return me, nil
}
