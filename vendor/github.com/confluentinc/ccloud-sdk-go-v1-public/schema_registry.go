package ccloud

import (
	"net/http"

	"github.com/dghubble/sling"
)

// SchemaRegistryService provides methods for creating and reading connectors
type SchemaRegistryService struct {
	client *http.Client
	sling  *sling.Sling
}

var _ SchemaRegistry = (*SchemaRegistryService)(nil)

// NewSchemaRegistryService returns a new SchemaRegistryService.
func NewSchemaRegistryService(client *Client) *SchemaRegistryService {
	return &SchemaRegistryService{
		client: client.HttpClient,
		sling:  client.sling,
	}
}

func (s *SchemaRegistryService) CreateSchemaRegistryCluster(config *SchemaRegistryClusterConfig) (*SchemaRegistryCluster, error) {
	body := &CreateSchemaRegistryClusterRequest{Config: config}
	reply := new(CreateSchemaRegistryClusterReply)
	_, err := s.sling.New().Post("/api/schema_registries").BodyProvider(Request(body)).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, WrapErr(err, "error creating SR cluster")
	}
	return reply.Cluster, nil
}

func (s *SchemaRegistryService) GetSchemaRegistryClusters(cluster *SchemaRegistryCluster) ([]*SchemaRegistryCluster, error) {
	reply := new(GetSchemaRegistryClustersReply)
	_, err := s.sling.New().Get("/api/schema_registries").QueryStruct(cluster).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, WrapErr(err, "error retrieving schema registry cluster")
	}
	return reply.Clusters, nil
}

func (s *SchemaRegistryService) GetSchemaRegistryCluster(cluster *SchemaRegistryCluster) (*SchemaRegistryCluster, error) {
	reply := new(GetSchemaRegistryClusterReply)
	_, err := s.sling.New().Get("/api/schema_registries/"+cluster.Id).QueryStruct(cluster).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, WrapErr(err, "error listing schema registry clusters")
	}
	return reply.Cluster, nil
}

func (s *SchemaRegistryService) UpdateSchemaRegistryCluster(cluster *SchemaRegistryCluster) (*SchemaRegistryCluster, error) {
	body := &UpdateSchemaRegistryClusterRequest{Cluster: cluster}
	reply := new(GetSchemaRegistryClusterReply)
	_, err := s.sling.New().Patch("/api/schema_registries/"+cluster.Id).QueryStruct(cluster).BodyProvider(Request(body)).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return nil, WrapErr(err, "error updating schema registry cluster")
	}
	return reply.Cluster, nil
}

func (s *SchemaRegistryService) DeleteSchemaRegistryCluster(cluster *SchemaRegistryCluster) error {
	reply := new(DeleteSchemaRegistryClusterReply)
	_, err := s.sling.New().Delete("/api/schema_registries/"+cluster.Id).QueryStruct(cluster).Receive(reply, reply)
	if err := ReplyErr(reply, err); err != nil {
		return WrapErr(err, "error deleting schema registry cluster")
	}
	return nil
}
