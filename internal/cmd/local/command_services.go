package local

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

type Service struct {
	startDependencies       []string
	stopDependencies        []string
	startCommand            string
	properties              string
	port                    int
	isConfluentPlatformOnly bool
}

var (
	services = map[string]Service{
		"connect": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
				"schema-registry",
			},
			stopDependencies:        []string{},
			startCommand:            "connect-distributed",
			properties:              "schema-registry/connect-avro-distributed.properties",
			port:                    8083,
			isConfluentPlatformOnly: false,
		},
		"control-center": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
				"schema-registry",
				"connect",
				"ksql-server",
			},
			stopDependencies:        []string{},
			startCommand:            "control-center-start",
			properties:              "confluent-control-center/control-center-dev.properties",
			port:                    9021,
			isConfluentPlatformOnly: true,
		},
		"kafka": {
			startDependencies: []string{
				"zookeeper",
			},
			stopDependencies: []string{
				"control-center",
				"ksql-server",
				"connect",
				"kafka-rest",
				"schema-registry",
			},
			startCommand:            "kafka-server-start",
			properties:              "kafka/server.properties",
			port:                    9092,
			isConfluentPlatformOnly: false,
		},
		"kafka-rest": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
				"schema-registry",
			},
			stopDependencies:        []string{},
			startCommand:            "kafka-rest-start",
			properties:              "kafka-rest/kafka-rest.properties",
			port:                    8082,
			isConfluentPlatformOnly: false,
		},
		"ksql-server": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
				"schema-registry",
			},
			stopDependencies:        []string{},
			startCommand:            "ksql-server-start",
			properties:              "ksqldb/ksql-server.properties", // TODO: ksql/ksql-server.properties for older versions
			port:                    8088,
			isConfluentPlatformOnly: false,
		},
		"schema-registry": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
			},
			stopDependencies:        []string{},
			startCommand:            "schema-registry-start",
			properties:              "schema-registry/schema-registry.properties",
			port:                    8081,
			isConfluentPlatformOnly: false,
		},
		"zookeeper": {
			startDependencies: []string{},
			stopDependencies: []string{
				"control-center",
				"ksql-server",
				"connect",
				"kafka-rest",
				"schema-registry",
				"kafka",
			},
			startCommand:            "zookeeper-server-start",
			properties:              "kafka/zookeeper.properties",
			port:                    2181,
			isConfluentPlatformOnly: false,
		},
	}

	orderedServices = []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
		"control-center",
	}
)

func NewServicesCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "services [command]",
			Short: "Manage Confluent Platform services.",
			Args:  cobra.MinimumNArgs(1),
		},
		cfg, prerunner)

	availableServices, _ := getAvailableServices()
	for _, service := range availableServices {
		servicesCommand.AddCommand(NewServiceCommand(service, prerunner, cfg))
	}
	servicesCommand.AddCommand(NewServicesListCommand(prerunner, cfg))
	servicesCommand.AddCommand(NewServicesStartCommand(prerunner, cfg))
	servicesCommand.AddCommand(NewServicesStatusCommand(prerunner, cfg))
	servicesCommand.AddCommand(NewServicesStopCommand(prerunner, cfg))
	servicesCommand.AddCommand(NewServicesTopCommand(prerunner, cfg))

	return servicesCommand.Command
}

func NewServicesListCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesListCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all Confluent Platform services.",
			Args:  cobra.NoArgs,
			RunE:  runServicesListCommand,
		},
		cfg, prerunner)

	return servicesListCommand.Command
}

func runServicesListCommand(command *cobra.Command, _ []string) error {
	availableServices, err := getAvailableServices()
	if err != nil {
		return err
	}

	command.Println("Available Services:")
	command.Println(buildTabbedList(availableServices))
	return nil
}

func NewServicesStartCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesStartCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "start",
			Short: "Start all Confluent Platform services.",
			Args:  cobra.NoArgs,
			RunE:  runServicesStartCommand,
		},
		cfg, prerunner)

	return servicesStartCommand.Command
}

func runServicesStartCommand(command *cobra.Command, _ []string) error {
	availableServices, err := getAvailableServices()
	if err != nil {
		return err
	}

	if err := notifyConfluentCurrent(command); err != nil {
		return err
	}

	// Topological order
	for i := 0; i < len(availableServices); i++ {
		service := availableServices[i]
		if err := startService(command, service); err != nil {
			return err
		}
	}

	return nil
}

func NewServicesStatusCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesStatusCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "status",
			Short: "Check the status of all Confluent Platform services.",
			Args:  cobra.NoArgs,
			RunE:  runServicesStatusCommand,
		},
		cfg, prerunner)

	return servicesStatusCommand.Command
}

func runServicesStatusCommand(command *cobra.Command, _ []string) error {
	availableServices, err := getAvailableServices()
	if err != nil {
		return err
	}

	for _, service := range availableServices {
		if err := printStatus(command, service); err != nil {
			return err
		}
	}

	return nil
}

func NewServicesStopCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesStopCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "stop",
			Short: "Stop all Confluent Platform services.",
			Args:  cobra.NoArgs,
			RunE:  runServicesStopCommand,
		},
		cfg, prerunner)

	return servicesStopCommand.Command
}

func runServicesStopCommand(command *cobra.Command, _ []string) error {
	availableServices, err := getAvailableServices()
	if err != nil {
		return err
	}

	if err := notifyConfluentCurrent(command); err != nil {
		return err
	}

	// Reverse topological order
	for i := len(availableServices) - 1; i >= 0; i-- {
		service := availableServices[i]
		if err := stopService(command, service); err != nil {
			return err
		}
	}

	return nil
}

func NewServicesTopCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesTopCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "top",
			Short: "Monitor all Confluent Platform services.",
			Args:  cobra.NoArgs,
			RunE:  runServicesTopCommand,
		},
		cfg, prerunner)

	return servicesTopCommand.Command
}

func runServicesTopCommand(_ *cobra.Command, _ []string) error {
	availableServices, err := getAvailableServices()
	if err != nil {
		return err
	}

	pids := []int{}
	for _, service := range availableServices {
		dir, err := getServiceDir(service)
		if err != nil {
			return err
		}

		isUp, err := isRunning(service, dir)
		if err != nil {
			return err
		}

		if isUp {
			pid, err := readInt(getPidFile(service, dir))
			if err != nil {
				return err
			}
			pids = append(pids, pid)
		}
	}

	if len(pids) == 0 {
		return fmt.Errorf("no services are running")
	}

	return top(pids)
}

func getServiceConfig(service string, dir string) map[string]string {
	config := map[string]string{}

	switch service {
	case "connect":
		config["bootstrap.servers"] = fmt.Sprintf("localhost:%d", services["kafka"].port)
	case "control-center":
		config["confluent.controlcenter.data.dir"] = filepath.Join(dir, "data")
	case "kafka":
		config["log.dirs"] = filepath.Join(dir, "data")
	case "kafka-rest":
		config["zookeeper.connect"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		config["schema.registry.url"] = fmt.Sprintf("http://localhost:%d", services["schema-registry"].port)
	case "ksql-server":
		config["state.dir"] = filepath.Join(dir, "data", "kafka-streams")
		config["kafkastore.connection.url"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		config["ksql.schema.registry.url"] = fmt.Sprintf("http://localhost:%d", services["schema-registry"].port)
	case "schema-registry":
		config["kafkastore.connection.url"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
	case "zookeeper":
		config["dataDir"] = filepath.Join(dir, "data")
	}

	return config
}

func top(pids []int) error {
	var top *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		args := make([]string, len(pids) * 2)
		for i := 0; i < len(pids); i++ {
			args[i * 2] = "-pid"
			args[i * 2 + 1] = strconv.Itoa(pids[i])
		}
		top = exec.Command("top", args...)
	case "linux":
		args := make([]string, len(pids))
		for i := 0; i < len(pids); i++ {
			args[i] = strconv.Itoa(pids[i])
		}
		top = exec.Command("top", "-p", strings.Join(args, ","))
	default:
		return fmt.Errorf("top not available on platform: %s", runtime.GOOS)
	}

	top.Stdout, top.Stderr, top.Stdin = os.Stdout, os.Stderr, os.Stdin
	return top.Run()
}

func getAvailableServices() ([]string, error) {
	isCP, err := isConfluentPlatform()

	var available []string
	for _, service := range orderedServices {
		if isCP || !services[service].isConfluentPlatformOnly {
			available = append(available, service)
		}
	}

	return available, err
}

func notifyConfluentCurrent(command *cobra.Command) error {
	current, err := getConfluentCurrent()
	if err != nil {
		return err
	}

	command.Printf("Using CONFLUENT_CURRENT: %s\n", current)
	return nil
}
