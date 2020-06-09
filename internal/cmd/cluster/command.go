package cluster

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

type command struct {
	*pcmd.CLICommand
	prerunner  pcmd.PreRunner
	config     *v3.Config
	metaClient Metadata
}

// New returns the Cobra command for `cluster`.
func New(prerunner pcmd.PreRunner, metaClient Metadata) *cobra.Command {
	cmd := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(&cobra.Command{
			Use:   "cluster",
			Short: "Retrieve metadata about Confluent clusters.",
		}, prerunner),
		prerunner:  prerunner,
		metaClient: metaClient,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewDescribeCommand(c.prerunner, c.metaClient))
	c.AddCommand(NewListCommand(c.config, c.prerunner))
	c.AddCommand(NewRegisterCommand(c.config, c.prerunner))
	c.AddCommand(NewUnregisterCommand(c.config, c.prerunner))
}
