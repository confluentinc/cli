package kafka

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/confluentinc/cli/v4/pkg/serdes"
)

func TestGetMessageString(t *testing.T) {
	message := &ckgo.Message{
		Value:          []byte("message"),
		TopicPartition: ckgo.TopicPartition{Offset: 2, Partition: 1},
		Timestamp:      time.Date(1997, time.July, 5, 0, 0, 0, 0, time.UTC),
	}
	valueDeserializer, err := serdes.GetDeserializationProvider("string")
	require.NoError(t, err)
	actual, err := getMessageString(message, valueDeserializer, ConsumerProperties{PrintOffset: true, Timestamp: true}, "doesn't matter")
	require.NoError(t, err)
	expected := "Timestamp:868060800000 Partition:1 Offset:2	message"
	require.Equal(t, expected, actual)
}
