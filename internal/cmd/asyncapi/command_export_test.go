package asyncapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/output"
)

func TestGetTopicDescription(t *testing.T) {
	detailsMock.channelDetails.currentTopic.TopicName = "topic1"
	err := detailsMock.getTopicDescription()
	require.NoError(t, err)
	require.Equal(t, "kafka topic description", detailsMock.channelDetails.currentTopicDescription)
}

func TestGetClusterDetails(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	flags := &flags{kafkaApiKey: ""}
	err = c.getClusterDetails(detailsMock, flags)
	require.NoError(t, err)
}

func TestGetSchemaRegistry(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	flags := &flags{schemaRegistryApiKey: "ASYNCAPIKEY", schemaRegistryApiSecret: "ASYNCAPISECRET"}
	err = c.getSchemaRegistry(detailsMock, flags)
	output.Println("")
	require.Error(t, err)
}

func TestGetSchemaDetails(t *testing.T) {
	detailsMock.channelDetails.currentSubject = "subject1"
	err := detailsMock.getSchemaDetails()
	require.NoError(t, err)
}

func TestGetChannelDetails(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)

	kafkaREST, err := c.GetKafkaREST()
	require.NoError(t, err)

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	require.NoError(t, err)

	topics, _, err := kafkaREST.CloudClient.ListKafkaTopics(kafkaClusterConfig.ID)
	require.NoError(t, err)

	detailsMock.topics = topics.Data
	detailsMock.channelDetails.currentSubject = "subject1"
	detailsMock.channelDetails.currentTopic = detailsMock.topics[0]
	schema, _, _ := detailsMock.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	detailsMock.channelDetails.schema = &schema
	flags := &flags{schemaRegistryApiKey: "ASYNCAPIKEY", schemaRegistryApiSecret: "ASYNCAPISECRET"}
	err = c.getChannelDetails(detailsMock, flags)
	require.NoError(t, err)
	//Protobuf Schema
	detailsMock.channelDetails.currentSubject = "subject2"
	detailsMock.channelDetails.currentTopic = detailsMock.topics[0]
	schema, _, _ = detailsMock.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject2", "1", nil)
	detailsMock.channelDetails.schema = &schema
	err = c.getChannelDetails(detailsMock, flags)
	require.Equal(t, err, fmt.Errorf("protobuf is not supported"))

}

func TestGetBindings(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	require.NoError(t, err)

	topics, _, err := detailsMock.kafkaRest.CloudClient.ListKafkaTopics(kafkaClusterConfig.ID)
	require.NoError(t, err)
	_, err = c.getBindings(detailsMock.kafkaRest, detailsMock.clusterId, topics.Data[0].TopicName)
	require.NoError(t, err)
}

func TestGetTags(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	schema, _, _ := detailsMock.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	detailsMock.srCluster = c.Config.Context().SchemaRegistryClusters["lsrc-asyncapi"]
	detailsMock.channelDetails.schema = &schema
	err = detailsMock.getTags()
	require.NoError(t, err)
}

func TestGetMessageCompatibility(t *testing.T) {
	_, err := getMessageCompatibility(detailsMock.srClient, *new(context.Context), "subject1")
	require.NoError(t, err)
}

func TestMsgName(t *testing.T) {
	require.Equal(t, "TopicNameMessage", msgName("topic name"))
}
