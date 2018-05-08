package http

import (
	"net/http"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/log"
)

// ConnectService provides methods for creating and reading connectors
type ConnectService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
}

// NewConnectService returns a new ConnectService.
func NewConnectService(client *Client) *ConnectService {
	return &ConnectService{
		client: client.httpClient,
		logger: client.logger,
		sling:  client.sling,
	}
}

// List returns the authenticated user's connect clusters within a given account.
func (s *ConnectService) List(accountID string) ([]*schedv1.ConnectCluster, *http.Response, error) {
	clusters := new(schedv1.GetConnectClustersReply)
	resp, err := s.sling.New().Get("/api/connectors?account_id="+accountID).Receive(clusters, clusters)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch connectors")
	}
	if clusters.Error != nil {
		return nil, resp, errors.Wrap(clusters.Error, "error fetching connectors")
	}
	return clusters.Clusters, resp, nil
}
