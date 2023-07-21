package local

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
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
	"github.com/confluentinc/cli/internal/pkg/form"
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
	cmd.Flags().String("plaintext-port", "", "The port number for plaintext producer and consumer clients. If not specified, a random free port is used.")

	return cmd
}

func (c *Command) kafkaStart(cmd *cobra.Command, args []string) error {
	if err := checkMachineArch(); err != nil {
		return err
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	if err := checkIsDockerRunning(dockerClient); err != nil {
		return err
	}

	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}

	for _, container := range containers {
		if container.Image == dockerImageName {
			output.Println("Confluent Local is already running.")
			prompt := form.NewPrompt()
			f := form.New(form.Field{
				ID:        "confirm",
				Prompt:    "Do you wish to start a new Confluent Local session? Current context will be lost.",
				IsYesOrNo: true,
			})
			if err := f.Prompt(prompt); err != nil {
				return err
			}
			if f.Responses["confirm"].(bool) {
				if err := c.stopAndRemoveConfluentLocal(dockerClient); err != nil {
					return err
				}
			} else {
				return nil
			}
		}
	}

	out, err := dockerClient.ImagePull(context.Background(), dockerImageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		response := new(ImagePullResponse)
		text := scanner.Text()
		if err := json.Unmarshal([]byte(text), &response); err != nil {
			return err
		}
		if response.Status == "Downloading" {
			output.Printf("\rDownloading: %s", response.Progress)
		} else if response.Status == "Extracting" {
			output.Printf("\rExtracting: %s", response.Progress)
		} else {
			output.Printf("\n%s", response.Status)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	output.Println("\r")

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
		return err
	}
	log.CliLogger.Tracef("Create confluent-local container success")

	err = dockerClient.ContainerStart(context.Background(), createResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	output.Printf("Started Confluent Local container %v.\nTo continue your Confluent Local experience, run `confluent local kafka topic create test` and `confluent local kafka topic produce test`.\n", getShortenedContainerId(createResp.ID))

	table := output.NewTable(cmd)
	table.Add(c.Config.LocalPorts)
	return table.Print()
}

func (c *Command) prepareAndSaveLocalPorts(cmd *cobra.Command, isTest bool) error {
	if c.Config.LocalPorts != nil {
		return nil
	}

	if isTest {
		c.Config.LocalPorts = &v1.LocalPorts{
			BrokerPort:     "2996",
			ControllerPort: "2997",
			KafkaRestPort:  "2998",
			PlaintextPort:  "2999",
		}
	} else {
		freePorts, err := freeport.GetFreePorts(3)
		if err != nil {
			return err
		}

		c.Config.LocalPorts = &v1.LocalPorts{
			KafkaRestPort:  strconv.Itoa(8082),
			PlaintextPort:  strconv.Itoa(freePorts[0]),
			BrokerPort:     strconv.Itoa(freePorts[1]),
			ControllerPort: strconv.Itoa(freePorts[2]),
		}

		kafkaRestPort, err := cmd.Flags().GetString("kafka-rest-port")
		if err != nil {
			return err
		}
		if kafkaRestPort != "" {
			c.Config.LocalPorts.KafkaRestPort = kafkaRestPort
		}

		plaintextPort, err := cmd.Flags().GetString("plaintext-port")
		if err != nil {
			return err
		}
		if plaintextPort != "" {
			c.Config.LocalPorts.PlaintextPort = plaintextPort
		}
	}

	if err := c.validateCustomizedPorts(); err != nil {
		return err
	}

	if err := c.Config.Save(); err != nil {
		return errors.Wrap(err, "failed to save local ports to configuration file")
	}

	return nil
}

func (c *Command) validateCustomizedPorts() error {
	kafkaRestLn, err := net.Listen("tcp", ":"+c.Config.LocalPorts.KafkaRestPort)
	if err != nil {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return err
		}
		invalidKafkaRestPort := c.Config.LocalPorts.KafkaRestPort
		c.Config.LocalPorts.KafkaRestPort = strconv.Itoa(freePort)
		log.CliLogger.Warnf("Kafka REST port %s is not available, using port %d instead.", invalidKafkaRestPort, freePort)
	} else {
		if err := kafkaRestLn.Close(); err != nil {
			return err
		}
	}

	plaintextLn, err := net.Listen("tcp", ":"+c.Config.LocalPorts.PlaintextPort)
	if err != nil {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return err
		}
		invalidPlaintextPort := c.Config.LocalPorts.PlaintextPort
		c.Config.LocalPorts.PlaintextPort = strconv.Itoa(freePort)
		log.CliLogger.Warnf("Plaintext port %s is not available, using port %d instead.", invalidPlaintextPort, freePort)
	} else {
		if err := plaintextLn.Close(); err != nil {
			return err
		}
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

func checkMachineArch() error {
	if runtime.GOOS == "windows" {
		return nil
	}

	cmd := exec.Command("uname", "-m") // outputs system architecture info
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	systemArch := strings.TrimSpace(string(output))
	if systemArch == "x86_64" {
		systemArch = "amd64"
	}
	if systemArch != runtime.GOARCH {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(`binary architecture "%s" does not match system architecture "%s"`, runtime.GOARCH, systemArch), "Download the CLI with the correct architecture to continue.")
	}
	return nil
}
