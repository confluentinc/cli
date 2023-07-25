package local

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	dockerImageName             = "confluentinc/confluent-local:latest"
	confluentLocalContainerName = "confluent-local"
	localhostPrefix             = "http://localhost:%s"
	localhost                   = "localhost"
	kafkaRestNotReadySuggestion = "Kafka REST connection is not ready. Re-running the command may solve the issue."
)

func (c *Command) newKafkaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage a single-node instance of Apache Kafka.",
	}

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
	_, err := dockerClient.Info(context.Background())
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.InstallAndStartDockerSuggestion)
	}

	return nil
}
