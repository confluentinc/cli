package switchover

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/internal/switchover/endpoint"
	"github.com/confluentinc/cli/v4/internal/switchover/pair"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "switchover",
		Short:       "Manage Kafka disaster recovery switchover pairs and endpoints.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(pair.New(prerunner))
	cmd.AddCommand(endpoint.New(prerunner))

	return cmd
}
