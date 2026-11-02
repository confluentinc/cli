package logs

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

// completeSources suggests the names of the available log sources, with their display names as descriptions.
func completeSources(c *pcmd.AuthenticatedCLICommand, cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	sources, err := c.V2Client.ListLogSources()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(sources.Data))
	for i, source := range sources.Data {
		suggestions[i] = fmt.Sprintf("%s\t%s", source.Name, source.DisplayName)
	}
	return suggestions
}
