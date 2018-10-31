package http

import (
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/codyaray/sling"
	"github.com/pkg/errors"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
)

// KafkaAPI REST endpoint templates as described in KafkaResource.java
// https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java

// APIBASE represents the root for all control-plane endpoints.
const APIBASE = "/2.0/kafka/%s"
// TOPICS represents the root for all topic related control-plane activities.
const TOPICS = APIBASE + "/topics"
// TOPIC represents the root for all control-plane activities related to a specific topic.
const TOPIC = TOPICS + "/%s"
// TOPICCONFIG endpoint which updates topic-level configuration overrides.
const TOPICCONFIG = TOPIC + "/config"
// TOPICCONFIGDEFAULT returns Kafka Topic defaults for the current Kafka Cluster context
const TOPICCONFIGDEFAULT = APIBASE + "/topic-defaults"
// ACL endpoint which represents the root for all ACL related control-plane activities.
const ACL = APIBASE + "/acls"
// ACLSEARCH endpoint which lists all ACLs associated with a specific resource.
const ACLSEARCH = ACL + ":search"

// KafkaService provides methods for creating and reading kafka clusters.
type KafkaService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
	api    *KafkaAPI
}

// KafkaAPI provides methods for interacting with the Kafka Cluster's control-plane
type KafkaAPI struct {
	id    string
	sling *sling.Sling
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
	if cluster.Id == "" {
		return nil, shared.ErrNotFound
	}
	body := &schedv1.DeleteKafkaClusterRequest{Cluster: cluster}
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

// ConfigureKafkaAPI returns a sling for the service control plane
func (s *KafkaService) ConfigureKafkaAPI(clusterID, apiEndpoint string) {
	s.api = &KafkaAPI{
		id:    clusterID,
		sling: sling.New().Client(s.client).Base(apiEndpoint),
	}
}

// ListTopic lists all non-internal topics in the current Kafka cluster context
func (s *KafkaService) ListTopic() ([]kafka.KafkaTopicDescription, error) {
	var topics []kafka.KafkaTopicDescription
	return topics, s.handleAPIRequest(s.api.sling.Get(fmt.Sprintf(TOPICS, s.api.id)), &topics)
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (s *KafkaService) CreateTopic(conf *kafka.KafkaAPITopicRequest) (error) {
	resp := s.handleAPIRequest(s.api.sling.
		Put(fmt.Sprintf(TOPICS, s.api.id)).
		BodyJSON(conf.Spec).
		QueryStruct(struct {
			Validate bool `url:"validate"`
		}{conf.Validate}), nil)
	s.logger.Log("ct err: ", fmt.Sprintf("%+v", resp))
	return resp
	}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (s *KafkaService) DescribeTopic(conf *kafka.KafkaAPITopicRequest) (*kafka.KafkaTopicDescription, error) {
	var topic *kafka.KafkaTopicDescription
	return topic, s.handleAPIRequest(s.api.sling.
		Get(fmt.Sprintf(TOPIC, s.api.id, conf.Spec.Name)), &topic)
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (s *KafkaService) DeleteTopic(conf *kafka.KafkaAPITopicRequest) (error) {
	return s.handleAPIRequest(s.api.sling.
		Delete(fmt.Sprintf(TOPIC, s.api.id, conf.Spec.Name)), nil)
}

// ListTopicConfig returns a Kafka Topic configuration from the current Kafka Cluster context
func (s *KafkaService) ListTopicConfig(conf *kafka.KafkaAPITopicRequest) ([]*kafka.KafkaTopicConfigEntry, error) {
	var topicConf *kafka.KafkaAPITopicConfigRequest
	return topicConf.Entries, s.handleAPIRequest(s.api.sling.Get(fmt.Sprintf(TOPICCONFIG, s.api.id, conf.Spec.Name)), &topicConf)
}

// ListTopicConfigDefault returns the default Kafka Topic configurations for the current Kafka Cluster context
func (s *KafkaService) ListTopicConfigDefault(conf *kafka.KafkaAPITopicRequest) (map[string]string, error) {
	var topicConf *kafka.KafkaTopicSpecification
	return topicConf.Configs, s.handleAPIRequest(s.api.sling.Get(fmt.Sprintf(TOPICCONFIGDEFAULT, s.api.id)), &topicConf)
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (s *KafkaService) UpdateTopic(conf *kafka.KafkaAPITopicRequest) (error) {
	// first fetch the original topic configuration then override where appropriate
	originals, err := s.ListTopicConfig(conf)
	if err != nil {
		return err
	}
	update(originals, conf.Spec.Configs)

	return s.handleAPIRequest(s.api.sling.
		Put(fmt.Sprintf(TOPICCONFIG, s.api.id, conf.Spec.Name)).
		BodyJSON(&kafka.KafkaAPITopicConfigRequest{}), nil)
}

// ListACL lists all ACLs for a given principal or resource
func (s *KafkaService) ListACL(conf *kafka.KafkaAPIACLFilterRequest) (*kafka.KafkaAPIACLFilterReply, error) {
	acls := &kafka.KafkaAPIACLFilterReply{}
	return acls, s.handleAPIRequest(s.api.sling.
		Post(fmt.Sprintf(ACLSEARCH, s.api.id)).BodyJSON(conf), &acls.Results)
}

// CreateACL registers a new ACL with the current Kafka Cluster context
func (s *KafkaService) CreateACL(conf *kafka.KafkaAPIACLRequest) (error) {
	return s.handleAPIRequest(s.api.sling.Post(fmt.Sprintf(ACL, s.api.id)).
		BodyJSON([]*kafka.KafkaAPIACLRequest{conf}), nil)
}

// DeleteACL delete an ACL with the current Kafka Cluster context
func (s *KafkaService) DeleteACL(conf *kafka.KafkaAPIACLFilterRequest) (error) {
	return s.handleAPIRequest(s.api.sling.Delete(fmt.Sprintf(ACL, s.api.id)).
		BodyJSON(conf), nil)
}

// update updates an KafkaTopicConfigEntries with the values in update
func update(original []*kafka.KafkaTopicConfigEntry, updates map[string]string) {
	for idx := range original {
		if value, ok := updates[original[idx].Name]; ok {
			original[idx].Value = value
			delete(updates, original[idx].Name)
		}
	}
}

// handleAPIRequest handles the interaction between the plugin and the current Kafka Cluster's control-plane
func (s *KafkaService) handleAPIRequest(sling *sling.Sling, success interface{}) (error) {
	if s.api == nil {
		return shared.ErrNotImplemented
	}

	req, err := sling.Request()
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	s.logger.Log("msg", fmt.Sprintf("request: %s %s result: %+v", req.Method, req.URL, resp.StatusCode))

	err = shared.HandleKafkaAPIError(resp, err)

	if err != nil {
		return err
	}

	if success != nil {
		return json.NewDecoder(resp.Body).Decode(success)
	}

	return nil
}
