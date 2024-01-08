package ccloud

import (
	"net/http"

	"github.com/dghubble/sling"
)

var _ EnvironmentMetadata = (*EnvironmentMetadataService)(nil)

// EnvironmentMetadataServices provides methods for getting information about cloud regions.
type EnvironmentMetadataService struct {
	client *http.Client
	sling  *sling.Sling
}

func NewEnvironmentMetadataService(client *Client) *EnvironmentMetadataService {
	return &EnvironmentMetadataService{
		client: client.HttpClient,
		sling:  client.sling,
	}
}

func (s *EnvironmentMetadataService) Get() ([]*CloudMetadata, error) {
	reply := new(GetEnvironmentMetadataReply)
	_, err := s.sling.New().Get("/api/env_metadata").Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, err
	}
	return reply.Clouds, nil
}
