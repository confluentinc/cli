package local

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/local"
)

type Service struct {
	startDependencies       []string
	stopDependencies        []string
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
			port:                    8088,
			isConfluentPlatformOnly: false,
		},
		"schema-registry": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
			},
			stopDependencies:        []string{},
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

	ch := local.NewConfluentHomeManager()

	availableServices, _ := getAvailableServices(ch)

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
	ch := local.NewConfluentHomeManager()

	availableServices, err := getAvailableServices(ch)
	if err != nil {
		return err
	}

	command.Println("Available Services:")
	command.Println(local.BuildTabbedList(availableServices))
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
	ch := local.NewConfluentHomeManager()

	availableServices, err := getAvailableServices(ch)
	if err != nil {
		return err
	}

	cc := local.NewConfluentCurrentManager()

	if err := notifyConfluentCurrent(command, cc); err != nil {
		return err
	}

	// Topological order
	for i := 0; i < len(availableServices); i++ {
		service := availableServices[i]
		if err := startService(command, ch, cc, service); err != nil {
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
	ch := local.NewConfluentHomeManager()

	availableServices, err := getAvailableServices(ch)
	if err != nil {
		return err
	}

	cc := local.NewConfluentCurrentManager()

	for _, service := range availableServices {
		if err := printStatus(command, cc, service); err != nil {
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
	ch := local.NewConfluentHomeManager()

	availableServices, err := getAvailableServices(ch)
	if err != nil {
		return err
	}

	cc := local.NewConfluentCurrentManager()

	if err := notifyConfluentCurrent(command, cc); err != nil {
		return err
	}

	// Reverse topological order
	for i := len(availableServices) - 1; i >= 0; i-- {
		service := availableServices[i]
		if err := stopService(command, cc, service); err != nil {
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

func runServicesTopCommand(command *cobra.Command, _ []string) error {
	ch := local.NewConfluentHomeManager()

	availableServices, err := getAvailableServices(ch)
	if err != nil {
		return err
	}

	cc := local.NewConfluentCurrentManager()

	var pids []int
	for _, service := range availableServices {
		isUp, err := isRunning(cc, service)
		if err != nil {
			return err
		}

		if isUp {
			pid, err := cc.GetPid(service)
			if err != nil {
				return err
			}
			pids = append(pids, pid)
		}
	}

	if len(pids) == 0 {
		command.PrintErrln("No services are running.")
		return nil
	}

	return top(pids)
}

func getConfig(ch local.ConfluentHome, cc local.ConfluentCurrent, service string) (map[string]string, error) {
	data, err := cc.GetDataDir(service)
	if err != nil {
		return map[string]string{}, err
	}

	isCP, err := ch.IsConfluentPlatform()
	if err != nil {
		return map[string]string{}, err
	}

	config := make(map[string]string)

	switch service {
	case "connect":
		config["bootstrap.servers"] = fmt.Sprintf("localhost:%d", services["kafka"].port)
		path, err := ch.GetConnectPluginPath()
		if err != nil {
			return map[string]string{}, err
		}
		config["plugin.path"] = path
		matches, err := ch.FindFile("share/java/kafka-connect-replicator/replicator-rest-extension-*.jar")
		if err != nil {
			return map[string]string{}, err
		}
		if len(matches) > 0 {
			classpath := fmt.Sprintf("%s:%s", os.Getenv("CLASSPATH"), matches[0])
			if err := os.Setenv("CLASSPATH", classpath); err != nil {
				return map[string]string{}, err
			}
			config["rest.extension.classes"] = "io.confluent.connect.replicator.monitoring.ReplicatorMonitoringExtension"
		}
		if isCP {
			config["producer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor"
			config["consumer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor"
		}
	case "control-center":
		config["confluent.controlcenter.data.dir"] = data
	case "kafka":
		config["log.dirs"] = data
		if isCP {
			config["metric.reporters"] = "io.confluent.metrics.reporter.ConfluentMetricsReporter"
			config["confluent.metrics.reporter.bootstrap.servers"] = fmt.Sprintf("localhost:%d", services["kafka"].port)
			config["confluent.metrics.reporter.topic.replicas"] = "1"
		}
	case "kafka-rest":
		config["schema.registry.url"] = fmt.Sprintf("http://localhost:%d", services["schema-registry"].port)
		config["zookeeper.connect"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		if isCP {
			config["producer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor"
			config["consumer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor"
		}
	case "ksql-server":
		config["kafkastore.connection.url"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		config["ksql.schema.registry.url"] = fmt.Sprintf("http://localhost:%d", services["schema-registry"].port)
		config["state.dir"] = data
		if isCP {
			config["producer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor"
			config["consumer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor"
		}
	case "schema-registry":
		config["kafkastore.connection.url"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		if isCP {
			config["producer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor"
			config["consumer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor"
		}
	case "zookeeper":
		config["dataDir"] = data
	}

	return config, nil
}

func top(pids []int) error {
	var top *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		args := make([]string, len(pids)*2)
		for i := 0; i < len(pids); i++ {
			args[i*2] = "-pid"
			args[i*2+1] = strconv.Itoa(pids[i])
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

	top.Stdout = os.Stdout
	top.Stderr = os.Stderr
	top.Stdin = os.Stdin

	return top.Run()
}

func getAvailableServices(ch local.ConfluentHome) ([]string, error) {
	isCP, err := ch.IsConfluentPlatform()

	var available []string
	for _, service := range orderedServices {
		if isCP || !services[service].isConfluentPlatformOnly {
			available = append(available, service)
		}
	}

	return available, err
}

func notifyConfluentCurrent(command *cobra.Command, cc local.ConfluentCurrent) error {
	dir, err := cc.GetCurrentDir()
	if err != nil {
		return err
	}

	command.Printf("Using CONFLUENT_CURRENT: %s\n", dir)
	return nil
}
