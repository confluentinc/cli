package kafka

import (
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type brokerCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newBrokerCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "broker",
		Short:       "Manage Kafka brokers.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &brokerCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand))

	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newGetTasksCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}

func checkAllOrBrokerIdSpecified(cmd *cobra.Command, args []string) (int32, bool, error) {
	if cmd.Flags().Changed("all") && len(args) > 0 {
		return -1, false, errors.New(errors.OnlySpecifyAllOrBrokerIDErrorMsg)
	}
	if !cmd.Flags().Changed("all") && len(args) == 0 {
		return -1, false, errors.New(errors.MustSpecifyAllOrBrokerIDErrorMsg)
	}
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return -1, false, err
	}
	if len(args) > 0 {
		brokerIdStr := args[0]
		brokerId, err := strconv.ParseInt(brokerIdStr, 10, 32)
		return int32(brokerId), false, err
	}
	return -1, all, nil
}
