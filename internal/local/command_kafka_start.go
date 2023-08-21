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

	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
)

var defaultKafkaRestPorts = []string{"8082", "9092", "10012", "20012"}

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

	cmd.Flags().StringSlice("kafka-rest-ports", nil, "A comma-separated list of Kafka REST port numbers for brokers.")
	cmd.Flags().StringSlice("plaintext-ports", nil, "A comma-separated list of port numbers for plaintext producer and consumer clients for brokers. If not specified, random free ports will be used.")
	cmd.Flags().Int32("brokers", 1, "Number of brokers in the Confluent Local cluster.") // range: [1, 4]
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
	log.CliLogger.Tracef("Successfully pulled Confluent Local image")

	numOfBrokers, err := cmd.Flags().GetInt32("brokers")
	if err != nil {
		return err
	}
	if err := c.prepareAndSaveLocalPorts(cmd, numOfBrokers, c.Config.IsTest); err != nil {
		return err
	}

	if c.Config.LocalPorts == nil {
		return errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}

	ports := c.Config.LocalPorts
	platform := &specsv1.Platform{OS: "linux", Architecture: runtime.GOARCH}
	natKafkaRestPorts := getNatKafkaRestPorts(ports, numOfBrokers)
	fmt.Println(natKafkaRestPorts)
	natPlaintextPorts := getNatPlaintextPorts(ports, numOfBrokers)
	fmt.Println(natPlaintextPorts)
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

	var containerIds []string
	for idx := int32(0); idx < numOfBrokers; idx++ {
		brokerId := idx + 1
		config := &container.Config{
			Image:    dockerImageName,
			Hostname: fmt.Sprintf("confluent-local-broker-%d", brokerId),
			Cmd:      containerStartCmd,
			ExposedPorts: nat.PortSet{
				natKafkaRestPorts[idx]: struct{}{},
				natPlaintextPorts[idx]: struct{}{},
			},
			Env: getContainerEnvironmentWithPorts(ports, idx, numOfBrokers),
		}

		hostConfig := &container.HostConfig{
			NetworkMode: container.NetworkMode("confluent-local-network"),
			PortBindings: nat.PortMap{
				natKafkaRestPorts[idx]: []nat.PortBinding{
					{
						HostIP:   localhost,
						HostPort: ports.KafkaRestPorts[idx],
					},
				},
				natPlaintextPorts[0]: []nat.PortBinding{
					{
						HostIP:   localhost,
						HostPort: ports.PlaintextPorts[idx],
					},
				},
			},
		}

		createResp, err := dockerClient.ContainerCreate(context.Background(), config, hostConfig, nil, platform, fmt.Sprintf("confluent-local-broker-%d", brokerId))
		if err != nil {
			return err
		}
		log.CliLogger.Trace(fmt.Sprintf("Successfully created a Confluent Local container for broker %d", brokerId))
		if err := dockerClient.ContainerStart(context.Background(), createResp.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
		containerIds = append(containerIds, getShortenedContainerId(createResp.ID))
	}

	table := output.NewTable(cmd)
	table.Add(c.Config.LocalPorts)
	err = table.Print()
	if err != nil {
		return err
	}
	output.Printf("Started Confluent Local containers %s.\nTo continue your Confluent Local experience, run `confluent local kafka topic create test` and `confluent local kafka topic produce test`.\n", strings.Join(containerIds, ","))

	return nil
}

