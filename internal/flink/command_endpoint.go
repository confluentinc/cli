package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

const (
	publicFlinkEndpointType  = "PUBLIC"
	privateFlinkEndpointType = "PRIVATE"
)

type flinkEndpointOut struct {
	Endpoint string `human:"Endpoint" serialized:"endpoint"`
	Cloud    string `human:"Cloud" serialized:"cloud"`
	Region   string `human:"Region" serialized:"region"`
	Type     string `human:"Type" serialized:"type"`
}

func (c *command) newEndpointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "endpoint",
		Short:       "Manage Flink endpoint.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newEndpointListCommand())
	cmd.AddCommand(c.newEndpointUseCommand())

	return cmd
}
