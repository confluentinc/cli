package cluster

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(prerunner pcmd.PreRunner, userAgent string) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Retrieve metadata about Confluent Platform clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.AddCommand(newDescribeCommand(prerunner, userAgent))
	cmd.AddCommand(newListCommand(prerunner))
	cmd.AddCommand(newRegisterCommand(prerunner))
	cmd.AddCommand(newUnregisterCommand(prerunner))

	return cmd
}
