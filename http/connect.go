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
	reply := new(schedv1.GetConnectClustersReply)
	resp, err := s.sling.New().Get("/api/connectors?account_id="+accountID).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch connectors")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error fetching connectors")
	}
	return reply.Clusters, resp, nil
}

// Describe returns details for a given connect cluster.
func (s *ConnectService) Describe(cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, *http.Response, error) {
	reply := new(schedv1.GetConnectClustersReply)
	resp, err := s.sling.New().Get("/api/connectors").QueryStruct(cluster).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch connectors")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error fetching connectors")
	}
	return reply.Clusters[0], resp, nil
}

func (s *ConnectService) CreateS3Sink(config *schedv1.ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, *http.Response, error) {
	body := &schedv1.CreateConnectS3SinkClusterRequest{Config: config}
	reply := new(schedv1.CreateConnectS3SinkClusterReply)
	resp, err := s.sling.New().Post("/api/connectors/s3-sinks").BodyJSON(body).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to create connector")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error creating connector")
	}
	return reply.Cluster, resp, nil
}
