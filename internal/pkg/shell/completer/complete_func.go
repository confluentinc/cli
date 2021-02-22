package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pkafka "github.com/confluentinc/cli/internal/pkg/kafka"
)

func ClusterFlagServerCompleterFunc(cmd *cobra.Command, client *ccloud.Client, environmentId string) func() []prompt.Suggest {
	return func() []prompt.Suggest {
		var suggestions []prompt.Suggest
		clusters, err := pkafka.ListKafkaClusters(client, environmentId)
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
