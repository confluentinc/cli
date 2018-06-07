package http

import (
	"net/http"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

// KafkaService provides methods for creating and reading kafka clusters
type KafkaService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
}

// NewKafkaService returns a new KafkaService.
func NewKafkaService(client *Client) *KafkaService {
	return &KafkaService{
		client: client.httpClient,
		logger: client.logger,
		sling:  client.sling,
	}
}

// List returns the authenticated user's kafka clusters.
func (s *KafkaService) List(cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, *http.Response, error) {
	reply := new(schedv1.GetKafkaClustersReply)
	resp, err := s.sling.New().Get("/api/clusters").QueryStruct(cluster).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch kafka clusters")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error fetching kafka clusters")
	}
	return reply.Clusters, resp, nil
}

// Describe returns details for a given kafka cluster.
func (s *KafkaService) Describe(cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, *http.Response, error) {
	if cluster.Id == "" {
		return nil, nil, shared.ErrNotFound
	}
	reply := new(schedv1.GetKafkaClusterReply)
	resp, err := s.sling.New().Get("/api/clusters/"+cluster.Id).QueryStruct(cluster).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch kafka clusters")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error fetching kafka clusters")
	}
	return reply.Cluster, resp, nil
}
