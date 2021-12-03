package ksql

import (
	"context"

	"github.com/c-bata/go-prompt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/shell/completer"
)

func (c *appCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *appCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.KSQL.List(context.Background(), req)
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

func (c *appCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *appCommand) ServerCompletableFlagChildren() map[string][]*cobra.Command {
	return c.completableFlagChildren
}

func (c *appCommand) ServerFlagComplete() map[string]func() []prompt.Suggest {
	return map[string]func() []prompt.Suggest{
		"cluster": completer.ClusterFlagServerCompleterFunc(c.Client, c.EnvironmentId()),
	}
}
