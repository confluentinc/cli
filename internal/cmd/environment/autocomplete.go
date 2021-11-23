package environment

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"
)

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return suggestions
	}
	for _, env := range environments {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        env.Id,
			Description: env.Name,
		})
	}
	return suggestions
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return autocompleteEnvironments(c.Client)
}

func autocompleteEnvironments(client *ccloud.Client) []string {
	environments, err := client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(environments))
	for i, environment := range environments {
		suggestions[i] = fmt.Sprintf("%s\t%s", environment.Id, environment.Name)
	}
	return suggestions
}
