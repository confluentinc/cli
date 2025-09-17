package kafka

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
)

func TestShareGroupConsumerOut(t *testing.T) {
	// Test the struct fields and tags
	consumer := &shareGroupConsumerOut{
		Cluster:    "test-cluster",
		ShareGroup: "test-share-group",
		Consumer:   "test-consumer",
		Client:     "test-client",
	}

	require.Equal(t, "test-cluster", consumer.Cluster)
	require.Equal(t, "test-share-group", consumer.ShareGroup)
	require.Equal(t, "test-consumer", consumer.Consumer)
	require.Equal(t, "test-client", consumer.Client)
}

func TestAddShareGroupFlag(t *testing.T) {
	// This test verifies that the addShareGroupFlag function sets up the flag correctly
	// We can't easily test the completion function without mocking, but we can verify
	// the basic structure

	// Create a mock command to test the flag addition
	cmd := &cobra.Command{}
	shareCmd := &shareCommand{}

	// This would normally be called in the command setup
	// We're testing that the function exists and can be called
	require.NotPanics(t, func() {
		shareCmd.addShareGroupFlag(cmd)
	})

	// Verify the flag was added
	groupFlag := cmd.Flags().Lookup("group")
	require.NotNil(t, groupFlag)
	require.Equal(t, "group", groupFlag.Name)
	require.Equal(t, "Share group ID.", groupFlag.Usage)
}

// Test helper function to create mock ShareGroupConsumerData
func createMockShareGroupConsumerData(clusterId, consumerId, clientId string) kafkarestv3.ShareGroupConsumerData {
	consumer := kafkarestv3.ShareGroupConsumerData{}
	consumer.SetClusterId(clusterId)
	consumer.SetConsumerId(consumerId)
	consumer.SetClientId(clientId)
	return consumer
}

func TestShareGroupConsumerDataMapping(t *testing.T) {
	// Test the mapping from SDK data to output struct
	mockConsumers := []kafkarestv3.ShareGroupConsumerData{
		createMockShareGroupConsumerData("cluster-1", "consumer-1", "client-1"),
		createMockShareGroupConsumerData("cluster-1", "consumer-2", "client-2"),
		createMockShareGroupConsumerData("cluster-1", "consumer-3", "client-3"),
	}

	groupName := "test-share-group"

	// Test mapping logic (simulating what happens in groupConsumerList)
	for i, consumer := range mockConsumers {
		output := &shareGroupConsumerOut{
			Cluster:    consumer.GetClusterId(),
			ShareGroup: groupName, // This comes from the command flag
			Consumer:   consumer.GetConsumerId(),
			Client:     consumer.GetClientId(),
		}

		require.Equal(t, "cluster-1", output.Cluster)
		require.Equal(t, groupName, output.ShareGroup)
		require.Equal(t, mockConsumers[i].GetConsumerId(), output.Consumer)
		require.Equal(t, mockConsumers[i].GetClientId(), output.Client)
	}
}

func TestShareGroupConsumerDataWithEmptyValues(t *testing.T) {
	// Test with empty/zero values
	consumer := createMockShareGroupConsumerData("", "", "")

	output := &shareGroupConsumerOut{
		Cluster:    consumer.GetClusterId(),
		ShareGroup: "test-group",
		Consumer:   consumer.GetConsumerId(),
		Client:     consumer.GetClientId(),
	}

	require.Equal(t, "", output.Cluster)
	require.Equal(t, "test-group", output.ShareGroup)
	require.Equal(t, "", output.Consumer)
	require.Equal(t, "", output.Client)
}

func TestShareGroupConsumerDataWithSpecialCharacters(t *testing.T) {
	// Test with special characters in IDs
	consumer := createMockShareGroupConsumerData("cluster-with-dashes", "consumer_with_underscores", "client.with.dots")

	output := &shareGroupConsumerOut{
		Cluster:    consumer.GetClusterId(),
		ShareGroup: "group-with-special-chars",
		Consumer:   consumer.GetConsumerId(),
		Client:     consumer.GetClientId(),
	}

	require.Equal(t, "cluster-with-dashes", output.Cluster)
	require.Equal(t, "group-with-special-chars", output.ShareGroup)
	require.Equal(t, "consumer_with_underscores", output.Consumer)
	require.Equal(t, "client.with.dots", output.Client)
}
