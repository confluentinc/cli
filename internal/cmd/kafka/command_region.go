package kafka

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)


type regionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewClusterCommand returns the Cobra command for Kafka cluster.
func NewRegionCommand(prerunner pcmd.PreRunner, config *v2.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "region",
			Short: "Cloud Regions.",
		},
		config, prerunner)
	cmd := &regionCommand{
		AuthenticatedCLICommand: cliCmd,
	}
	cmd.init()
	return cmd.Command
}

func (c *regionCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List cloud provider regions.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})
}

func (c *regionCommand) list(cmd *cobra.Command, args []string) error {
	clouds, err := c.Client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	for _, cloud := range clouds {
		fmt.Println(cloud.Name)
	}
	return nil
}
