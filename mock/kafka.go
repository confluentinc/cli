package mock

import (
	"context"
	"fmt"
	"reflect"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/golang/protobuf/proto"
)

// Compile-time check interface adherence
var _ chttp.Kafka = (*MockKafka)(nil)

type MockKafka struct {
	Expect chan interface{}
}

func NewKafkaMock(value interface{}, expect chan interface{}) error {
	client := &MockKafka{expect}
	rv := reflect.ValueOf(value)
	rv.Elem().Set(reflect.ValueOf(client))
	return nil
}

func (m *MockKafka) CreateAPIKey(_ context.Context, apiKey *authv1.APIKey) (*authv1.APIKey, error) {
	return apiKey, nil
}

func (m *MockKafka) List(_ context.Context, cluster *kafkav1.Cluster) ([]*kafkav1.Cluster, error) {
	return []*kafkav1.Cluster{cluster}, nil
}

func (m *MockKafka) Describe(_ context.Context, cluster *kafkav1.Cluster) (*kafkav1.Cluster, error) {
	return cluster, nil
}

func (m *MockKafka) Create(_ context.Context, config *kafkav1.ClusterConfig) (*kafkav1.Cluster, error) {
	return &kafkav1.Cluster{}, nil
}

func (m *MockKafka) Delete(_ context.Context, cluster *kafkav1.Cluster) error {
	return nil
}

func (m *MockKafka) ListTopics(ctx context.Context, cluster *kafkav1.Cluster) ([]*kafkav1.TopicDescription, error) {
	return []*kafkav1.TopicDescription{
		{Name: "test1"},
		{Name: "test2"},
		{Name: "test3"}}, nil
}

func (m *MockKafka) DescribeTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) (*kafkav1.TopicDescription, error) {
	node := &kafkav1.KafkaNode{Id: 1}
	tp := &kafkav1.TopicPartitionInfo{Leader: node, Replicas: []*kafkav1.KafkaNode{node}}
	return &kafkav1.TopicDescription{Partitions: []*kafkav1.TopicPartitionInfo{tp}},
		assertEquals(topic, <-m.Expect)
}

func (m *MockKafka) CreateTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *MockKafka) DeleteTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *MockKafka) ListTopicConfig(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) (*kafkav1.TopicConfig, error) {
	return nil, assertEquals(topic, <-m.Expect)
}

func (m *MockKafka) UpdateTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {

	return assertEquals(topic, <-m.Expect)
}

func (m *MockKafka) ListACL(ctx context.Context, cluster *kafkav1.Cluster, filter *kafkav1.ACLFilter) ([]*kafkav1.ACLBinding, error) {
	return nil, assertEquals(filter, <-m.Expect)
}

func (m *MockKafka) CreateACL(ctx context.Context, cluster *kafkav1.Cluster, binding []*kafkav1.ACLBinding) error {
	return assertEquals(binding[0], <-m.Expect)
}

func (m *MockKafka) DeleteACL(ctx context.Context, cluster *kafkav1.Cluster, filter *kafkav1.ACLFilter) error {
	return assertEquals(filter, <-m.Expect)
}

func assertEquals(actual interface{}, expected interface{}) error {
	actualMessage := actual.(proto.Message)
	expectedMessage := expected.(proto.Message)

	if !proto.Equal(actualMessage, expectedMessage) {
		return fmt.Errorf("actual: %+v\nexpected: %+v", actual, expected)
	}
	return nil
}
