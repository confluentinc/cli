package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
	v3internal "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafkarest/v3"
)

func TestGetStringBrokerFromShareGroup(t *testing.T) {
	// Test with valid broker relationship
	shareGroup := v3internal.ShareGroupData{}
	coordinator := v3internal.Relationship{}
	coordinator.SetRelated("/kafka/v3/clusters/cluster-1/brokers/broker-1")
	shareGroup.SetCoordinator(coordinator)

	broker := getStringBrokerFromShareGroup(shareGroup)
	require.Equal(t, "broker-1", broker)

	// Test with empty relationship
	shareGroup2 := v3internal.ShareGroupData{}
	coordinator2 := v3internal.Relationship{}
	coordinator2.SetRelated("")
	shareGroup2.SetCoordinator(coordinator2)

	broker = getStringBrokerFromShareGroup(shareGroup2)
	require.Equal(t, "", broker)

	// Test with relationship that doesn't contain "brokers/"
	shareGroup3 := v3internal.ShareGroupData{}
	coordinator3 := v3internal.Relationship{}
	coordinator3.SetRelated("/kafka/v3/clusters/cluster-1")
	shareGroup3.SetCoordinator(coordinator3)

	broker = getStringBrokerFromShareGroup(shareGroup3)
	require.Equal(t, "", broker)

	// Test with relationship ending with "brokers/" but no broker ID
	shareGroup4 := v3internal.ShareGroupData{}
	coordinator4 := v3internal.Relationship{}
	coordinator4.SetRelated("/kafka/v3/clusters/cluster-1/brokers/")
	shareGroup4.SetCoordinator(coordinator4)

	broker = getStringBrokerFromShareGroup(shareGroup4)
	require.Equal(t, "", broker)

	// Test with invalid type (should return "N/A")
	broker = getStringBrokerFromShareGroup("invalid-type")
	require.Equal(t, "N/A", broker)
}

func TestFormatAssignedTopicPartitions(t *testing.T) {
	// Test with empty topic partitions
	topicPartitions := []v3internal.ShareGroupTopicPartitionData{}
	result := formatAssignedTopicPartitions(topicPartitions)
	require.Equal(t, "None", result)

	// Test with nil topic partitions
	result = formatAssignedTopicPartitions(nil)
	require.Equal(t, "None", result)

	// Test with single topic partition
	tp1 := v3internal.ShareGroupTopicPartitionData{}
	tp1.SetTopicName("topic1")
	tp1.SetPartitionId(0)
	topicPartitions = []v3internal.ShareGroupTopicPartitionData{tp1}

	result = formatAssignedTopicPartitions(topicPartitions)
	require.Equal(t, "topic1", result)

	// Test with multiple partitions of same topic (should return unique topic only)
	tp2 := v3internal.ShareGroupTopicPartitionData{}
	tp2.SetTopicName("topic1")
	tp2.SetPartitionId(1)
	tp3 := v3internal.ShareGroupTopicPartitionData{}
	tp3.SetTopicName("topic1")
	tp3.SetPartitionId(2)
	topicPartitions = []v3internal.ShareGroupTopicPartitionData{tp1, tp2, tp3}

	result = formatAssignedTopicPartitions(topicPartitions)
	require.Equal(t, "topic1", result)

	// Test with multiple different topics
	tp4 := v3internal.ShareGroupTopicPartitionData{}
	tp4.SetTopicName("topic2")
	tp4.SetPartitionId(0)
	tp5 := v3internal.ShareGroupTopicPartitionData{}
	tp5.SetTopicName("topic3")
	tp5.SetPartitionId(0)
	topicPartitions = []v3internal.ShareGroupTopicPartitionData{tp1, tp4, tp5}

	result = formatAssignedTopicPartitions(topicPartitions)
	// Since we use a map, the order is not guaranteed, so we check that all topics are present
	require.Contains(t, result, "topic1")
	require.Contains(t, result, "topic2")
	require.Contains(t, result, "topic3")
	require.Contains(t, result, ",") // Should contain commas as separators

	// Test with mixed scenario: multiple partitions of same topic and different topics
	tp6 := v3internal.ShareGroupTopicPartitionData{}
	tp6.SetTopicName("topic2")
	tp6.SetPartitionId(1)
	topicPartitions = []v3internal.ShareGroupTopicPartitionData{tp1, tp2, tp4, tp6, tp5}

	result = formatAssignedTopicPartitions(topicPartitions)
	// Should contain all unique topics: topic1, topic2, topic3
	require.Contains(t, result, "topic1")
	require.Contains(t, result, "topic2")
	require.Contains(t, result, "topic3")
	// Should not contain partition information
	require.NotContains(t, result, "-0")
	require.NotContains(t, result, "-1")
	require.NotContains(t, result, "-2")
}
