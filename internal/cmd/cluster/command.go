package cluster

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.StateFlagCommand
}

func New(prerunner pcmd.PreRunner, userAgent string) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Retrieve metadata about Confluent Platform clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &command{pcmd.NewAnonymousStateFlagCommand(cmd, prerunner)}

	c.AddCommand(newDescribeCommand(prerunner, userAgent))
	c.AddCommand(newListCommand(prerunner))
	c.AddCommand(newRegisterCommand(prerunner))
	c.AddCommand(newUnregisterCommand(prerunner))

	return c.Command
}
