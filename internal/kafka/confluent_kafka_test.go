package kafka

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v3/pkg/serdes"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

func TestGetMessageString(t *testing.T) {
	message := &ckafka.Message{
		Value:          []byte("message"),
		TopicPartition: ckafka.TopicPartition{Offset: 2, Partition: 1},
		Timestamp:      time.Date(1997, time.July, 5, 0, 0, 0, 0, time.UTC),
	}
	valueDeserializer, err := serdes.GetDeserializationProvider("string")
	require.NoError(t, err)
	actual, err := getMessageString(message, valueDeserializer, ConsumerProperties{PrintOffset: true, Timestamp: true})
	require.NoError(t, err)
	expected := "Timestamp:868060800000 Partition:1 Offset:2	message"
	require.Equal(t, expected, actual)
}
