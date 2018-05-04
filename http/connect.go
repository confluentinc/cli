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
		sling:  client.sling,
	}
}

// List returns the authenticated user's connect clusters within a given account.
func (s *ConnectService) List(accountID string) ([]*proto.Connector, *http.Response, error) {
	clusters := new(proto.ListResponse)
	apiErr := new(ApiError)
	resp, err := s.sling.New().Get("/api/connectors?account_id="+accountID).Receive(clusters, apiErr)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch connectors")
	}
	return clusters.Clusters, resp, apiErr.OrNil()
}
