package http

import (
	"net/http"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	proto "github.com/confluentinc/cli/shared/connect"
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
		sling: sling.New().Client(client.httpClient).Base(client.baseURL),
	}
}

// List returns the authenticated user's connect clusters.
func (s *ConnectService) List(accountID string) ([]*proto.Connector, *http.Response, error) {
	clusters := new(proto.ListResponse)
	confluentError := new(ConfluentError)
	resp, err := s.sling.New().Get("/api/connectors?account_id="+accountID).Receive(clusters, confluentError)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch connectors")
	}
	return clusters.Clusters, resp, nil
}
