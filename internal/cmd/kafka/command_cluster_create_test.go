package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
)

func TestGetKafkaProvisionEstimate_Basic(t *testing.T) {
	expected := "It may take up to 5 minutes for the Kafka cluster to be ready."
	require.Equal(t, expected, getKafkaProvisionEstimate(ccstructs.Sku_BASIC))
}

func TestGetKafkaProvisionEstimate_Dedicated(t *testing.T) {
	expected := "It may take up to 1 hour for the Kafka cluster to be ready. The organization admin will receive an email once the dedicated cluster is provisioned."
	require.Equal(t, expected, getKafkaProvisionEstimate(ccstructs.Sku_DEDICATED))
}
