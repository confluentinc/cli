package local

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

const (
	dockerImageName             = "confluentinc/confluent-local:latest"
	localhostPrefix             = "http://localhost:%s"
	localhost                   = "0.0.0.0"
	kafkaRestNotReadySuggestion = "Kafka REST connection is not ready. Re-running the command may solve the issue."
)

func (c *command) newKafkaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage a local instance of Apache Kafka.",
	}

	cmd.AddCommand(c.newKafkaBrokerCommand())
	cmd.AddCommand(c.newKafkaClusterCommand())
	cmd.AddCommand(c.newKafkaStartCommand())
	cmd.AddCommand(c.newKafkaStopCommand())
	cmd.AddCommand(c.newKafkaTopicCommand())

	return cmd
}

func getShortenedContainerId(id string) string {
	containerIdShortLength := 10
	return id[:containerIdShortLength]
}

func checkIsDockerRunning(dockerClient *client.Client) error {
	if _, err := dockerClient.Info(context.Background()); err != nil {
		return errors.NewErrorWithSuggestions(
			err.Error(),
			"Make sure Docker has been installed following the guide at https://docs.docker.com/engine/install/ and is running.",
		)
	}

	return nil
}
