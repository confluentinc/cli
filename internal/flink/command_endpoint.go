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
	IsCurrent bool   `human:"Current" serialized:"is_current,omitempty"`
	Endpoint  string `human:"Endpoint" serialized:"endpoint,omitempty"`
	Cloud     string `human:"Cloud" serialized:"cloud,omitempty"`
	Region    string `human:"Region" serialized:"region,omitempty"`
	Type      string `human:"Type" serialized:"type,omitempty"`
}

func (c *command) newEndpointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "endpoint",
		Short:       "Manage Flink endpoint.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newEndpointListCommand())
	cmd.AddCommand(c.newEndpointUseCommand())
	cmd.AddCommand(c.newEndpointUnsetCommand())

	return cmd
}
