package local

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/phayes/freeport"
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	statusBlacklist = []string{"Pulling fs layer", "Waiting", "Downloading", "Download complete", "Verifying Checksum", "Extracting", "Pull complete"}
	localArch       = "amd64"
)

type imagePullOut struct {
	Status string `json:"status"`
}

func (c *command) newKafkaStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start a single-node instance of Apache Kafka.",
		Args:  cobra.NoArgs,
		RunE:  c.kafkaStart,
	}
}

func (c *command) kafkaStart(cmd *cobra.Command, args []string) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	if err := checkIsDockerRunning(dockerClient); err != nil {
		return err
	}

	out, err := dockerClient.ImagePull(context.Background(), dockerImageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, out)
	if err != nil {
		return err
	}

	for _, ss := range strings.Split(buf.String(), "\n") {
		var output imagePullOut
		if err := json.Unmarshal([]byte(ss), &output); err != nil {
			continue
		}
		var inBlacklist bool
		for _, s := range statusBlacklist {
			if output.Status == s {
				inBlacklist = true
			}
		}
		if !inBlacklist {
			fmt.Printf("%v\n", output.Status)
		}
	}

	log.CliLogger.Tracef("Pull confluent-local image success")

	platform := &specsv1.Platform{OS: "linux", Architecture: localArch}

	if err := c.prepareAndSaveLocalPorts(c.Config.IsTest); err != nil {
		return err
	}
	if c.Config.LocalPorts == nil {
		return errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}
	ports := c.Config.LocalPorts
	natKafkaRestPort := nat.Port(ports.KafkaRestPort + "/tcp")
	natPlaintextPort := nat.Port(ports.PlaintextPort + "/tcp")

	config := &container.Config{
		Image:    dockerImageName,
		Hostname: "broker",
		Cmd:      strslice.StrSlice{"bash", "-c", "'/etc/confluent/docker/run'"},
		ExposedPorts: nat.PortSet{
			natKafkaRestPort: struct{}{},
			natPlaintextPort: struct{}{},
		},
		Env: getContainerEnvironmentWithPorts(ports),
	}
	hostConfig := &container.HostConfig{PortBindings: nat.PortMap{
		natKafkaRestPort: []nat.PortBinding{
			{
				HostIP:   localhost,
				HostPort: ports.KafkaRestPort,
			},
		},
		natPlaintextPort: []nat.PortBinding{
			{
				HostIP:   localhost,
				HostPort: ports.PlaintextPort,
			},
		},
	},
	}

	createResp, err := dockerClient.ContainerCreate(context.Background(), config, hostConfig, nil, platform, confluentLocalContainerName)
	if err != nil {
		return errors.CatchContainerNameInUseError(err)
	}
	log.CliLogger.Tracef("Create confluent-local container success")

	err = dockerClient.ContainerStart(context.Background(), createResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	output.Printf("Started Confluent Local container %v.\nTo continue your Confluent Local experience, run `confluent local kafka topic create test` and `confluent local kafka topic produce test`.\n", getShortenedContainerId(createResp.ID))
	return nil
}

func (c *command) prepareAndSaveLocalPorts(isTest bool) error {
	if c.Config.LocalPorts != nil {
		return nil
	}

	if isTest {
		c.Config.LocalPorts = &v1.LocalPorts{
			KafkaRestPort:  "2996",
			PlaintextPort:  "2997",
			BrokerPort:     "2998",
			ControllerPort: "2999",
		}
	} else {
		freePorts, err := freeport.GetFreePorts(4)
		if err != nil {
			return err
		}

		c.Config.LocalPorts = &v1.LocalPorts{
			KafkaRestPort:  strconv.Itoa(freePorts[0]),
			PlaintextPort:  strconv.Itoa(freePorts[1]),
			BrokerPort:     strconv.Itoa(freePorts[2]),
			ControllerPort: strconv.Itoa(freePorts[3]),
		}
	}

	if err := c.Config.Save(); err != nil {
		return errors.Wrap(err, "failed to save local ports to configuration file")
	}

	return nil
}

func getContainerEnvironmentWithPorts(ports *v1.LocalPorts) []string {
	return []string{
		"KAFKA_BROKER_ID=1",
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
		fmt.Sprintf("KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://broker:%s,PLAINTEXT_HOST://localhost:%s", ports.BrokerPort, ports.PlaintextPort),
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
		"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0",
		"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1",
		"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1",
		"KAFKA_PROCESS_ROLES=broker,controller",
		"KAFKA_NODE_ID=1",
		fmt.Sprintf("KAFKA_CONTROLLER_QUORUM_VOTERS=1@broker:%s", ports.ControllerPort),
		fmt.Sprintf("KAFKA_LISTENERS=PLAINTEXT://broker:%s,CONTROLLER://broker:%s,PLAINTEXT_HOST://0.0.0.0:%s", ports.BrokerPort, ports.ControllerPort, ports.PlaintextPort),
		"KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT",
		"KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER",
		"KAFKA_LOG_DIRS=/tmp/kraft-combined-logs",
		"KAFKA_REST_HOST_NAME=rest-proxy",
		fmt.Sprintf("KAFKA_REST_BOOTSTRAP_SERVERS=broker:%s", ports.BrokerPort),
		fmt.Sprintf("KAFKA_REST_LISTENERS=http://0.0.0.0:%s", ports.KafkaRestPort),
	}
}
