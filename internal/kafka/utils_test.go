package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
)

const mockSRUrl = "mock://"

func newMockSRClient(t *testing.T) schemaregistry.Client {
	t.Helper()
	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(mockSRUrl))
	require.NoError(t, err)
	return client
}

func seedAssociation(t *testing.T, client schemaregistry.Client, topic, kafkaClusterId, mode, subject string) {
	t.Helper()
	// The mock requires the subject to have a registered schema before it
	// will accept an association referencing it.
	_, err := client.Register(subject, schemaregistry.SchemaInfo{
		Schema:     `{"type":"record","name":"R","fields":[{"name":"f","type":"int"}]}`,
		SchemaType: "AVRO",
	}, false)
	require.NoError(t, err)
	_, err = client.CreateOrUpdateAssociation(schemaregistry.AssociationCreateOrUpdateRequest{
		ResourceName:      topic,
		ResourceNamespace: kafkaClusterId,
		ResourceID:        topic + ":" + kafkaClusterId,
		ResourceType:      "topic",
		Associations: []schemaregistry.AssociationCreateOrUpdateInfo{{
			Subject:         subject,
			AssociationType: mode,
		}},
	})
	require.NoError(t, err)
}

func TestResolveSubject(t *testing.T) {
	t.Run("nil client falls back to TopicNameStrategy", func(t *testing.T) {
		require.Equal(t, "topic1-value", resolveSubject(nil, "lkc-123", "topic1", "value"))
	})

	t.Run("empty kafkaClusterId falls back to TopicNameStrategy", func(t *testing.T) {
		client := newMockSRClient(t)
		require.Equal(t, "topic1-value", resolveSubject(client, "", "topic1", "value"))
	})

	t.Run("no association falls back to TopicNameStrategy", func(t *testing.T) {
		client := newMockSRClient(t)
		require.Equal(t, "topic1-value", resolveSubject(client, "lkc-123", "topic1", "value"))
	})

	t.Run("matching association returns its subject", func(t *testing.T) {
		client := newMockSRClient(t)
		seedAssociation(t, client, "topic1", "lkc-123", "value", "custom-value-subject")
		require.Equal(t, "custom-value-subject", resolveSubject(client, "lkc-123", "topic1", "value"))
	})

	t.Run("association for other mode falls back", func(t *testing.T) {
		client := newMockSRClient(t)
		seedAssociation(t, client, "topic1", "lkc-123", "key", "custom-key-subject")
		require.Equal(t, "topic1-value", resolveSubject(client, "lkc-123", "topic1", "value"))
	})

	t.Run("association under different cluster id falls back", func(t *testing.T) {
		client := newMockSRClient(t)
		seedAssociation(t, client, "topic1", "lkc-other", "value", "should-not-be-used")
		require.Equal(t, "topic1-value", resolveSubject(client, "lkc-123", "topic1", "value"))
	})
}
