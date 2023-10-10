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
	"github.com/confluentinc/cli/v3/pkg/utils"
)

const (
	confluentBrokerPrefix     = "confluent-local-broker-%d"
	controllerVoterPrefix     = "%d@confluent-local-broker-%d:%s"
	bootstrapServerPrefix     = "confluent-local-broker-%d:%s"
	confluentLocalNetworkName = "confluent-local-network"
)

type portsOut struct {
	KafkaRestPort  string `human:"Kafka REST Port" json:"kafka_rest_port"`
	PlaintextPorts string `human:"Plaintext Ports" json:"plaintext_ports"`
}

type ImagePullResponse struct {
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Progress string `json:"progress,omitempty"`
	ID       string `json:"id,omitempty"`
}

func (c *command) newKafkaStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a local Apache Kafka instance.",
		Args:  cobra.NoArgs,
		RunE:  c.kafkaStart,
	}

	cmd.Flags().String("kafka-rest-port", "8082", "Kafka REST port number.")
	cmd.Flags().StringSlice("plaintext-ports", nil, "A comma-separated list of port numbers for plaintext producer and consumer clients for brokers. If not specified, random free ports will be used.")
	cmd.Flags().Int32("brokers", 1, "Number of brokers (between 1 and 4, inclusive) in the Confluent Local Kafka cluster.")
	return cmd
}

