package cluster

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/spf13/cobra"
)

type command struct {
	*pcmd.CLICommand
	prerunner  pcmd.PreRunner
	config     *v3.Config
	metaClient Metadata
}

// New returns the Cobra command for `cluster`.
func New(prerunner pcmd.PreRunner, config *v3.Config, metaClient Metadata) *cobra.Command {
	cmd := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(&cobra.Command{
			Use:   "cluster",
			Short: "Retrieve metadata about Confluent clusters.",
		}, config, prerunner),
		prerunner:  prerunner,
		config:     config,
		metaClient: metaClient,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewDescribeCommand(c.config, c.prerunner, c.metaClient))
	c.AddCommand(NewListCommand(c.config, c.prerunner))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
