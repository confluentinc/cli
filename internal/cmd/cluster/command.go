package cluster

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type command struct {
	*pcmd.StateFlagCommand
}

func New(prerunner pcmd.PreRunner, userAgent string, logger *log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Retrieve metadata about Confluent Platform clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &command{StateFlagCommand: pcmd.NewAnonymousStateFlagCommand(cmd, prerunner, nil)}

	c.AddCommand(newDescribeCommand(prerunner, userAgent, logger))
	c.AddCommand(newListCommand(prerunner))
	c.AddCommand(newRegisterCommand(prerunner))
	c.AddCommand(newUnregisterCommand(prerunner))

	return c.Command
}
