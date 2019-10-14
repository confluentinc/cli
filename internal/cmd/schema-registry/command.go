package schema_registry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/log"

	ccsdk "github.com/confluentinc/ccloud-sdk-go"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config       *config.Config
	ccClient     ccsdk.SchemaRegistry
	metricClient ccsdk.Metrics
	srClient     *srsdk.APIClient
	ch           *pcmd.ContextResolver
	logger       *log.Logger
}

func New(prerunner pcmd.PreRunner, config *config.Config, ccloudClient ccsdk.SchemaRegistry, ch *pcmd.ContextResolver, srClient *srsdk.APIClient, metricClient ccsdk.Metrics, logger *log.Logger) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "schema-registry",
			Short:             `Manage Schema Registry.`,
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config:       config,
		ccClient:     ccloudClient,
		ch:           ch,
		srClient:     srClient,
		metricClient: metricClient,
		logger:       logger,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewClusterCommand(c.config, c.ccClient, c.ch, c.srClient, c.metricClient, c.logger))
	c.AddCommand(NewSubjectCommand(c.config, c.ch, c.srClient))
	c.AddCommand(NewSchemaCommand(c.config, c.ch, c.srClient))
}
