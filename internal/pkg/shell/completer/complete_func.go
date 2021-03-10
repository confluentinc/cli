package completer

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"

	"github.com/confluentinc/ccloud-sdk-go"
	pkafka "github.com/confluentinc/cli/internal/pkg/kafka"
)

func ClusterFlagServerCompleterFunc(client *ccloud.Client, environmentId string) func() []prompt.Suggest {
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

func ServiceAccountFlagCompleterFunc(client *ccloud.Client) func() []prompt.Suggest {
	return func() []prompt.Suggest {
		var suggestions []prompt.Suggest
		users, err := client.User.GetServiceAccounts(context.Background())
		if err != nil {
			return suggestions
		}
		for _, user := range users {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        fmt.Sprintf("%d", user.Id),
				Description: user.ServiceName,
			})
		}
		return suggestions
	}
}
