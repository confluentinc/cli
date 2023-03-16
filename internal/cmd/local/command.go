package local

import (
	"runtime"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const imageName = "523370736235.dkr.ecr.us-west-2.amazonaws.com/confluentinc/kafka-local:latest"
const testTopicName = "jsontest"

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local",
		Short: "Manage a local Confluent Platform development environment.",
		Long:  "Use the \"confluent local\" commands to try out Confluent Platform by running a single-node instance locally on your machine. Keep in mind, these commands require Java to run.",
		Args:  cobra.NoArgs,
	}

	if runtime.GOOS == "windows" {
		cmd.Hidden = true
	}

	cmd.AddCommand(NewKafkaCommand(prerunner))

	return cmd
}
