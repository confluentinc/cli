package schema_registry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/log"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config   *config.Config
	logger   *log.Logger
	srClient *srsdk.APIClient
}

func New(prerunner pcmd.PreRunner, config *config.Config, srClient *srsdk.APIClient, logger *log.Logger) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "schema-registry",
			Short:             `Manage Schema Registry.`,
			PersistentPreRunE: prerunner.Authenticated(config),
		},
		config:   config,
		srClient: srClient,
		logger:   logger,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewClusterCommand(c.config, c.srClient, c.logger))
	c.AddCommand(NewSubjectCommand(c.config, c.srClient))
	c.AddCommand(NewSchemaCommand(c.config, c.srClient))
}
