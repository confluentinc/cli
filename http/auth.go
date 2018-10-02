package http

import (
	"net/http"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

const (
	loginPath = "/api/sessions"
	mePath    = "/api/me"
)

var (
	// special case: login doesn't return a JSON error, just a 404
	errUnauthorized = &corev1.Error{Code: 401, Message: "unauthorized"}
)

// AuthService provides methods for authenticating to Confluent Control Plane
type AuthService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
}

var _ Auth = (*AuthService)(nil)

// NewAuthService returns a new AuthService.
func NewAuthService(client *Client) *AuthService {
	return &AuthService{
		client: client.httpClient,
		logger: client.logger,
		sling:  client.sling,
	}
}

// Login attempts to login a user by username and password, returning either a token or an error.
func (a *AuthService) Login(username, password string) (string, error) {
	payload := map[string]string{"email": username, "password": password}
	req, err := a.sling.New().Post(loginPath).BodyJSON(payload).Request()
	if err != nil {
		return "", err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
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

// User returns the AuthConfig for the authenticated user.
func (a *AuthService) User() (*shared.AuthConfig, error) {
	me := &orgv1.GetUserReply{}
	_, err := a.sling.New().Get(mePath).Receive(me, me)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch user info") // you just don't get /me
	}
	if me.Error != nil {
		return nil, errors.Wrap(err, "error fetching user info")
	}
	return &shared.AuthConfig{
		User:    me.User,
		Account: me.Account,
	}, nil
}
