package pipeline

import (
	"time"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	launchdarkly "github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/spf13/cobra"
)

type Pipeline struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	KsqlCluster string    `json:"ksql_cluster"`
	SecretNames []string  `json:"secret_names,omitempty"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var (
	pipelineListFields           = []string{"Id", "Name", "Description", "KsqlCluster", "State"}
	pipelineListHumanLabels      = []string{"ID", "Name", "Description", "KSQL Cluster", "State"}
	pipelineListStructuredLabels = []string{"id", "name", "description", "ksql_cluster", "state"}
	pipelineDescribeFields       = []string{"Id", "Name", "Description", "KsqlCluster", "SecretNames", "State", "CreatedAt", "UpdatedAt"}
	pipelineDescribeHumanLabels  = map[string]string{
		"Id":          "ID",
		"Name":        "Name",
		"Description": "Description",
		"KsqlCluster": "KSQL Cluster",
		"SecretNames": "Secret Names",
		"State":       "State",
		"CreatedAt":   "Created At",
		"UpdatedAt":   "Updated At",
	}
	pipelineDescribeStructuredLabels = map[string]string{
		"Id":          "id",
		"Name":        "name",
		"Description": "description",
		"KsqlCluster": "ksql_cluster",
		"SecretNames": "secret_names",
		"State":       "state",
		"CreatedAt":   "created_at",
		"UpdatedAt":   "updated_at",
	}
	secretMappingWithoutEmptyValue = `^([a-zA-Z_][a-zA-Z0-9_]*)=(.+)$`
	secretMappingWithEmptyValue    = `^([a-zA-Z_][a-zA-Z0-9_]*)=(.*)$`
	secretNameSizeLimit            = 128
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

	c.AddCommand(c.newActivateCommand(prerunner))
	c.AddCommand(c.newCreateCommand(prerunner, enableSourceCode))
	c.AddCommand(c.newDeactivateCommand(prerunner))
	c.AddCommand(c.newDeleteCommand(prerunner))
	c.AddCommand(c.newDescribeCommand(prerunner))
	c.AddCommand(c.newListCommand(prerunner))
	c.AddCommand(c.newSaveCommand(prerunner, enableSourceCode))
	c.AddCommand(c.newUpdateCommand(prerunner, enableSourceCode))

	return c.Command
}
