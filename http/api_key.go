package http

import (
	"net/http"
	"strconv"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

// APIKeyService provides methods for managing API keys on Confluent Control Plane.
type APIKeyService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
}

var _ APIKey = (*APIKeyService)(nil)

// NewAPIKeyService returns a new APIKeyService.
func NewAPIKeyService(client *Client) *APIKeyService {
	return &APIKeyService{
		client: client.httpClient,
		logger: client.logger,
		sling:  client.sling,
	}
}

// Create makes a new API Key
func (s *APIKeyService) Create(key *schedv1.ApiKey) (*schedv1.ApiKey, *http.Response, error) {
	request := &schedv1.CreateApiKeyRequest{ApiKey: key}
	reply := new(schedv1.CreateApiKeyReply)
	resp, err := s.sling.New().Post("/api/api_keys").BodyJSON(request).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to create API key")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error creating API key")
	}
	return reply.ApiKey, resp, nil
}

// Delete removes an API Key
func (s *APIKeyService) Delete(key *schedv1.ApiKey) (*http.Response, error) {
	// check if key's id is the zero value - implies that this required field is not set
	if key.Id == 0 {
		return nil, shared.ErrNotFound
	}

	request := &schedv1.DeleteApiKeyRequest{ApiKey: key}
	reply := new(schedv1.DeleteApiKeyReply)
	resp, err := s.sling.New().Delete("/api/api_keys/"+strconv.Itoa(int(key.Id))).BodyJSON(request).Receive(reply, reply)
	if err != nil {
		return resp, errors.Wrap(err, "unable to delete API key")
	}
	if reply.Error != nil {
		return resp, errors.Wrap(reply.Error, "error deleting API key")
	}
	return resp, nil
}
