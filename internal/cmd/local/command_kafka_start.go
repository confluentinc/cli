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

	"github.com/confluentinc/cli/internal/pkg/config"
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
		Short: "Start a single-node or two-node instance of Apache Kafka.",
		Args:  cobra.NoArgs,
		RunE:  c.kafkaStart,
	}

	cmd.Flags().StringSlice("kafka-rest-ports", []string{"8082", "9092"}, "The port number for Kafka REST for brokers.")
	cmd.Flags().StringSlice("plaintext-ports", nil, "The port number for plaintext producer and consumer clients for brokers. If not specified, a random free port will be used.")
	cmd.Flags().Bool("multi-broker", false, "Start Confluent Local cluster with two brokers.")
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
	natKafkaRestPorts := []nat.Port{nat.Port(ports.KafkaRestPorts[0] + "/tcp"), nat.Port(ports.KafkaRestPorts[1] + "/tcp")}
	natPlaintextPorts := []nat.Port{nat.Port(ports.PlaintextPorts[0] + "/tcp"), nat.Port(ports.PlaintextPorts[1] + "/tcp")}
	containerStartCmd := strslice.StrSlice{"bash", "-c", "'/etc/confluent/docker/run'"}

	// create a customized network
	_, err = dockerClient.NetworkCreate(
		context.Background(),
		"confluent-local-network",
		types.NetworkCreate{
			CheckDuplicate: true,
			Driver:         "bridge",
		},
	)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	config1 := &container.Config{
		Image:    dockerImageName,
		Hostname: "broker1",
		Cmd:      containerStartCmd,
		ExposedPorts: nat.PortSet{
			natKafkaRestPorts[0]: struct{}{},
			natPlaintextPorts[0]: struct{}{},
		},
		Env: getContainerEnvironmentWithPorts(ports, 0),
	}
	hostConfig1 := &container.HostConfig{
		NetworkMode: container.NetworkMode("confluent-local-network"),
		PortBindings: nat.PortMap{
			natKafkaRestPorts[0]: []nat.PortBinding{
				{
					HostIP:   localhost,
					HostPort: ports.KafkaRestPorts[0],
				},
			},
			natPlaintextPorts[0]: []nat.PortBinding{
				{
					HostIP:   localhost,
					HostPort: ports.PlaintextPorts[0],
				},
			},
		},
	}

	config2 := &container.Config{
		Image:    dockerImageName,
		Hostname: "broker2",
		Cmd:      containerStartCmd,
		ExposedPorts: nat.PortSet{
			natKafkaRestPorts[1]: struct{}{},
			natPlaintextPorts[1]: struct{}{},
		},
		Env: getContainerEnvironmentWithPorts(ports, 1),
	}
	hostConfig2 := &container.HostConfig{
		NetworkMode: container.NetworkMode("confluent-local-network"),
		PortBindings: nat.PortMap{
			natKafkaRestPorts[1]: []nat.PortBinding{
				{
					HostIP:   localhost,
					HostPort: ports.KafkaRestPorts[1],
				},
			},
			natPlaintextPorts[1]: []nat.PortBinding{
				{
					HostIP:   localhost,
					HostPort: ports.PlaintextPorts[1],
				},
			},
		},
	}

	createResp1, err := dockerClient.ContainerCreate(context.Background(), config1, hostConfig1, nil, platform, "broker1")
	if err != nil {
		return err
	}
	log.CliLogger.Tracef("Create confluent-local container for broker 1 success")
	err = dockerClient.ContainerStart(context.Background(), createResp1.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	multiBroker, err := cmd.Flags().GetBool("multi-broker")
	if err != nil {
		return err
	}
	if multiBroker {
		createResp2, err := dockerClient.ContainerCreate(context.Background(), config2, hostConfig2, nil, platform, "broker2")
		if err != nil {
			return err
		}
		log.CliLogger.Tracef("Create confluent-local container for broker 2 success")
		err = dockerClient.ContainerStart(context.Background(), createResp2.ID, types.ContainerStartOptions{})
		if err != nil {
			return err
		}

		output.Printf("Started Confluent Local containers %v, %v.\nTo continue your Confluent Local experience, run `confluent local kafka topic create test` and `confluent local kafka topic produce test`.\n", getShortenedContainerId(createResp1.ID), getShortenedContainerId(createResp2.ID))
	} else {
		output.Printf("Started Confluent Local containers %v.\nTo continue your Confluent Local experience, run `confluent local kafka topic create test` and `confluent local kafka topic produce test`.\n", getShortenedContainerId(createResp1.ID))

	}

	table := output.NewTable(cmd)
	table.Add(c.Config.LocalPorts)
	return table.Print()
}

