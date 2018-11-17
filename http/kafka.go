package http

import (
	"fmt"
	"bytes"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/codyaray/sling"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
)

// KafkaAPI REST endpoint templates as described in KafkaResource.java
// https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaResource.java
const (
	// APIBASE represents the root for all control-plane endpoints.
	APIBASE = "/2.0/kafka/%s"
	// TOPICS represents the root for all topic related control-plane activities.
	TOPICS = APIBASE + "/topics"
	// TOPIC represents the root for all control-plane activities related to a specific topic.
	TOPIC = TOPICS + "/%s"
	// TOPICCONFIG endpoint which updates topic-level configuration overrides.
	TOPICCONFIG = TOPIC + "/config"
	// ACL endpoint which represents the root for all ACL related control-plane activities.
	ACL = APIBASE + "/acls"
	// ACLSEARCH endpoint which lists all ACLs associated with a specific resource.
	ACLSEARCH = ACL + ":search"
)

// KafkaService provides methods for creating and reading kafka clusters.
type KafkaService struct {
	client *http.Client
	sling  *sling.Sling
	logger *log.Logger
	api    map[string]string
}

var _ Kafka = (*KafkaService)(nil)

// NewKafkaService returns a new KafkaService.
func NewKafkaService(client *Client) *KafkaService {
	return &KafkaService{
		client: client.httpClient,
		logger: client.logger,
		sling:  client.sling,
		api:    make(map[string]string),
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

// KafkaAPI returns a sling for the service control plane
func (s *KafkaService) kafkaAPI(cluster *schedv1.KafkaCluster) (string, error) {
	if handle, ok := s.api[cluster.Id]; ok{
		return handle, nil
	}

	cluster, _, err := s.Describe(cluster)
	if err != nil {
		return "", err
	}

	s.api[cluster.Id] = cluster.ApiEndpoint
	return cluster.ApiEndpoint, nil
}

// ListTopics lists all non-internal topics in the current Kafka cluster context
func (s *KafkaService) ListTopics(cluster *schedv1.KafkaCluster) ([]kafka.KafkaTopicDescription, error) {
	var topics []kafka.KafkaTopicDescription

	return topics, s.handleAPIRequest(cluster, newBasicRequest(
		http.MethodGet, TOPICS, cluster.Id), &topics)
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (s *KafkaService) CreateTopic(cluster *schedv1.KafkaCluster, topic *kafka.Topic) (error) {
	return s.handleAPIRequest(cluster,
		newOptionsBasicRequest(http.MethodPut, TOPICS,
			url.Values{"validate": []string{"false"}}, topic.Spec, cluster.Id), nil)
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (s *KafkaService) DescribeTopic(cluster *schedv1.KafkaCluster, topic *kafka.Topic) (*kafka.KafkaTopicDescription, error) {
	resp := &kafka.KafkaTopicDescription{}
	topicConf, err := s.ListTopicConfig(cluster, topic)
	if err != nil {
		return nil, err
	}
	resp.Config = topicConf.Entries

	return resp, s.handleAPIRequest(cluster, newBasicRequest(http.MethodGet, TOPIC, cluster.Id, topic.Spec.Name), &resp)
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (s *KafkaService) DeleteTopic(cluster *schedv1.KafkaCluster, topic *kafka.Topic) (error) {
	return s.handleAPIRequest(cluster, newBasicRequest(http.MethodDelete, TOPIC, cluster.Id, topic.Spec.Name), nil)
}

// ListTopicConfig returns a Kafka Topic configuration from the current Kafka Cluster context
func (s *KafkaService) ListTopicConfig(cluster *schedv1.KafkaCluster, topic *kafka.Topic) (*kafka.TopicConfig, error) {
	topicConf := &kafka.TopicConfig{}
	return topicConf, s.handleAPIRequest(cluster, newBasicRequest(http.MethodGet, TOPICCONFIG, cluster.Id, topic.Spec.Name), &topicConf)
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (s *KafkaService) UpdateTopic(cluster *schedv1.KafkaCluster, topic *kafka.Topic) (error) {
	// first fetch the original topic configuration then override where appropriate
	topicConf, err := s.ListTopicConfig(cluster, topic)
	if err != nil {
		return err
	}

	if err := update(topicConf.Entries, topic.Spec.Configs); err != nil {
		return err
	}

	return s.handleAPIRequest(cluster, newRequest(http.MethodPut, TOPICCONFIG, topicConf, cluster.Id, topic.Spec.Name), nil)
}

// ListACL lists all ACLs for a given principal or resource
func (s *KafkaService) ListACL(conf *kafka.ACLFilter) (*kafka.KafkaAPIACLFilterReply, error) {
	acls := &kafka.KafkaAPIACLFilterReply{}
	return acls, nil /*s.handleAPIRequest(s.api.sling.
		Post(fmt.Sprintf(ACLSEARCH, s.api.id)).BodyJSON(conf), &acls.Results)*/
}

// CreateACL registers a new ACL with the current Kafka Cluster context
func (s *KafkaService) CreateACL(conf *kafka.ACLSpec) (error) {
	return nil /*s.handleAPIRequest(s.api.sling.Post(fmt.Sprintf(ACL, s.api.id)).
		BodyJSON([]*kafka.ACLSpec{conf}), nil)*/
}

// DeleteACL delete an ACL with the current Kafka Cluster context
func (s *KafkaService) DeleteACL(conf *kafka.ACLFilter) (error) {
	return nil /*s.handleAPIRequest(s.api.sling.Delete(fmt.Sprintf(ACL, s.api.id)).
		BodyJSON(conf), nil)*/
}

// update updates an KafkaTopicConfigEntries with the values in update
func update(original []*kafka.KafkaTopicConfigEntry, updates map[string]string) error {
	for idx := range original {
		if value, ok := updates[original[idx].Name]; ok {
			original[idx].Value = value
			delete(updates, original[idx].Name)
		}
	}
	if len(updates) > 0 {
		return fmt.Errorf("invalid configuration entries %s", updates)
	}
	return nil
}

// REST API request
type api struct {
	method    string
	endpoint  string
	query url.Values
	arguments []interface{}
	body      interface{}
}

// newRequest returns new KafkaAPI request */
func newRequest(method, endpoint string, body interface{}, arguments ...interface{}) *api {
	return newOptionsRequest(method, endpoint, nil, body, arguments...)
}

func newBasicRequest(method, endpoint string, arguments ...interface{}) *api {
	return newRequest(method, endpoint, nil, arguments...)
}

func newOptionsBasicRequest(method, endpoint string, query url.Values, arguments ...interface{}) *api {
	return newOptionsRequest(method, endpoint, query, nil, arguments...)
}

func newOptionsRequest(method, endpoint string, query url.Values, body interface{}, arguments ...interface{}) * api {
	return &api{
		method:    method,
		endpoint:  endpoint,
		arguments: arguments,
		body:      body,
	}
}

// handleAPIRequest handles the interaction between the plugin and the current Kafka Cluster's control-plane
func (s *KafkaService) handleAPIRequest(cluster *schedv1.KafkaCluster, request *api, success interface{}) error {
	base, err := s.kafkaAPI(cluster)
	if err != nil {
		return err
	}

	endpoint, err := url.Parse(fmt.Sprintf(base + request.endpoint, request.arguments...))
	if err != nil {
		return err
	}
	endpoint.RawQuery = request.query.Encode()

	if err != nil {
		return err
	}

	outbuf, err := json.Marshal(request.body)
	if err != nil {
		return err
	}

	req := &http.Request{
		Method: request.method,
		URL:    endpoint,
		Body:   ioutil.NopCloser(bytes.NewBuffer(outbuf)),
		Close: true,
	}

	resp, err := s.client.Do(req)

	err = shared.HandleKafkaAPIError(resp, err)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	s.logger.Log("msg", fmt.Sprintf("request: %s %s result: %+v", req.Method, req.URL, resp.StatusCode))

	if success != nil {
		return json.NewDecoder(resp.Body).Decode(success)
	}

	return nil
}
