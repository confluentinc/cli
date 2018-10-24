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

var _ Kafka = (*KafkaService)(nil)

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

// Create provisions a new kafka cluster as described by the given config.
func (s *KafkaService) Create(config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, *http.Response, error) {
	body := &schedv1.CreateKafkaClusterRequest{Config: config}
	reply := new(schedv1.CreateKafkaClusterReply)
	resp, err := s.sling.New().Post("/api/clusters").BodyJSON(body).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to create kafka cluster")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error creating kafka cluster")
	}
	return reply.Cluster, resp, nil
}

// Delete terminates the given kafka cluster.
func (s *KafkaService) Delete(cluster *schedv1.KafkaCluster) (*http.Response, error) {
	return s.delete(cluster, false)
}

// DeletePhysical terminates the given physical kafka cluster and all active logical kafkas.
// This is only supported for users who login with an @confluent.io email address.
func (s *KafkaService) DeletePhysical(cluster *schedv1.KafkaCluster) (*http.Response, error) {
	return s.delete(cluster, true)
}

func (s *KafkaService) delete(cluster *schedv1.KafkaCluster, destroyPhysical bool) (*http.Response, error) {
	if cluster.Id == "" {
		return nil, shared.ErrNotFound
	}
	body := &schedv1.DeleteKafkaClusterRequest{Cluster: cluster, DestroyPhysical: destroyPhysical}
	reply := new(schedv1.DeleteKafkaClusterReply)
	resp, err := s.sling.New().Delete("/api/clusters/"+cluster.Id).BodyJSON(body).Receive(reply, reply)
	if err != nil {
		return resp, errors.Wrap(err, "unable to delete kafka cluster: "+cluster.Id)
	}
	if reply.Error != nil {
		return resp, errors.Wrap(reply.Error, "error deleting kafka cluster")
	}
	return resp, nil
}
