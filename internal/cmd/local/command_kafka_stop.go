package local

import (
	"context"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

func (c *localCommand) newStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "---",
		Long:  "---",
		Args:  cobra.NoArgs,
		RunE:  c.stop,
	}

	return cmd
}

func (c *localCommand) stop(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container := range containers {
		if container.Image == imageName {
			log.CliLogger.Tracef("Stopping Confluent local container " + container.ID[:10])
			noWaitTimeout := 0 // to not wait for the container to exit gracefully
			if err := dockerClient.ContainerStop(ctx, container.ID, containertypes.StopOptions{Timeout: &noWaitTimeout}); err != nil {
				return err
			}
			log.CliLogger.Tracef("Confluent local container stopped")
			if err := dockerClient.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
				return err
			}
			log.CliLogger.Tracef("Confluent local container removed")
		}
	}

	output.Printf("Thank you for using Confluent local.\n") //refine/
	return nil
}
