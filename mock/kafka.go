package mock

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
)

// Compile-time check interface adherence
var _ ccloud.Kafka = (*Kafka)(nil)

type Kafka struct {
	Expect chan interface{}
}

func (m *Kafka) Update(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
	return cluster, nil
}

func (m *Kafka) GetTopicDefaults(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.TopicSpecification, error) {
	return &schedv1.TopicSpecification{}, nil
}

func (m *Kafka) GetTopicDefaultConfig(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.TopicConfig, error) {
	return &schedv1.TopicConfig{}, nil
}

func NewKafkaMock(expect chan interface{}) *Kafka {
	return &Kafka{expect}
}

func (m *Kafka) CreateAPIKey(_ context.Context, apiKey *schedv1.ApiKey) (*schedv1.ApiKey, error) {
	return apiKey, nil
}

func (m *Kafka) List(_ context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
	return []*schedv1.KafkaCluster{cluster}, nil
}

func (m *Kafka) Describe(_ context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
	return cluster, nil
}

func (m *Kafka) Create(_ context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
	return &schedv1.KafkaCluster{}, nil
}

func (m *Kafka) Delete(_ context.Context, cluster *schedv1.KafkaCluster) error {
	return nil
}

func (m *Kafka) ListTopics(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.TopicDescription, error) {
	return []*schedv1.TopicDescription{
		{Name: "test1"},
		{Name: "test2"},
		{Name: "test3"}}, nil
}

func (m *Kafka) DescribeTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *schedv1.Topic) (*schedv1.TopicDescription, error) {
	node := &schedv1.KafkaNode{Id: 1}
	tp := &schedv1.TopicPartitionInfo{Leader: node, Replicas: []*schedv1.KafkaNode{node}}
	return &schedv1.TopicDescription{Partitions: []*schedv1.TopicPartitionInfo{tp}},
		assertEquals(topic, <-m.Expect)
}

func (m *Kafka) CreateTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *schedv1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) DeleteTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *schedv1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) ListTopicConfig(ctx context.Context, cluster *schedv1.KafkaCluster, topic *schedv1.Topic) (*schedv1.TopicConfig, error) {
	return nil, assertEquals(topic, <-m.Expect)
}

func (m *Kafka) UpdateTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *schedv1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) ListACLs(ctx context.Context, cluster *schedv1.KafkaCluster, filter *schedv1.ACLFilter) ([]*schedv1.ACLBinding, error) {
	expect := <-m.Expect
	if filter.PatternFilter.PatternType == schedv1.PatternTypes_ANY {
		expect.(*schedv1.ACLFilter).PatternFilter.PatternType = schedv1.PatternTypes_ANY
	}
	if filter.EntryFilter.Operation == schedv1.ACLOperations_ANY {
		expect.(*schedv1.ACLFilter).EntryFilter.Operation = schedv1.ACLOperations_ANY
	}
	if filter.EntryFilter.PermissionType == schedv1.ACLPermissionTypes_ANY {
		expect.(*schedv1.ACLFilter).EntryFilter.PermissionType = schedv1.ACLPermissionTypes_ANY
	}
	return nil, assertEquals(filter, expect)
}

func (m *Kafka) CreateACLs(ctx context.Context, cluster *schedv1.KafkaCluster, bindings []*schedv1.ACLBinding) error {
	return assertEqualBindings(bindings, <-m.Expect)
}

func (m *Kafka) DeleteACLs(ctx context.Context, cluster *schedv1.KafkaCluster, filters []*schedv1.ACLFilter) error {
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

func assertEqualBindings(actual []*schedv1.ACLBinding, expected interface{}) error {
	exp := expected.([]*schedv1.ACLBinding)
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

func assertEqualFilters(actual []*schedv1.ACLFilter, expected interface{}) error {
	exp := expected.([]*schedv1.ACLFilter)
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
