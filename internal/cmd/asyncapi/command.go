package asyncapi

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "asyncapi",
		Short:       "Manages asyncapi document tooling.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := pcmd.NewAnonymousCLICommand(cmd, prerunner)

	c.AddCommand(newExportCommand(prerunner))

	return c.Command
}
