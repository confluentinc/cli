package http

import (
	"fmt"
	"net/http"
	"strconv"

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

var _ User = (*UserService)(nil)

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

func formatInt32(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

// Describe returns details for a given user.
func (s *UserService) CreateServiceAccount(user *orgv1.User) (*orgv1.User, *http.Response, error) {
	body := &orgv1.CreateServiceAccountRequest{User: user}
	reply := new(orgv1.CreateServiceAccountReply)
	userId := formatInt32(user.Id)
	fmt.Printf("Http: User go  %d" , userId)
	s.logger.Log("msg", "Http: Create SA")
	//data := "/api/users"+userId
	s.logger.Log("Http: User go  %s" , body)
	resp, err := s.sling.New().Post("/api/users"+userId).BodyJSON(body).Receive(reply, reply)
	if err != nil {
		s.logger.Log("Unable to create SA %s" , err)
		return nil, resp, errors.Wrap(err, "Unnnable to create service account")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error creating service account")
	}

	return reply.User, resp, nil
}

// Describe returns details for a given user.
func (s *UserService) UpdateServiceAccount(user *orgv1.User) (*http.Response, error) {
	body := &orgv1.UpdateServiceAccountRequest{User: user}
	reply := new(orgv1.UpdateServiceAccountReply)
	userId := formatInt32(user.Id)
	resp, err := s.sling.New().Post("/api/users"+userId).BodyJSON(body).Receive(reply, reply)
	if err != nil {
		return resp, errors.Wrap(err, "unable to update service account")
	}
	if reply.Error != nil {
		return resp, errors.Wrap(reply.Error, "error updating service account")
	}

	return resp, nil
}

// Describe returns details for a given user.
func (s *UserService) DeactivateServiceAccount(user *orgv1.User) (*http.Response, error) {
	body := &orgv1.DeactivateServiceAccountRequest{User: user}
	reply := new(orgv1.DeactivateServiceAccountReply)
	userId := formatInt32(user.Id)
	resp, err := s.sling.New().Delete("/api/users"+userId).BodyJSON(body).Receive(reply, reply)
	if err != nil {
		return resp, errors.Wrap(err, "unable to deactivate service account")
	}
	if reply.Error != nil {
		return resp, errors.Wrap(reply.Error, "error deactivate service account")
	}

	return resp, nil
}

// Describe returns details for a given user.
func (s *UserService) GetServiceAccounts(user *orgv1.User) ([]*orgv1.User, *http.Response, error) {
	body := &orgv1.GetServiceAccountsRequest{User: user}
	reply := new(orgv1.GetServiceAccountsReply)
	userId := formatInt32(user.Id)
	resp, err := s.sling.New().Get("/api/users"+userId).BodyJSON(body).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to list service account")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error list service account")
	}

	return reply.Users, resp, nil
}
