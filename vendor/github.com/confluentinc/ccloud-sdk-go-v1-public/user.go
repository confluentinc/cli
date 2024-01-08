package ccloud

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

const (
	// USERS base endpoint for user operations
	USERS = "/api/users"
	// USER specific user operation endpoint
	USER            = USERS + "/%s"
	SERVICEACCOUNTS = "/api/service_accounts"
	SERVICEACCOUNT  = SERVICEACCOUNTS + "/%d"
	// deprecated: legacy endpoint for email invite to join org. org id in path is ignored, only the org id in the auth context matters
	INVITEURL = "/api/organizations/0/invites"
	// new invitation flow
	INVITATIONURL = "/api/invitations"
	// USER_PROFILES base endpoint for use profile operations
	USER_PROFILES = "/api/user_profiles"
	// customer-initiated org deletion flow
	SUSPEND_ORG_URL_TEMPLATE = "/api/organizations/%d/suspend"
)

// UserService provides methods for managing users on Confluent Control Plane.
type UserService struct {
	client *http.Client
	sling  *sling.Sling
}

var _ UserInterface = (*UserService)(nil)

// NewUserService returns a new UserService.
func NewUserService(client *Client) *UserService {
	return &UserService{
		client: client.HttpClient,
		sling:  client.sling,
	}
}

// List returns the users in the authenticated user's organization.
func (s *UserService) List() ([]*User, error) {
	reply := new(GetUsersReply)
	_, err := s.sling.New().Get(USERS).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, WrapErr(err, "error retrieving users")
	}
	return reply.Users, nil
}

// GetServiceAccounts returns a collection of service account users.
func (s *UserService) GetServiceAccounts() ([]*User, error) {
	reply := new(GetServiceAccountsReply)

	_, err := s.sling.New().Get(SERVICEACCOUNTS).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, WrapErr(err, "error listing service accounts")
	}

	return reply.Users, nil
}

// GetServiceAccount returns a service account user.
func (s *UserService) GetServiceAccount(id int32) (*User, error) {
	url := fmt.Sprintf(SERVICEACCOUNT, id)
	res := new(GetServiceAccountReply)

	_, err := s.sling.New().Get(url).Receive(res, res)
	return res.User, WrapErr(ReplyErr(res, err), "error getting service account")
}

// LoginRealm gets a user's login realm information. can be used to determine if the user is an SSO user
func (s *UserService) LoginRealm(req *GetLoginRealmRequest) (*GetLoginRealmReply, error) {
	reply := new(GetLoginRealmReply)

	if req == nil || req.Email == "" {
		return nil, errors.New("non-nil user object must be passed and have Email set in order to call LoginRealm")
	}

	_, err := s.sling.New().Get("/api/login/realm").QueryStruct(req).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, WrapErr(err, "error checking login realm")
	}

	return reply, nil
}
