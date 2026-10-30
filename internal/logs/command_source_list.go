package logs

import (
	"fmt"
	"strings"

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
	Scopes      string `human:"Scopes"`
	MaxLookback string `human:"Max Lookback"`
}

func newSourceListCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List log sources.",
		Long:  "List every log source that `confluent logs fetch` can query. Results are always fetched fresh.",
		Args:  cobra.NoArgs,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the available log sources:",
				Code: "confluent logs source list",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
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

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutputFromJsonTags(cmd, sources)
	}

	list := output.NewList(cmd)
	for _, source := range sources.Data {
		list.Add(&sourceOut{
			Name:        source.Name,
			DisplayName: source.DisplayName,
			Resource:    source.ResourceType,
			IdPrefix:    dashIfEmpty(source.GetIdPrefix()),
			Scopes:      formatScopes(source.AllowedScopes),
			MaxLookback: fmt.Sprintf("%dd", source.MaxLookbackDays),
		})
	}
	return list.Print()
}

// formatScopes renders allowed scopes the way users type them, e.g. "organization, environment, resource".
func formatScopes(scopes []string) string {
	lower := make([]string, len(scopes))
	for i, scope := range scopes {
		lower[i] = strings.ToLower(strings.ReplaceAll(scope, "_", "-"))
	}
	return strings.Join(lower, ", ")
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
