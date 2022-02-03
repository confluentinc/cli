package kafka

import (
	"testing"

	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	"github.com/stretchr/testify/require"
)

func TestGetKafkaProvisionEstimate_Basic(t *testing.T) {
	expected := "It may take up to 5 minutes for the Kafka cluster to be ready."
	require.Equal(t, expected, getKafkaProvisionEstimate(productv1.Sku_BASIC))
}

func TestGetKafkaProvisionEstimate_Dedicated(t *testing.T) {
	expected := "It may take up to 1 hour for the Kafka cluster to be ready. The admin will receive an email once the dedicated cluster is provisioned."
	require.Equal(t, expected, getKafkaProvisionEstimate(productv1.Sku_DEDICATED))
}
