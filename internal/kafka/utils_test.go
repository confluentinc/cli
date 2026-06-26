package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"

	"github.com/confluentinc/cli/v4/pkg/serdes"
)

const mockSRUrl = "mock://"

func newMockSRClient(t *testing.T) schemaregistry.Client {
	t.Helper()
	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(mockSRUrl))
	require.NoError(t, err)
	return client
}

func seedAssociation(t *testing.T, client schemaregistry.Client, kafkaClusterId, mode, subject string) {
	t.Helper()
	const topic = "topic1"
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

func TestNewSchemaRegistryClient(t *testing.T) {
	tests := []struct {
		name string
		auth serdes.SchemaRegistryAuth
	}{
		{"basic authentication", serdes.SchemaRegistryAuth{ApiKey: "key", ApiSecret: "secret"}},
		{"bearer authentication", serdes.SchemaRegistryAuth{Token: "token"}},
		{"no authentication", serdes.SchemaRegistryAuth{}},
		{"with TLS paths", serdes.SchemaRegistryAuth{CertificateAuthorityPath: "ca", ClientCertPath: "cert", ClientKeyPath: "key"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := newSchemaRegistryClient(mockSRUrl, "lsrc-123", test.auth)
			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

func TestResolveProduceSubject(t *testing.T) {
	auth := serdes.SchemaRegistryAuth{}

	t.Run("empty kafkaClusterId falls back to TopicNameStrategy", func(t *testing.T) {
		require.Equal(t, "topic1-value", resolveProduceSubject(mockSRUrl, "lsrc-1", "", "topic1", "value", auth))
	})

	t.Run("empty srEndpoint falls back to TopicNameStrategy", func(t *testing.T) {
		require.Equal(t, "topic1-value", resolveProduceSubject("", "lsrc-1", "lkc-123", "topic1", "value", auth))
	})

	t.Run("no association falls back to TopicNameStrategy", func(t *testing.T) {
		require.Equal(t, "topic1-value", resolveProduceSubject(mockSRUrl, "lsrc-1", "lkc-123", "topic1", "value", auth))
	})

	t.Run("client build failure falls back to TopicNameStrategy", func(t *testing.T) {
		require.Equal(t, "topic1-value", resolveProduceSubject("://bad", "lsrc-1", "lkc-123", "topic1", "value", auth))
	})
}

func TestAssociatedValueSubject(t *testing.T) {
	t.Run("no association returns empty", func(t *testing.T) {
		client := newMockSRClient(t)
		require.Empty(t, associatedValueSubject(client, "lkc-123", "topic1"))
	})

	t.Run("matching value association returns its subject", func(t *testing.T) {
		client := newMockSRClient(t)
		seedAssociation(t, client, "lkc-123", "value", "custom-value-subject")
		require.Equal(t, "custom-value-subject", associatedValueSubject(client, "lkc-123", "topic1"))
	})
}

func TestResolveAssociatedValueSubject(t *testing.T) {
	auth := serdes.SchemaRegistryAuth{}

	t.Run("non-protobuf format returns empty", func(t *testing.T) {
		require.Empty(t, resolveAssociatedValueSubject("avro", mockSRUrl, "lsrc-1", "lkc-123", "topic1", auth))
	})

	t.Run("empty kafkaClusterId returns empty", func(t *testing.T) {
		require.Empty(t, resolveAssociatedValueSubject("protobuf", mockSRUrl, "lsrc-1", "", "topic1", auth))
	})

	t.Run("empty srEndpoint returns empty", func(t *testing.T) {
		require.Empty(t, resolveAssociatedValueSubject("protobuf", "", "lsrc-1", "lkc-123", "topic1", auth))
	})

	t.Run("protobuf with no association returns empty", func(t *testing.T) {
		require.Empty(t, resolveAssociatedValueSubject("protobuf", mockSRUrl, "lsrc-1", "lkc-123", "topic1", auth))
	})

	t.Run("client build failure returns empty", func(t *testing.T) {
		require.Empty(t, resolveAssociatedValueSubject("protobuf", "://bad", "lsrc-1", "lkc-123", "topic1", auth))
	})
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
		seedAssociation(t, client, "lkc-123", "value", "custom-value-subject")
		require.Equal(t, "custom-value-subject", resolveSubject(client, "lkc-123", "topic1", "value"))
	})

	t.Run("association for other mode falls back", func(t *testing.T) {
		client := newMockSRClient(t)
		seedAssociation(t, client, "lkc-123", "key", "custom-key-subject")
		require.Equal(t, "topic1-value", resolveSubject(client, "lkc-123", "topic1", "value"))
	})

	t.Run("association under different cluster id falls back", func(t *testing.T) {
		client := newMockSRClient(t)
		seedAssociation(t, client, "lkc-other", "value", "should-not-be-used")
		require.Equal(t, "topic1-value", resolveSubject(client, "lkc-123", "topic1", "value"))
	})

	t.Run("association lookup error falls back to TopicNameStrategy", func(t *testing.T) {
		client := newMockSRClient(t)
		// An empty topic makes the associations lookup return an error, exercising the error path.
		require.Equal(t, "-value", resolveSubject(client, "lkc-123", "", "value"))
	})
}
