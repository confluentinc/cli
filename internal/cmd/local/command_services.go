package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
)

type Service struct {
	startDependencies       []string
	stopDependencies        []string
	startCommand            string
	properties              string
	isConfluentPlatformOnly bool
}

var (
	services = map[string]Service{
		"connect": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
			},
			stopDependencies: []string{
				"control-center",
			},
			startCommand:            "connect-distributed",
			properties:              "schema-registry/connect-avro-distributed.properties",
			isConfluentPlatformOnly: false,
		},
		"control-center": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
				"connect",
				"schema-registry",
				"ksql-server",
			},
			stopDependencies:        []string{},
			startCommand:            "control-center-start",
			properties:              "confluent-control-center/control-center-dev.properties",
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
			isConfluentPlatformOnly: false,
		},
		"ksql-server": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
				"schema-registry",
			},
			stopDependencies: []string{
				"control-center",
			},
			startCommand:            "ksql-server-start",
			properties:              "ksqldb/ksql-server.properties", // TODO: ksql/ksql-server.properties for older versions
			isConfluentPlatformOnly: false,
		},
		"schema-registry": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
			},
			stopDependencies: []string{
				"control-center",
				"ksql-server",
				"connect",
				"kafka-rest",
			},
			startCommand:            "schema-registry-start",
			properties:              "schema-registry/schema-registry.properties",
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
			isConfluentPlatformOnly: false,
		},
	}

	orderedServices = []string{
		"zookeeper",
		"kafka",
		"connect",
		"kafka-rest",
		"schema-registry",
		"ksql-server",
		"control-center",
	}
)

func NewServicesCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "services [command]",
			Short: "Manage all Confluent Platform services.",
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

	return servicesCommand.Command
}

func NewServicesListCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesListCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all Confluent Platform services.",
			Args:  cobra.NoArgs,
			RunE:  runListCommand,
		},
		cfg, prerunner)

	return servicesListCommand.Command
}

func runListCommand(command *cobra.Command, _ []string) error {
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
			RunE:  runStartCommand,
		},
		cfg, prerunner)

	return servicesStartCommand.Command
}

func runStartCommand(command *cobra.Command, _ []string) error {
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
			RunE:  runStatusCommand,
		},
		cfg, prerunner)

	return servicesStatusCommand.Command
}

func runStatusCommand(command *cobra.Command, _ []string) error {
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
			RunE:  runStopCommand,
		},
		cfg, prerunner)

	return servicesStopCommand.Command
}

func runStopCommand(command *cobra.Command, _ []string) error {
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
