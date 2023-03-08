package asyncapi

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/confluentinc/cli/internal/pkg/output"
)

func TestGetTopicDescription(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)

	kafkaREST, err := c.GetKafkaREST()
	require.NoError(t, err)

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	require.NoError(t, err)

	topics, _, err := kafkaREST.CloudClient.ListKafkaTopics(kafkaClusterConfig.ID)
	require.NoError(t, err)

	details.topics = topics.Data
	details.channelDetails.currentSubject = "subject1"
	details.channelDetails.currentTopic = details.topics[0]
	err = details.getTopicDescription()
	require.NoError(t, err)
	require.Equal(t, "kafka topic description", details.channelDetails.currentTopicDescription)
}

func TestGetClusterDetails(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	flags := &flags{kafkaApiKey: ""}
	err = c.getClusterDetails(details, flags)
	require.NoError(t, err)
}

func TestGetSchemaRegistry(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	flags := &flags{schemaRegistryApiKey: "ASYNCAPIKEY", schemaRegistryApiSecret: "ASYNCAPISECRET"}
	err = c.getSchemaRegistry(details, flags)
	output.Println("")
	require.Error(t, err)
}

func TestGetSchemaDetails(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)

	kafkaREST, err := c.GetKafkaREST()
	require.NoError(t, err)

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	require.NoError(t, err)

	topics, _, err := kafkaREST.CloudClient.ListKafkaTopics(kafkaClusterConfig.ID)
	require.NoError(t, err)

	details.topics = topics.Data
	details.channelDetails.currentSubject = "subject1"
	details.channelDetails.currentTopic = details.topics[0]
	schema, _, _ := details.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	details.channelDetails.schema = &schema
	err = details.getSchemaDetails()
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

	details.topics = topics.Data
	details.channelDetails.currentSubject = "subject1"
	details.channelDetails.currentTopic = details.topics[0]
	schema, _, _ := details.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	details.channelDetails.schema = &schema
	flags := &flags{schemaRegistryApiKey: "ASYNCAPIKEY", schemaRegistryApiSecret: "ASYNCAPISECRET"}
	err = c.getChannelDetails(details, flags)
	require.NoError(t, err)
	//Protobuf Schema
	details.channelDetails.currentSubject = "subject2"
	details.channelDetails.currentTopic = details.topics[0]
	schema, _, _ = details.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject2", "1", nil)
	details.channelDetails.schema = &schema
	err = c.getChannelDetails(details, flags)
	require.Error(t, err, fmt.Errorf("protobuf"))

}

func TestGetBindings(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)

	kafkaREST, err := c.GetKafkaREST()
	require.NoError(t, err)

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	require.NoError(t, err)

	topics, _, err := kafkaREST.CloudClient.ListKafkaTopics(kafkaClusterConfig.ID)
	require.NoError(t, err)

	_, err = c.getBindings(details.cluster, topics.Data[0])
	require.NoError(t, err)
}

func TestGetTags(t *testing.T) {
	c, err := newCmd()
	require.NoError(t, err)
	schema, _, _ := details.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	details.srCluster = c.Config.Context().SchemaRegistryClusters["lsrc-asyncapi"]
	details.channelDetails.schema = &schema
	err = details.getTags()
	require.NoError(t, err)
}

func TestGetMessageCompatibility(t *testing.T) {
	_, err := getMessageCompatibility(details.srClient, *new(context.Context), "subject1")
	require.NoError(t, err)
}

func TestMsgName(t *testing.T) {
	require.Equal(t, "TopicNameMessage", msgName("topic name"))
}
