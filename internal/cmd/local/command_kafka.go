package local

import (
	"context"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

const (
	confluentLocalImageName     = "confluentinc/confluent-local:latest"
	confluentLocalContainerName = "confluent-local"
	localhostPrefix             = "http://localhost:%s"
	localhost                   = "localhost"
)

func (c *command) newKafkaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage a single-node instance of Apache Kafka.",
	}

	cmd.AddCommand(c.newKafkaStartCommand())
	cmd.AddCommand(c.newKafkaStopCommand())
	cmd.AddCommand(c.newTopicCommand())

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
