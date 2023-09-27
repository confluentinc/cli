package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetStringBroker(t *testing.T) {
	broker := getStringBroker("/kafka/v3/clusters/cluster-1/brokers/broker-1")
	require.Equal(t, "broker-1", broker)

	broker = getStringBroker("/kafka/v3/clusters/cluster-1/brokers/")
	require.Equal(t, "", broker)

	broker = getStringBroker("/kafka/v3/clusters/cluster-1")
	require.Equal(t, "", broker)
}