func (c *Command) prepareAndSaveLocalPorts(cmd *cobra.Command, isTest bool) error {
	if c.Config.LocalPorts != nil {
		return nil
	}

	if isTest {
		c.Config.LocalPorts = &config.LocalPorts{
			BrokerPorts:     []string{"2996", "2997"},
			ControllerPorts: []string{"2998", "2999"},
			KafkaRestPorts:  []string{"3000", "3001"},
			PlaintextPorts:  []string{"3002", "3003"},
		}
	} else {
		freePorts, err := freeport.GetFreePorts(6)
		if err != nil {
			return err
		}

		c.Config.LocalPorts = &config.LocalPorts{
			KafkaRestPorts:  []string{strconv.Itoa(8082), strconv.Itoa(9092)},
			PlaintextPorts:  []string{strconv.Itoa(freePorts[0]), strconv.Itoa(freePorts[1])},
			BrokerPorts:     []string{strconv.Itoa(freePorts[2]), strconv.Itoa(freePorts[3])},
			ControllerPorts: []string{strconv.Itoa(freePorts[4]), strconv.Itoa(freePorts[5])},
		}

		kafkaRestPorts, err := cmd.Flags().GetStringSlice("kafka-rest-ports")
		if err != nil {
			return err
		}
		if len(kafkaRestPorts) != 0 {
			c.Config.LocalPorts.KafkaRestPorts = kafkaRestPorts
		}

		plaintextPorts, err := cmd.Flags().GetStringSlice("plaintext-ports")
		if err != nil {
			return err
		}
		if len(plaintextPorts) != 0 {
			c.Config.LocalPorts.PlaintextPorts = plaintextPorts
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
	for idx := 0; idx < len(c.Config.LocalPorts.KafkaRestPorts); idx++ {
		brokerId := idx + 1
		kafkaRestLn, err := net.Listen("tcp", ":"+c.Config.LocalPorts.KafkaRestPorts[idx])
		if err != nil {
			freePort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}
			invalidKafkaRestPort := c.Config.LocalPorts.KafkaRestPorts[idx]
			c.Config.LocalPorts.KafkaRestPorts[idx] = strconv.Itoa(freePort)
			log.CliLogger.Warnf("Kafka REST port %s is not available, using port %d for broker %d instead.", invalidKafkaRestPort, freePort, brokerId)
		} else {
			if err := kafkaRestLn.Close(); err != nil {
				return err
			}
		}

		plaintextLn, err := net.Listen("tcp", ":"+c.Config.LocalPorts.PlaintextPorts[idx])
		if err != nil {
			freePort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}
			invalidPlaintextPort := c.Config.LocalPorts.PlaintextPorts[idx]
			c.Config.LocalPorts.PlaintextPorts[idx] = strconv.Itoa(freePort)
			log.CliLogger.Warnf("Plaintext port %s is not available, using port %d for broker %d instead.", invalidPlaintextPort, freePort, brokerId)
		} else {
			if err := plaintextLn.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func getContainerEnvironmentWithPorts(ports *config.LocalPorts, idx int) []string {
	brokerId := idx + 1
	a := []string{
		fmt.Sprintf("KAFKA_BROKER_ID=%d", brokerId),
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
		fmt.Sprintf("KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://broker%d:%s,PLAINTEXT_HOST://localhost:%s", brokerId, ports.BrokerPorts[idx], ports.PlaintextPorts[idx]),
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
		"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0",
		"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1",
		"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1",
		"KAFKA_PROCESS_ROLES=broker,controller",
		fmt.Sprintf("KAFKA_NODE_ID=%d", brokerId),
		fmt.Sprintf("KAFKA_CONTROLLER_QUORUM_VOTERS=1@broker1:%s,2@broker2:%s", ports.ControllerPorts[0], ports.ControllerPorts[1]),
		fmt.Sprintf("KAFKA_LISTENERS=PLAINTEXT://broker%d:%s,CONTROLLER://broker%d:%s,PLAINTEXT_HOST://0.0.0.0:%s", brokerId, ports.BrokerPorts[idx], brokerId, ports.ControllerPorts[idx], ports.PlaintextPorts[idx]),
		"KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT",
		"KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER",
		"KAFKA_LOG_DIRS=/tmp/kraft-combined-logs",
		"KAFKA_REST_HOST_NAME=rest-proxy",
	}
	if idx == 0 { // configure krest proxy only for the first broker.
		a = append(a, fmt.Sprintf("KAFKA_REST_LISTENERS=http://0.0.0.0:%s", ports.KafkaRestPorts[0]))
		a = append(a, fmt.Sprintf("KAFKA_REST_BOOTSTRAP_SERVERS=broker1:%s,broker2:%s", ports.BrokerPorts[0], ports.BrokerPorts[1]))
	}
	fmt.Println("a", a)
	return a
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
