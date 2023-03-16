package local

import (
	"context"
	"fmt"
	"strconv"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
)

func (c *localCommand) newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "---",
		Long:  "---",
		Args:  cobra.NoArgs,
		RunE:  c.start,
	}

	return cmd
}

func (c *localCommand) start(cmd *cobra.Command, args []string) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	_, err = dockerClient.Info(context.Background())
	if err != nil {
		return err
	}

	err = c.prepareValidPorts()
	if err != nil {
		return err
	}
	// pull the image from ecr. it will be public so no creds needed?
	// out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	// if err != nil {
	// 	return err
	// }
	// defer out.Close()
	// io.Copy(os.Stdout, out)
	// log.CliLogger.Tracef("Pull confluent-local image success")

	// create container
	ports := c.Config.LocalPorts
	platform := &specsv1.Platform{OS: "linux", Architecture: "amd64"}
	config := &container.Config{
		Image:        imageName,
		Hostname:     "broker",
		Cmd:          strslice.StrSlice{"bash", "-c", "'/etc/confluent/docker/run'"},
		ExposedPorts: nat.PortSet{nat.Port(ports.RestPort + "/tcp"): struct{}{}, nat.Port(ports.PlaintextPort + "/tcp"): struct{}{}},
		Env: []string{
			"KAFKA_BROKER_ID=1",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			// "KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://broker:29092,PLAINTEXT_HOST://localhost:9092", // fix
			fmt.Sprintf("KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://broker:%s,PLAINTEXT_HOST://localhost:%s", ports.BrokerPort, ports.PlaintextPort),
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1",
			"KAFKA_PROCESS_ROLES=broker,controller",
			"KAFKA_NODE_ID=1",
			// "KAFKA_CONTROLLER_QUORUM_VOTERS=1@broker:29093", // fix
			fmt.Sprintf("KAFKA_CONTROLLER_QUORUM_VOTERS=1@broker:%s", ports.ControllerPort),
			// "KAFKA_LISTENERS=PLAINTEXT://broker:29092,CONTROLLER://broker:29093,PLAINTEXT_HOST://0.0.0.0:9092", // fix
			fmt.Sprintf("KAFKA_LISTENERS=PLAINTEXT://broker:%s,CONTROLLER://broker:%s,PLAINTEXT_HOST://0.0.0.0:%s", ports.BrokerPort, ports.ControllerPort, ports.PlaintextPort),
			"KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT",
			"KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER",
			"KAFKA_LOG_DIRS=/tmp/kraft-combined-logs",
			"KAFKA_REST_HOST_NAME=rest-proxy",
			// "KAFKA_REST_BOOTSTRAP_SERVERS=broker:29092", // fix
			fmt.Sprintf("KAFKA_REST_BOOTSTRAP_SERVERS=broker:%s", ports.BrokerPort),
			// "KAFKA_REST_LISTENERS=http://0.0.0.0:8082", // fix
			fmt.Sprintf("KAFKA_REST_LISTENERS=http://0.0.0.0:%s", ports.RestPort),
		},
	}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(ports.PlaintextPort + "/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: ports.PlaintextPort,
				},
			},
		},
	}

	createResp, err := dockerClient.ContainerCreate(context.Background(), config, hostConfig, nil, platform, "confluent-local")
	if err != nil {
		return errors.CatchContainerNameInUseError(err)
	}
	log.CliLogger.Tracef("Create confluent-local container success")

	err = dockerClient.ContainerStart(context.Background(), createResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	log.CliLogger.Tracef("Start confluent-local container success")

	output.Printf("Started Confluent local container %v.\nContinue your experience with Confluent local running `confluent local kafka produce` and `confluent local kafka consume`.\n", createResp.ID[:10])
	return nil
}

func (c *localCommand) prepareValidPorts() error {
	if c.Config.LocalPorts != nil {
		return nil
	}

	restPortInt, err := freeport.GetFreePort()
	if err != nil {
		return err
	}

	plaintextPortInt, err := freeport.GetFreePort()
	if err != nil {
		return err
	}

	brokerPortInt, err := freeport.GetFreePort()
	if err != nil {
		return err
	}

	controllerPortInt, err := freeport.GetFreePort()
	if err != nil {
		return err
	}

	c.Config.LocalPorts = &v1.LocalPorts{
		RestPort:       strconv.Itoa(restPortInt),
		PlaintextPort:  strconv.Itoa(plaintextPortInt),
		BrokerPort:     strconv.Itoa(brokerPortInt),
		ControllerPort: strconv.Itoa(controllerPortInt),
	}

	if err := c.Config.Save(); err != nil {
		return errors.Wrap(err, errors.SavePortsToConfigErrorMsg)
	}

	return nil
}
