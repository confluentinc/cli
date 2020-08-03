package mock

import (
	"context"
	"fmt"

	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"

	"github.com/golang/protobuf/proto"

	linkv1 "github.com/confluentinc/cc-structs/kafka/clusterlink/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
)

// Compile-time check interface adherence
var _ ccloud.Kafka = (*Kafka)(nil)

type Kafka struct {
	Expect chan interface{}
}

func (m *Kafka) Update(_ context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
	return cluster, nil
}

func (m *Kafka) GetTopicDefaults(_ context.Context, _ *schedv1.KafkaCluster) (*schedv1.TopicSpecification, error) {
	return &schedv1.TopicSpecification{}, nil
}

func (m *Kafka) GetTopicDefaultConfig(_ context.Context, _ *schedv1.KafkaCluster) (*schedv1.TopicConfig, error) {
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

func (m *Kafka) Create(_ context.Context, _ *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
	return &schedv1.KafkaCluster{Deployment: &schedv1.Deployment{Sku: productv1.Sku_BASIC}}, nil
}

func (m *Kafka) Delete(_ context.Context, _ *schedv1.KafkaCluster) error {
	return nil
}

func (m *Kafka) ListTopics(_ context.Context, _ *schedv1.KafkaCluster) ([]*schedv1.TopicDescription, error) {
	return []*schedv1.TopicDescription{
		{Name: "test1"},
		{Name: "test2"},
		{Name: "test3"}}, nil
}

func (m *Kafka) DescribeTopic(_ context.Context, _ *schedv1.KafkaCluster, topic *schedv1.Topic) (*schedv1.TopicDescription, error) {
	node := &schedv1.KafkaNode{Id: 1}
	tp := &schedv1.TopicPartitionInfo{Leader: node, Replicas: []*schedv1.KafkaNode{node}}
	return &schedv1.TopicDescription{Partitions: []*schedv1.TopicPartitionInfo{tp}},
		assertEquals(topic, <-m.Expect)
}

func (m *Kafka) CreateTopic(_ context.Context, _ *schedv1.KafkaCluster, topic *schedv1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) DeleteTopic(_ context.Context, _ *schedv1.KafkaCluster, topic *schedv1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) ListTopicConfig(_ context.Context, _ *schedv1.KafkaCluster, topic *schedv1.Topic) (*schedv1.TopicConfig, error) {
	return nil, assertEquals(topic, <-m.Expect)
}

func (m *Kafka) UpdateTopic(_ context.Context, _ *schedv1.KafkaCluster, topic *schedv1.Topic) error {
	return assertEquals(topic, <-m.Expect)
}

func (m *Kafka) ListACLs(_ context.Context, _ *schedv1.KafkaCluster, filter *schedv1.ACLFilter) ([]*schedv1.ACLBinding, error) {
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

func (m *Kafka) CreateACLs(_ context.Context, _ *schedv1.KafkaCluster, bindings []*schedv1.ACLBinding) error {
	return assertEqualBindings(bindings, <-m.Expect)
}

func (m *Kafka) DeleteACLs(_ context.Context, _ *schedv1.KafkaCluster, filters []*schedv1.ACLFilter) error {
	return assertEqualFilters(filters, <-m.Expect)
}

func (s *Kafka) ListLinks(_ context.Context, _ *schedv1.KafkaCluster) ([]string, error) {
	return nil, nil
}

func (s *Kafka) CreateLink(_ context.Context, _ *schedv1.KafkaCluster, _ string, _ *linkv1.LinkSourceCluster, _ *linkv1.CreateLinkOptions) error {
	return nil
}

func (s *Kafka) DeleteLink(_ context.Context, _ *schedv1.KafkaCluster, _ string, _ *linkv1.DeleteLinkOptions) error {
	return nil
}

func (s *Kafka) DescribeLink(_ context.Context, _ *schedv1.KafkaCluster, _ string) (*linkv1.LinkProperties, error) {
	return nil, nil
}

func (s *Kafka) AlterLink(_ context.Context, _ *schedv1.KafkaCluster, _ string, _ *linkv1.LinkProperties, _ *linkv1.AlterLinkOptions) error {
	return nil
}

func (m *Kafka) AlterMirror(_ context.Context, _ *schedv1.KafkaCluster, _ *schedv1.AlterMirrorOp) (*schedv1.AlterMirrorResult, error) {
	return nil, nil
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
