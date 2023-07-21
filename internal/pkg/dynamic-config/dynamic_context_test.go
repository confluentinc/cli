package dynamicconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func TestFindKafkaCluster_Unexpired(t *testing.T) {
	update := time.Now()

	d := &DynamicContext{
		Context: &v1.Context{
			KafkaClusterContext: &v1.KafkaClusterContext{
				KafkaClusterConfigs: map[string]*v1.KafkaClusterConfig{
					"lkc-123456": {LastUpdate: update, Bootstrap: "pkc-abc12.us-west-2.aws.confluent.cloud:1234"},
				},
			},
		},
	}

	config, err := d.FindKafkaCluster("lkc-123456")
	require.NoError(t, err)
	require.True(t, config.LastUpdate.Equal(update))
}
