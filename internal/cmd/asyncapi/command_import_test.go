package asyncapi

import (
	"testing"

	v3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/stretchr/testify/require"
	spec2 "github.com/swaggest/go-asyncapi/spec-2.4.0"
)

func TestAddTopic(t *testing.T) {
	kb := KafkaBinding{
		XPartitions: 1,
		XReplicas:   3,
		XConfigs: BindingsXConfigs{
			CleanupPolicy:     "delete",
			DeleteRetentionMs: 8.64e+07,
		},
	}
	//If topic does not exist
	_, err := addTopic(details, "testTopic", kb, false)
	require.NoError(t, err)
	//If topic exists & overwrite is false
	details.topics = append(details.topics, v3.TopicData{TopicName: "testTopic"})
	_, err = addTopic(details, "testTopic", kb, false)
	require.NoError(t, err)
}

func TestResolveSchemaType(t *testing.T) {
	require.Equal(t, resolveSchemaType("avro"), "AVRO")
	require.Equal(t, resolveSchemaType("json"), "JSON")
	require.Equal(t, resolveSchemaType("proto"), "PROTOBUF")
}

func TestRegisterSchema(t *testing.T) {
	components := Components{
		Messages: map[string]Message{
			"TestTopicMessage": Message{
				ContentType:  "application/json",
				SchemaFormat: "application/schema+json;version=draft-07",
				Payload:      `{"type": "string"}`,
			},
		},
	}
	id, err := registerSchema(details, "testTopic", components)
	require.NoError(t, err)
	require.Equal(t, int32(100001), id)
}

func TestUpdateSubjectCompatibility(t *testing.T) {
	err := updateSubjectCompatibility(details, "BACKWARD", "testTopic-value")
	require.NoError(t, err)
	err = updateSubjectCompatibility(details, "INVALID", "testTopic-value")
	require.Error(t, err)
}

func TestAddSchemaTags(t *testing.T) {
	components := Components{
		Messages: map[string]Message{
			"TestTopicMessage": Message{
				Tags: []spec2.Tag{
					{
						Name: "Tag1",
					},
					{
						Name: "Tag2",
					},
				},
			},
		},
	}
	tags, tagDefs, err := addSchemaTags(details, components, "testTopic", int32(100001))
	require.NoError(t, err)
	require.Contains(t, tags[0].EntityName, "lsrc-asyncapi:.:100001")
	require.Contains(t, tags[0].EntityType, "sr_schema")
	require.Contains(t, tags[0].TypeName, "Tag1")
	require.Contains(t, tags[1].TypeName, "Tag2")
	require.Equal(t, tagDefs[0].EntityTypes, []string{"cf_entity"})
}

func TestAddTopicTags(t *testing.T) {
	subscribe := Operation{
		TopicTags: []spec2.Tag{
			{
				Name: "Tag1",
			},
			{
				Name: "Tag2",
			},
		}}
	tags, tagDefs, err := addTopicTags(details, subscribe, "testTopic")
	require.NoError(t, err)
	require.NoError(t, err)
	require.Contains(t, tags[0].EntityName, "lsrc-asyncapi:lkc-asyncapi:testTopic")
	require.Contains(t, tags[0].EntityType, "kafka_topic")
	require.Contains(t, tags[0].TypeName, "Tag1")
	require.Contains(t, tags[1].TypeName, "Tag2")
	require.Equal(t, tagDefs[0].EntityTypes, []string{"cf_entity"})
}
