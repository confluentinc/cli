package http

import (
	"net/http"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/log"
)

// UserService provides methods for managing users on Confluent Control Plane.
type UserService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
}

// NewUserService returns a new UserService.
func NewUserService(client *Client) *UserService {
	return &UserService{
		client: client.httpClient,
		logger: client.logger,
		sling:  client.sling,
	}
}

// List returns the users in the authenticated user's organization.
func (s *UserService) List() ([]*orgv1.User, *http.Response, error) {
	reply := new(orgv1.GetUsersReply)
	resp, err := s.sling.New().Get("/api/users").Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch users")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error fetching users")
	}
	return reply.Users, resp, nil
}

// Describe returns details for a given user.
func (s *UserService) Describe(user *orgv1.User) (*orgv1.User, *http.Response, error) {
	reply := new(orgv1.GetUsersReply)
	resp, err := s.sling.New().Get("/api/users").QueryStruct(user).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch kafka users")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error fetching kafka users")
	}
	// Since we're hitting the get-all endpoint instead of get-one, simulate a NotFound error if no matches return
	if len(reply.Users) == 0 {
		return nil, resp, errNotFound
	}
	return reply.Users[0], resp, nil
}
