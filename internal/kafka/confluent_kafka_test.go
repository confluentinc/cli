package kafka

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/confluentinc/cli/v4/pkg/serdes"
)

func TestConsumeMessage_ValueSubjectFallback(t *testing.T) {
	tests := []struct {
		name         string
		valueSubject string
	}{
		{"empty ValueSubject falls back to Subject", ""},
		{"explicit ValueSubject is used", "topic1-value-context"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := &ckgo.Message{
				Key:            []byte("my-key"),
				Value:          []byte("my-value"),
				TopicPartition: ckgo.TopicPartition{Offset: 1, Partition: 0},
			}
			var buf bytes.Buffer
			handler := &GroupHandler{
				KeyFormat:      "string",
				ValueFormat:    "string",
				Subject:        "topic1-value",
				ValueSubject:   test.valueSubject,
				KafkaClusterId: "lkc-123",
				Topic:          "topic1",
				Out:            &buf,
				Properties:     ConsumerProperties{PrintKey: true, Delimiter: "\t"},
			}
			require.NoError(t, ConsumeMessage(message, handler))
			require.Contains(t, buf.String(), "my-key")
			require.Contains(t, buf.String(), "my-value")
		})
	}
}

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
