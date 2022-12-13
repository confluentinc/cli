package pipeline

import (
	"time"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/spf13/cobra"
)

type Pipeline struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	KsqlCluster string    `json:"ksql_cluster"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var (
	pipelineListFields           = []string{"Id", "Name", "Description", "KsqlCluster", "State"}
	pipelineListHumanLabels      = []string{"ID", "Name", "Description", "KSQL Cluster", "State"}
	pipelineListStructuredLabels = []string{"id", "name", "description", "ksql_cluster", "state"}
	pipelineDescribeFields       = []string{"Id", "Name", "Description", "KsqlCluster", "State", "CreatedAt", "UpdatedAt"}
	pipelineDescribeHumanLabels  = map[string]string{
		"Id":          "ID",
		"Name":        "Name",
		"Description": "Description",
		"KsqlCluster": "KSQL Cluster",
		"State":       "State",
		"CreatedAt":   "Created At",
		"UpdatedAt":   "Updated At",
	}
	pipelineDescribeStructuredLabels = map[string]string{
		"Id":          "id",
		"Name":        "name",
		"Description": "description",
		"KsqlCluster": "ksql_cluster",
		"State":       "state",
		"CreatedAt":   "created_at",
		"UpdatedAt":   "updated_at",
	}
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

	c.AddCommand(c.newActivateCommand(prerunner))
	c.AddCommand(c.newCreateCommand(prerunner))
	c.AddCommand(c.newDeactivateCommand(prerunner))
	c.AddCommand(c.newDeleteCommand(prerunner))
	c.AddCommand(c.newDescribeCommand(prerunner))
	c.AddCommand(c.newListCommand(prerunner))
	c.AddCommand(c.newSaveCommand(prerunner))
	c.AddCommand(c.newUpdateCommand(prerunner))

	return c.Command
}
