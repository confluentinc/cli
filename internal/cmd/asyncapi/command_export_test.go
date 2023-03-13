package asyncapi

import (
	"context"
	"fmt"
	"testing"

	v3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
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
	c := mockAsyncApiCommand()
	flags := &flags{kafkaApiKey: ""}
	err := c.getClusterDetails(detailsMock, flags)
	require.NoError(t, err)
}

func TestGetSchemaRegistry(t *testing.T) {
	c := mockAsyncApiCommand()
	flags := &flags{schemaRegistryApiKey: "ASYNCAPIKEY", schemaRegistryApiSecret: "ASYNCAPISECRET"}
	err := c.getSchemaRegistry(detailsMock, flags)
	output.Println("")
	require.Error(t, err)
}

func TestGetSchemaDetails(t *testing.T) {
	detailsMock.channelDetails.currentSubject = "subject1"
	err := detailsMock.getSchemaDetails()
	require.NoError(t, err)
}

func TestGetChannelDetails(t *testing.T) {
	c := mockAsyncApiCommand()
	topicsData := []v3.TopicData{
		{TopicName: "topic1"},
	}
	detailsMock.topics = topicsData
	detailsMock.channelDetails.currentSubject = "subject1"
	detailsMock.channelDetails.currentTopic = detailsMock.topics[0]
	detailsMock.kafkaRest, _ = c.GetKafkaREST()
	schema, _, _ := detailsMock.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	detailsMock.channelDetails.schema = &schema
	flags := &flags{schemaRegistryApiKey: "ASYNCAPIKEY", schemaRegistryApiSecret: "ASYNCAPISECRET"}
	err := c.getChannelDetails(detailsMock, flags)
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
	c := mockAsyncApiCommand()
	detailsMock.kafkaRest, _ = c.GetKafkaREST()
	_, err := c.getBindings(detailsMock.kafkaRest, detailsMock.clusterId, "topic1")
	require.NoError(t, err)
}

func TestGetTags(t *testing.T) {
	schema, _, _ := detailsMock.srClient.DefaultApi.GetSchemaByVersion(*new(context.Context), "subject1", "1", nil)
	detailsMock.channelDetails.schema = &schema
	err := detailsMock.getTags()
	require.NoError(t, err)
}

func TestGetMessageCompatibility(t *testing.T) {
	_, err := getMessageCompatibility(detailsMock.srClient, *new(context.Context), "subject1")
	require.NoError(t, err)
}

func TestMsgName(t *testing.T) {
	require.Equal(t, "TopicNameMessage", msgName("topic name"))
}
