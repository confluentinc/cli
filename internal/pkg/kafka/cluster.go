package kafka

import (
	"context"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func ClusterFlagServerCompleterFunc(cmd *cobra.Command, client *ccloud.Client, environmentId string) func() []prompt.Suggest {
	return func() []prompt.Suggest {
		var suggestions []prompt.Suggest
		if !pcmd.CanCompleteCommand(cmd) {
			return suggestions
		}
		clusters, err := ListKafkaClusters(client, environmentId)
		if err != nil {
			return suggestions
		}
		for _, cluster := range clusters {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        cluster.Id,
				Description: cluster.Name,
			})
		}
		return suggestions
	}
}

func ListKafkaClusters(client *ccloud.Client, environmentId string) ([]*schedv1.KafkaCluster, error) {
	req := &schedv1.KafkaCluster{AccountId: environmentId}
	return client.Kafka.List(context.Background(), req)
}
