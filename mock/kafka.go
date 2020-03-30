package mock

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
)

// Compile-time check interface adherence
var _ ccloud.Kafka = (*Kafka)(nil)

type Kafka struct {
	Expect chan interface{}
}

func (m *Kafka) GetTopicDefaults(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.TopicSpecification, error) {
	return &kafkav1.TopicSpecification{}, nil
}

func (m *Kafka) GetTopicDefaultConfig(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.TopicConfig, error) {
	return &kafkav1.TopicConfig{}, nil
}

func NewKafkaMock(expect chan interface{}) *Kafka {
	return &Kafka{expect}
}

func (m *Kafka) CreateAPIKey(_ context.Context, apiKey *authv1.ApiKey) (*authv1.ApiKey, error) {
	return apiKey, nil
}

func (m *Kafka) List(_ context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.KafkaCluster, error) {
	return []*kafkav1.KafkaCluster{cluster}, nil
}

func (m *Kafka) Describe(_ context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
	return cluster, nil
}

func (m *Kafka) Create(_ context.Context, config *kafkav1.KafkaClusterConfig) (*kafkav1.KafkaCluster, error) {
	return &kafkav1.KafkaCluster{}, nil
}

func (m *Kafka) Delete(_ context.Context, cluster *kafkav1.KafkaCluster) error {
	return nil
}

func (m *Kafka) ListTopics(ctx context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.TopicDescription, error) {
	return []*kafkav1.TopicDescription{
		{Name: "test1"},
		{Name: "test2"},
		{Name: "test3"}}, nil
}

func (m *Kafka) DescribeTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) (*kafkav1.TopicDescription, error) {
	node := &kafkav1.KafkaNode{Id: 1}
	tp := &kafkav1.TopicPartitionInfo{Leader: node, Replicas: []*kafkav1.KafkaNode{node}}
	return &kafkav1.TopicDescription{Partitions: []*kafkav1.TopicPartitionInfo{tp}},
		assertEquals(topic, <-m.Expect)
}

func (m *Kafka) CreateTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) DeleteTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) ListTopicConfig(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) (*kafkav1.TopicConfig, error) {
	return nil, assertEquals(topic, <-m.Expect)
}

func (m *Kafka) UpdateTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) ListACLs(ctx context.Context, cluster *kafkav1.KafkaCluster, filter *kafkav1.ACLFilter) ([]*kafkav1.ACLBinding, error) {
	expect := <-m.Expect
	if filter.PatternFilter.PatternType == kafkav1.PatternTypes_ANY {
		expect.(*kafkav1.ACLFilter).PatternFilter.PatternType = kafkav1.PatternTypes_ANY
	}
	if filter.EntryFilter.Operation == kafkav1.ACLOperations_ANY {
		expect.(*kafkav1.ACLFilter).EntryFilter.Operation = kafkav1.ACLOperations_ANY
	}
	if filter.EntryFilter.PermissionType == kafkav1.ACLPermissionTypes_ANY {
		expect.(*kafkav1.ACLFilter).EntryFilter.PermissionType = kafkav1.ACLPermissionTypes_ANY
	}
	return nil, assertEquals(filter, expect)
}

func (m *Kafka) CreateACLs(ctx context.Context, cluster *kafkav1.KafkaCluster, bindings []*kafkav1.ACLBinding) error {
	return assertEqualBindings(bindings, <-m.Expect)
}

func (m *Kafka) DeleteACLs(ctx context.Context, cluster *kafkav1.KafkaCluster, filters []*kafkav1.ACLFilter) error {
	return assertEqualFilters(filters, <-m.Expect)
}

func assertEquals(actual interface{}, expected interface{}) error {
	actualMessage := actual.(proto.Message)
	expectedMessage := expected.(proto.Message)

	if !proto.Equal(actualMessage, expectedMessage) {
		return fmt.Errorf("actual: %+v\nexpected: %+v", actual, expected)
	}
	return nil
}

func assertEqualBindings(actual []*kafkav1.ACLBinding, expected interface{}) error {
	exp := expected.([]*kafkav1.ACLBinding)
	if len(actual) != len(exp) {
		return fmt.Errorf("Length is not equal. actual: %d, expected: %d", len(actual), len(exp))
	}
	for i, actualMessage := range actual {
		if err := assertEquals(actualMessage, exp[i]); err != nil {
			return fmt.Errorf("Index %d is not equal. %s", i, err)
		}
	}
	return nil
}

func assertEqualFilters(actual []*kafkav1.ACLFilter, expected interface{}) error {
	exp := expected.([]*kafkav1.ACLFilter)
	if len(actual) != len(exp) {
		return fmt.Errorf("Length is not equal. actual: %d, expected: %d", len(actual), len(exp))
	}
	for i, actualMessage := range actual {
		if err := assertEquals(actualMessage, exp[i]); err != nil {
			return fmt.Errorf("Index %d is not equal. %s", i, err)
		}
	}
	return nil
}
