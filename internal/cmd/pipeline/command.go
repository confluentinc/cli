package pipeline

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
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
	c.AddCommand(c.newUpdateCommand(prerunner))

	return c.Command
}
