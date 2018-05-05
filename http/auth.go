package http

import (
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
	// special case: login doesn't return a JSON error, just a 404
	errUnauthorized = &ApiError{Err: &apiError{Code: 401, Message: "unauthorized"}}
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
		sling:  client.sling,
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
		return "", errUnauthorized
	case http.StatusOK:
		var token string
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "auth_token" {
				token = cookie.Value
				break
			}
		}
		if token == "" {
			return "", errUnauthorized
		}
		return token, nil
	}
	return "", errUnauthorized
}

func (a *AuthService) User() (*shared.AuthConfig, error) {
	me := &shared.AuthConfig{}
	apiError := &ApiError{}
	_, err := a.sling.New().Get(mePath).Receive(me, apiError)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch user info") // you just don't get /me
	}
	return me, apiError.OrNil()
}
