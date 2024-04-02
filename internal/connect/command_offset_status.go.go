package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type alterStatusOut struct {
	Id        string `human:"ID" serialized:"id"`
	Phase     string `human:"Phase" serialized:"phase"`
	Message   string `human:"Message,omitempty" serialized:"message,omitempty"`
	AppliedAt string `human:"Applied at,omitempty" serialized:"applied_at,omitempty"`
}

func (c *offsetCommand) newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "status",
		Short:             "Get connector offset alter status",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.alterStatus,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get connector offset alter status for a lccId in the current or specified Kafka cluster context.",
				Code: "confluent connect offset status lcc-123456",
			},
			examples.Example{
				Code: "confluent connect offset status lcc-123456 --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}
func (c *offsetCommand) alterStatus(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connectorIdToName, err := c.mapConnectorIdToName(environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	offsetStatus, err := c.V2Client.AlterConnectorOffsetsRequestStatus(connectorIdToName[args[0]], environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	var appliedAt string
	if offsetStatus.AppliedAt.IsSet() {
		appliedAt = offsetStatus.AppliedAt.Get().String()
	}
	var phase string
	var message string
	_, isStatusSet := offsetStatus.GetStatusOk()
	if isStatusSet {
		phase = offsetStatus.GetStatus().Phase
		if messagePtr := offsetStatus.GetStatus().Message; messagePtr != nil {
			message = *messagePtr
		}
	}
	table := output.NewTable(cmd)
	table.Add(&alterStatusOut{
		Id:        args[0],
		Phase:     phase,
		Message:   message,
		AppliedAt: appliedAt,
	})
	return table.Print()
}