func (c *Command) prepareAndSaveLocalPorts(cmd *cobra.Command, numOfBrokers int32, isTest bool) error {
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
		freePorts, err := freeport.GetFreePorts(int(3 * numOfBrokers))
		if err != nil {
			return err
		}

		c.Config.LocalPorts = &config.LocalPorts{}
		for i := 0; i < int(numOfBrokers); i++ {
			c.Config.LocalPorts.KafkaRestPorts = append(c.Config.LocalPorts.KafkaRestPorts, defaultKafkaRestPorts[i])
			c.Config.LocalPorts.PlaintextPorts = append(c.Config.LocalPorts.PlaintextPorts, strconv.Itoa(freePorts[i]))
			c.Config.LocalPorts.BrokerPorts = append(c.Config.LocalPorts.BrokerPorts, strconv.Itoa(freePorts[i+int(numOfBrokers)]))
			c.Config.LocalPorts.ControllerPorts = append(c.Config.LocalPorts.ControllerPorts, strconv.Itoa(freePorts[i+2*int(numOfBrokers)]))
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
	for idx, port := range c.Config.LocalPorts.KafkaRestPorts {
		brokerId := idx + 1
		kafkaRestLn, err := net.Listen("tcp", ":"+port)
		if err != nil {
			freePort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}
			c.Config.LocalPorts.KafkaRestPorts[idx] = strconv.Itoa(freePort)
			log.CliLogger.Warnf("Kafka REST port %s is not available, using port %d for broker %d instead.", port, freePort, brokerId)
		} else {
			if err := kafkaRestLn.Close(); err != nil {
				return err
			}
		}
	}
	for idx, port := range c.Config.LocalPorts.PlaintextPorts {
		brokerId := idx + 1
		plaintextLn, err := net.Listen("tcp", ":"+port)
		if err != nil {
			freePort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}
			c.Config.LocalPorts.PlaintextPorts[idx] = strconv.Itoa(freePort)
			log.CliLogger.Warnf("Plaintext port %s is not available, using port %d for broker %d instead.", port, freePort, brokerId)
		} else {
			if err := plaintextLn.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func getContainerEnvironmentWithPorts(ports *config.LocalPorts, idx int32, numOfBrokers int32) []string {
	brokerId := idx + 1
	a := []string{
		fmt.Sprintf("KAFKA_BROKER_ID=%d", brokerId),
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
		fmt.Sprintf("KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://confluent-local-broker-%d:%s,PLAINTEXT_HOST://localhost:%s", brokerId, ports.BrokerPorts[idx], ports.PlaintextPorts[idx]),
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
		"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0",
		"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1",
		"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1",
		"KAFKA_PROCESS_ROLES=broker,controller",
		fmt.Sprintf("KAFKA_NODE_ID=%d", brokerId),
		getKafkaControllerQuorumVoters(ports, numOfBrokers),
		fmt.Sprintf("KAFKA_LISTENERS=PLAINTEXT://confluent-local-broker-%d:%s,CONTROLLER://confluent-local-broker-%d:%s,PLAINTEXT_HOST://0.0.0.0:%s", brokerId, ports.BrokerPorts[idx], brokerId, ports.ControllerPorts[idx], ports.PlaintextPorts[idx]),
		"KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT",
		"KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER",
		"KAFKA_LOG_DIRS=/tmp/kraft-combined-logs",
		"KAFKA_REST_HOST_NAME=rest-proxy",
	}
	if idx == 0 { // configure krest proxy only for the first broker.
		a = append(a, fmt.Sprintf("KAFKA_REST_LISTENERS=http://0.0.0.0:%s", ports.KafkaRestPorts[0]))
		a = append(a, getKafkaRestBootstrapServers(ports, numOfBrokers))
	}
	fmt.Println("a", a)
	return a
}

func getNatKafkaRestPorts(ports *config.LocalPorts, numOfBrokers int32) []nat.Port {
	res := []nat.Port{}
	for i := 0; i < int(numOfBrokers); i++ {
		res = append(res, nat.Port(ports.KafkaRestPorts[i]+"/tcp"))
	}
	return res
}

func getNatPlaintextPorts(ports *config.LocalPorts, numOfBrokers int32) []nat.Port {
	res := []nat.Port{}
	for i := 0; i < int(numOfBrokers); i++ {
		res = append(res, nat.Port(ports.PlaintextPorts[i]+"/tcp"))
	}
	return res
}

func getKafkaControllerQuorumVoters(ports *config.LocalPorts, numOfBrokers int32) string {
	voters := fmt.Sprintf("KAFKA_CONTROLLER_QUORUM_VOTERS=1@confluent-local-broker-1:%s", ports.ControllerPorts[0])
	for i := int32(1); i < numOfBrokers; i++ {
		voters += fmt.Sprintf(",%v@confluent-local-broker-%v:%s", i+1, i+1, ports.ControllerPorts[i])
	}
	return voters
}

func getKafkaRestBootstrapServers(ports *config.LocalPorts, numOfBrokers int32) string {
	servers := fmt.Sprintf("KAFKA_REST_BOOTSTRAP_SERVERS=confluent-local-broker-1:%s", ports.BrokerPorts[0])
	for i := int32(1); i < numOfBrokers; i++ {
		servers += fmt.Sprintf(",confluent-local-broker-%v:%s", i+1, ports.BrokerPorts[i])
	}
	return servers
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
