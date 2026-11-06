package logs

import (
	"encoding/json"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type sourceListCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type sourceOut struct {
	Name        string `human:"Name"`
	DisplayName string `human:"Display Name"`
	Resource    string `human:"Resource"`
	IdPrefix    string `human:"ID Prefix"`
	MaxLookback int32  `human:"Max Lookback Days"`
}

func newSourceListCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List log sources.",
		Long:  "List the log sources you can query. Results are always fetched fresh.",
		Args:  cobra.NoArgs,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the available log sources.",
				Code: "confluent logs source list",
			},
		),
	}

	c := &sourceListCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.list

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *sourceListCommand) list(cmd *cobra.Command, _ []string) error {
	sources, err := c.V2Client.ListLogSources()
	if err != nil {
		return err
	}
	// The list is always fetched fresh, and it refreshes what completion reuses.
	storeLogSources(c.AuthenticatedCLICommand, discoveryCacheKey(c.Context.GetPlatformServer(), c.Context.GetCurrentOrganization()), sources.Data, time.Now())

	if output.GetFormat(cmd).IsSerialized() {
		data := make([]json.RawMessage, 0, len(sources.Data))
		for _, source := range sources.Data {
			item, err := sourceJson(source)
			if err != nil {
				return err
			}
			data = append(data, item)
		}
		return output.SerializedOutputFromJsonTags(cmd, data)
	}

	list := output.NewList(cmd)
	for _, source := range sources.Data {
		list.Add(&sourceOut{
			Name:        source.Name,
			DisplayName: source.DisplayName,
			Resource:    source.ResourceType,
			IdPrefix:    source.GetIdPrefix(),
			MaxLookback: source.MaxLookbackDays,
		})
	}
	return list.Print()
}
