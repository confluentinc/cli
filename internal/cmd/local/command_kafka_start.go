package local

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"

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

type ImagePullResponse struct {
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Progress string `json:"progress,omitempty"`
	ID       string `json:"id,omitempty"`
}

func (c *Command) newKafkaStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a single-node instance of Apache Kafka.",
		Args:  cobra.NoArgs,
		RunE:  c.kafkaStart,
	}

	cmd.Flags().String("kafka-rest-port", "8082", "The port number for Kafka REST.")
	cmd.Flags().String("plaintext-port", "", "The port number for plaintext producer and consumer clients. If not specified, a random free port will be used.")

	return cmd
}

func (c *Command) kafkaStart(cmd *cobra.Command, args []string) error {
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

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		var response ImagePullResponse
		text := scanner.Text()

		err := json.Unmarshal([]byte(text), &response)
		if err != nil {
			return err
		}

		if response.Status == "Downloading" {
			fmt.Printf("\rDownloading: %s", response.Progress)
		} else if response.Status == "Extracting" {
			fmt.Printf("\rExtracting: %s", response.Progress)
		} else {

			fmt.Printf("\n%s", response.Status)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Print("\r")

	log.CliLogger.Tracef("Pull confluent-local image success")

	if err := c.prepareAndSaveLocalPorts(cmd, c.Config.IsTest); err != nil {
		return err
	}

	if c.Config.LocalPorts == nil {
		return errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}
	ports := c.Config.LocalPorts
	platform := &specsv1.Platform{OS: "linux", Architecture: runtime.GOARCH}
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

func (c *Command) prepareAndSaveLocalPorts(cmd *cobra.Command, isTest bool) error {
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
		freePorts, err := freeport.GetFreePorts(3)
		if err != nil {
			return err
		}

		c.Config.LocalPorts = &v1.LocalPorts{
			PlaintextPort:  strconv.Itoa(freePorts[0]),
			BrokerPort:     strconv.Itoa(freePorts[1]),
			ControllerPort: strconv.Itoa(freePorts[2]),
		}

		if kafkaRestPort, err := cmd.Flags().GetString("kafka-rest-port"); err == nil && kafkaRestPort != "" {
			c.Config.LocalPorts.KafkaRestPort = kafkaRestPort
		} else {
			freePort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}
			c.Config.LocalPorts.KafkaRestPort = strconv.Itoa(freePort)
		}

		if plaintextPort, err := cmd.Flags().GetString("plaintext-port"); err == nil && plaintextPort != "" {
			c.Config.LocalPorts.PlaintextPort = plaintextPort
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
