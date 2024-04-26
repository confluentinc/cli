package local

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newKafkaStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the local Apache Kafka service.",
		Args:  cobra.NoArgs,
		RunE:  c.kafkaStop,
	}
}

func (c *command) kafkaStop(_ *cobra.Command, _ []string) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	if err := checkIsDockerRunning(dockerClient); err != nil {
		return err
	}

	return c.stopAndRemoveConfluentLocal(dockerClient)
}

func (c *command) stopAndRemoveConfluentLocal(dockerClient *client.Client) error {
	dockerContainers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return err
	}

	for _, dockerContainer := range dockerContainers {
		if dockerContainer.Image == dockerImageName {
			log.CliLogger.Tracef("Stopping Confluent Local container " + getShortenedContainerId(dockerContainer.ID))
			noWaitTimeout := 0 // to not wait for the container to exit gracefully
			if err := dockerClient.ContainerStop(context.Background(), dockerContainer.ID, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
				return err
			}
			log.CliLogger.Tracef("Confluent Local container stopped")
			if err := dockerClient.ContainerRemove(context.Background(), dockerContainer.ID, container.RemoveOptions{Force: true}); err != nil {
				return err
			}
			log.CliLogger.Tracef("Confluent Local container removed")

			output.Printf(c.Config.EnableColor, "Confluent Local has been stopped: removed container \"%s\".\n", getShortenedContainerId(dockerContainer.ID))
		}
	}

	c.Config.LocalPorts = nil
	if err := c.Config.Save(); err != nil {
		return fmt.Errorf("failed to remove local ports from config: %w", err)
	}

	return nil
}