func (c *command) kafkaStart(cmd *cobra.Command, args []string) error {
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
			output.Println(c.Config.EnableColor, "Confluent Local is already running.")
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
			output.Printf(c.Config.EnableColor, "\rDownloading: %s", response.Progress)
		} else if response.Status == "Extracting" {
			output.Printf(c.Config.EnableColor, "\rExtracting: %s", response.Progress)
		} else {
			output.Println(c.Config.EnableColor, "")
			output.Printf(c.Config.EnableColor, response.Status)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	output.Println(c.Config.EnableColor, "\r")

	log.CliLogger.Tracef("Successfully pulled Confluent Local image")

	brokers, err := cmd.Flags().GetInt32("brokers")
	if err != nil {
		return err
	}
	if brokers < 1 || brokers > 4 {
		return fmt.Errorf("--brokers must be an integer between 1 and 4, inclusive.")
	}

	if err := c.prepareAndSaveLocalPorts(cmd, brokers, c.Config.IsTest); err != nil {
		return err
	}
	if c.Config.LocalPorts == nil {
		return errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}

	ports := c.Config.LocalPorts
	platform := &specsv1.Platform{
		OS:           "linux",
		Architecture: runtime.GOARCH,
	}
	natKafkaRestPort := nat.Port(ports.KafkaRestPort + "/tcp")
	natPlaintextPorts := getNatPlaintextPorts(ports)
	containerStartCmd := strslice.StrSlice{"bash", "-c", "'/etc/confluent/docker/run'"}

	options := types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
	}
	if _, err := dockerClient.NetworkCreate(context.Background(), confluentLocalNetworkName, options); err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	var containerIds []string
	for idx := int32(0); idx < brokers; idx++ {
		brokerId := idx + 1
		config := &container.Config{
			Image:        dockerImageName,
			Hostname:     fmt.Sprintf(confluentBrokerPrefix, brokerId),
			Cmd:          containerStartCmd,
			ExposedPorts: nat.PortSet{natPlaintextPorts[idx]: struct{}{}},
			Env:          getContainerEnvironmentWithPorts(ports, idx, brokers),
		}

		hostConfig := &container.HostConfig{
			NetworkMode: container.NetworkMode("confluent-local-network"),
			PortBindings: nat.PortMap{natPlaintextPorts[idx]: []nat.PortBinding{{
				HostIP:   localhost,
				HostPort: ports.PlaintextPorts[idx],
			}}},
		}

		// expose Kafka REST port for broker 1
		if idx == 0 {
			config.ExposedPorts[natKafkaRestPort] = struct{}{}
			hostConfig.PortBindings[natKafkaRestPort] = []nat.PortBinding{{
				HostIP:   localhost,
				HostPort: ports.KafkaRestPort,
			}}
		}

		createResp, err := dockerClient.ContainerCreate(context.Background(), config, hostConfig, nil, platform, fmt.Sprintf(confluentBrokerPrefix, brokerId))
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
	portsData := &portsOut{
		c.Config.LocalPorts.KafkaRestPort,
		strings.Join(c.Config.LocalPorts.PlaintextPorts, ","),
	}
	table.Add(portsData)
	if err := table.Print(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Started Confluent Local containers %s.\n", utils.ArrayToCommaDelimitedString(containerIds, ""))
	output.Println(c.Config.EnableColor, "To continue your Confluent Local experience, run `confluent local kafka topic create <topic>` and `confluent local kafka topic produce <topic>`.")
	return nil
}

func (c *command) prepareAndSaveLocalPorts(cmd *cobra.Command, brokers int32, isTest bool) error {
	if c.Config.LocalPorts != nil {
		return nil
	}

	if isTest {
		c.Config.LocalPorts = &config.LocalPorts{
			BrokerPorts:     []string{"2996", "2997"},
			ControllerPorts: []string{"2998", "2999"},
			KafkaRestPort:   "8082",
			PlaintextPorts:  []string{"3002", "3003"},
		}
	} else {
		freePorts, err := freeport.GetFreePorts(int(3 * brokers))
		if err != nil {
			return err
		}

		c.Config.LocalPorts = &config.LocalPorts{KafkaRestPort: "8082"}
		for i := 0; i < int(brokers); i++ {
			c.Config.LocalPorts.PlaintextPorts = append(c.Config.LocalPorts.PlaintextPorts, strconv.Itoa(freePorts[i]))
			c.Config.LocalPorts.BrokerPorts = append(c.Config.LocalPorts.BrokerPorts, strconv.Itoa(freePorts[i+int(brokers)]))
			c.Config.LocalPorts.ControllerPorts = append(c.Config.LocalPorts.ControllerPorts, strconv.Itoa(freePorts[i+2*int(brokers)]))
		}

		kafkaRestPort, err := cmd.Flags().GetString("kafka-rest-port")
		if err != nil {
			return err
		}
		if kafkaRestPort != "" {
			c.Config.LocalPorts.KafkaRestPort = kafkaRestPort
		}

		plaintextPorts, err := cmd.Flags().GetStringSlice("plaintext-ports")
		if err != nil {
			return err
		}
		if len(plaintextPorts) > 0 {
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

func (c *command) validateCustomizedPorts() error {
	kafkaRestListener, err := net.Listen("tcp", ":"+c.Config.LocalPorts.KafkaRestPort)
	if err != nil {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return err
		}
		invalidPort := c.Config.LocalPorts.KafkaRestPort
		c.Config.LocalPorts.KafkaRestPort = strconv.Itoa(freePort)
		log.CliLogger.Warnf("Kafka REST port %s is not available, using port %d instead.", invalidPort, freePort)
	} else {
		if err := kafkaRestListener.Close(); err != nil {
			return err
		}
	}

	for idx, port := range c.Config.LocalPorts.PlaintextPorts {
		brokerId := idx + 1
		plaintextListener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			freePort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}
			c.Config.LocalPorts.PlaintextPorts[idx] = strconv.Itoa(freePort)
			log.CliLogger.Warnf("Plaintext port %s is not available, using port %d for broker %d instead.", port, freePort, brokerId)
		} else {
			if err := plaintextListener.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func getContainerEnvironmentWithPorts(ports *config.LocalPorts, idx int32, brokers int32) []string {
	brokerId := idx + 1
	envs := []string{
		fmt.Sprintf("KAFKA_BROKER_ID=%d", brokerId),
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
		fmt.Sprintf("KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://confluent-local-broker-%d:%s,PLAINTEXT_HOST://localhost:%s", brokerId, ports.BrokerPorts[idx], ports.PlaintextPorts[idx]),
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
		"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0",
		"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1",
		"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1",
		"KAFKA_PROCESS_ROLES=broker,controller",
		fmt.Sprintf("KAFKA_NODE_ID=%d", brokerId),
		getKafkaControllerQuorumVoters(ports, brokers),
		fmt.Sprintf("KAFKA_LISTENERS=PLAINTEXT://confluent-local-broker-%d:%s,CONTROLLER://confluent-local-broker-%d:%s,PLAINTEXT_HOST://0.0.0.0:%s", brokerId, ports.BrokerPorts[idx], brokerId, ports.ControllerPorts[idx], ports.PlaintextPorts[idx]),
		"KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT",
		"KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER",
		"KAFKA_LOG_DIRS=/tmp/kraft-combined-logs",
		"KAFKA_REST_HOST_NAME=rest-proxy",
	}
	if idx == 0 { // configure Kafka REST proxy broker 1
		envs = append(envs, fmt.Sprintf("KAFKA_REST_LISTENERS=http://0.0.0.0:%s", ports.KafkaRestPort))
		envs = append(envs, getKafkaRestBootstrapServers(ports, brokers))
	}
	return envs
}

func getNatPlaintextPorts(ports *config.LocalPorts) []nat.Port {
	res := []nat.Port{}
	for _, port := range ports.PlaintextPorts {
		res = append(res, nat.Port(port+"/tcp"))
	}
	return res
}

func getKafkaControllerQuorumVoters(ports *config.LocalPorts, brokers int32) string {
	voters := []string{fmt.Sprintf("KAFKA_CONTROLLER_QUORUM_VOTERS=1@confluent-local-broker-1:%s", ports.ControllerPorts[0])}
	for i := int32(1); i < brokers; i++ {
		voters = append(voters, fmt.Sprintf(controllerVoterPrefix, i+1, i+1, ports.ControllerPorts[i]))
	}
	return strings.Join(voters, ",")
}

func getKafkaRestBootstrapServers(ports *config.LocalPorts, brokers int32) string {
	servers := []string{fmt.Sprintf("KAFKA_REST_BOOTSTRAP_SERVERS=confluent-local-broker-1:%s", ports.BrokerPorts[0])}
	for i := int32(1); i < brokers; i++ {
		servers = append(servers, fmt.Sprintf(bootstrapServerPrefix, i+1, ports.BrokerPorts[i]))
	}
	return strings.Join(servers, ",")
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
