package ccloud

import (
	"net/http"

	"github.com/dghubble/sling"
)

// AuthService provides methods for authenticating to Confluent Control Plane
type AuthService struct {
	client *http.Client
	sling  *sling.Sling
}

var _ Auth = (*AuthService)(nil)

// NewAuthService returns a new AuthService.
func NewAuthService(client *Client) *AuthService {
	return &AuthService{
		client: client.HttpClient,
		sling:  client.sling,
	}
}

// Login attempts to log a user in with an Auth0 ID token, returning either a (CCloud) token or an error.
func (a *AuthService) Login(req *AuthenticateRequest) (*AuthenticateReply, error) {
	return a.login("/api/sessions", req)
}

func (a *AuthService) OktaLogin(req *AuthenticateRequest) (*AuthenticateReply, error) {
	return a.login("/api/okta/auth/sessions", req)
}

func (a *AuthService) login(path string, req *AuthenticateRequest) (*AuthenticateReply, error) {
	res := new(AuthenticateReply)

	httpResp, err := a.sling.New().Post(path).BodyProvider(Request(req)).Receive(res, res)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		switch httpResp.StatusCode {
		case http.StatusForbidden:
			if res.Error != nil && res.Error.Message == "Your organization has been suspended, please contact support if you want to unsuspend it." {
				return nil, &SuspendedOrganizationError{}
			}
			return nil, &InvalidLoginError{}
		case http.StatusInternalServerError:
			return nil, &UnknownLoginError{}
		default:
			return nil, &InvalidLoginError{}
		}
	}

	return res, nil
}

// User returns the AuthConfig for the authenticated user.
func (a *AuthService) User() (*GetMeReply, error) {
	reply := &GetMeReply{}
	_, err := a.sling.New().Get("/api/me").Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, WrapErr(err, "error retrieving user")
	}
	return &GetMeReply{
		User:         reply.User,
		Account:      reply.Account,
		Accounts:     reply.Accounts,
		Organization: reply.Organization,
	}, nil
}
