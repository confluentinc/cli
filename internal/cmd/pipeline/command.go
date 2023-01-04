package pipeline

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	launchdarkly "github.com/confluentinc/cli/internal/pkg/featureflags"
)

type out struct {
	Id          string    `human:"ID" serialized:"id"`
	Name        string    `human:"Name" serialized:"name"`
	Description string    `human:"Description" serialized:"description"`
	KsqlCluster string    `human:"KSQL Cluster" serialized:"ksql_cluster"`
	State       string    `human:"State" serialized:"state"`
	CreatedAt   time.Time `human:"Created At" serialized:"created_at"`
	UpdatedAt   time.Time `human:"Updated At" serialized:"updated_at"`
}

var (
	secretMappingWithoutEmptyValue = `^([a-zA-Z_][a-zA-Z0-9_]*)=(.+)$`
	secretMappingWithEmptyValue    = `^([a-zA-Z_][a-zA-Z0-9_]*)=(.*)$`
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "pipeline",
		Short:       "Manage Stream Designer pipelines.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
		prerunner:                     prerunner,
	}

	dc := dynamicconfig.New(cfg, nil, nil, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)
	enableSourceCode := launchdarkly.Manager.BoolVariation("cli.stream_designer.source_code.enable", dc.Context(), v1.CliLaunchDarklyClient, true, false)

	c.AddCommand(c.newActivateCommand())
	c.AddCommand(c.newCreateCommand(enableSourceCode))
	c.AddCommand(c.newDeactivateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newSaveCommand(enableSourceCode))
	c.AddCommand(c.newUpdateCommand(enableSourceCode))

	return c.Command
}
