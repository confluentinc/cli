package http

import (
	"net/http"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

// KsqlService provides methods for creating and reading ksql clusters
type KsqlService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
}

var _ KSQL = (*KsqlService)(nil)

// NewKsqlService returns a new KsqlService.
func NewKsqlService(client *Client) *KsqlService {
	return &KsqlService{
		client: client.httpClient,
		logger: client.logger,
		sling:  client.sling,
	}
}

// List returns the authenticated user's connect clusters within a given account.
func (s *KsqlService) List(cluster *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, *http.Response, error) {
	reply := new(schedv1.GetKSQLClustersReply)
	resp, err := s.sling.New().Get("/api/ksqls").QueryStruct(cluster).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch ksql clusters")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error fetching ksql clusters")
	}
	return reply.Clusters, resp, nil
}

// Describe returns details for a given cluster.
func (s *KsqlService) Describe(cluster *schedv1.KSQLCluster) (*schedv1.KSQLCluster, *http.Response, error) {
	if cluster.Id == "" {
		return nil, nil, shared.ErrNotFound
	}
	reply := new(schedv1.GetKSQLClusterReply)
	resp, err := s.sling.New().Get("/api/ksqls/"+cluster.Id).QueryStruct(cluster).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to fetch ksql cluster")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error fetching ksql cluster")
	}
	return reply.Cluster, resp, nil
}

// Delete destroys a given cluster.
func (s *KsqlService) Delete(cluster *schedv1.KSQLCluster) (*http.Response, error) {
	if cluster.Id == "" {
		return nil, shared.ErrNotFound
	}
	reply := new(schedv1.DeleteKSQLClusterReply)
	resp, err := s.sling.New().Delete("/api/ksqls/"+cluster.Id).QueryStruct(cluster).Receive(reply, reply)
	if err != nil {
		return resp, errors.Wrap(err, "unable to delete ksql")
	}
	if reply.Error != nil {
		return resp, errors.Wrap(reply.Error, "error deleting ksql")
	}
	return resp, nil
}

// Create provisions a new kafka cluster as described by the given config.
func (s *KsqlService) Create(config *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, *http.Response, error) {
	body := &schedv1.CreateKSQLClusterRequest{Config: config}
	reply := new(schedv1.CreateKSQLClusterReply)
	resp, err := s.sling.New().Post("/api/ksqls").BodyJSON(body).Receive(reply, reply)
	if err != nil {
		return nil, resp, errors.Wrap(err, "unable to create ksql cluster")
	}
	if reply.Error != nil {
		return nil, resp, errors.Wrap(reply.Error, "error creating ksql cluster")
	}
	return reply.Cluster, resp, nil
}
