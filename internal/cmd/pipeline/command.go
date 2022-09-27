package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
)

type Pipeline struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

var (
	pipelineListFields           = []string{"Id", "Name", "State"}
	pipelineListHumanLabels      = []string{"ID", "Name", "State"}
	pipelineListStructuredLabels = []string{"id", "name", "state"}
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "pipeline",
		Short:       "Manage Stream Designer pipelines.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		prerunner:               prerunner,
	}

	c.AddCommand(c.newActivateCommand(prerunner))
	c.AddCommand(c.newCreateCommand(prerunner))
	c.AddCommand(c.newDeactivateCommand(prerunner))
	c.AddCommand(c.newDeleteCommand(prerunner))
	c.AddCommand(c.newDescribeCommand(prerunner))
	c.AddCommand(c.newListCommand(prerunner))
	c.AddCommand(c.newUpdateCommand(prerunner))

	dc := dynamicconfig.New(cfg, nil, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)

	return c.Command
}
