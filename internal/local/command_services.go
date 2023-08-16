package local

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/local"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/types"
)

type Service struct {
	startDependencies       []string
	stopDependencies        []string
	port                    int
	isConfluentPlatformOnly bool
	envPrefix               string
}

var (
	services = map[string]*Service{
		"connect": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
				"schema-registry",
			},
			stopDependencies:        []string{},
			port:                    8083,
			isConfluentPlatformOnly: false,
			envPrefix:               "CONNECT",
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
			envPrefix:               "CONTROL_CENTER",
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
			envPrefix:               "SAVED_KAFKA",
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
			envPrefix:               "KAFKAREST",
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
			envPrefix:               "KSQL",
		},
		"schema-registry": {
			startDependencies: []string{
				"zookeeper",
				"kafka",
			},
			stopDependencies:        []string{},
			port:                    8081,
			isConfluentPlatformOnly: false,
			envPrefix:               "SCHEMA_REGISTRY",
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
			envPrefix:               "ZOOKEEPER",
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

func NewServicesCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "services",
			Short: "Manage Confluent Platform services.",
			Args:  cobra.NoArgs,
		}, prerunner)

	availableServices, _ := c.getAvailableServices()

	for _, service := range availableServices {
		c.AddCommand(NewServiceCommand(service, prerunner))
	}

	c.AddCommand(NewServicesListCommand(prerunner))
	c.AddCommand(NewServicesStartCommand(prerunner))
	c.AddCommand(NewServicesStatusCommand(prerunner))
	c.AddCommand(NewServicesStopCommand(prerunner))
	c.AddCommand(NewServicesTopCommand(prerunner))

	return c.Command
}

func NewServicesListCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all Confluent Platform services.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServicesListCommand
	return c.Command
}

func (c *Command) runServicesListCommand(_ *cobra.Command, _ []string) error {
	services, err := c.getAvailableServices()
	if err != nil {
		return err
	}

	sort.Strings(services)

	serviceNames := make([]string, len(services))
	for i, service := range services {
		serviceNames[i] = writeServiceName(service)
	}

	output.Printf(errors.AvailableServicesMsg, local.BuildTabbedList(serviceNames))
	return nil
}

func NewServicesStartCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "start",
			Short: "Start all Confluent Platform services.",
			Args:  cobra.NoArgs,
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "Start all available services:",
					Code: "confluent local services start",
				},
				examples.Example{
					Text: "Start Apache Kafka® and ZooKeeper as its dependency:",
					Code: "confluent local services kafka start",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runServicesStartCommand

	return c.Command
}

func (c *Command) runServicesStartCommand(_ *cobra.Command, _ []string) error {
	availableServices, err := c.getAvailableServices()
	if err != nil {
		return err
	}

	if err := c.notifyConfluentCurrent(); err != nil {
		return err
	}

	// Topological order
	for i := 0; i < len(availableServices); i++ {
		service := availableServices[i]
		if err := c.startService(service, ""); err != nil {
			return err
		}
	}

	return nil
}

func NewServicesStatusCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "status",
			Short: "Check the status of all Confluent Platform services.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServicesStatusCommand
	return c.Command
}

func (c *Command) runServicesStatusCommand(cmd *cobra.Command, _ []string) error {
	availableServices, err := c.getAvailableServices()
	if err != nil {
		return err
	}

	if err := c.notifyConfluentCurrent(); err != nil {
		return err
	}

	sort.Strings(availableServices)
	for _, service := range availableServices {
		if err := c.printStatus(service); err != nil {
			return err
		}
	}

	return nil
}

func NewServicesStopCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "stop",
			Short: "Stop all Confluent Platform services.",
			Args:  cobra.NoArgs,
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "Stop all running services:",
					Code: "confluent local services stop",
				},
				examples.Example{
					Text: "Stop Apache Kafka® and its dependent services.",
					Code: "confluent local services kafka stop",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runServicesStopCommand

	return c.Command
}

func (c *Command) runServicesStopCommand(cmd *cobra.Command, _ []string) error {
	availableServices, err := c.getAvailableServices()
	if err != nil {
		return err
	}

	if err := c.notifyConfluentCurrent(); err != nil {
		return err
	}

	// Reverse topological order
	for i := len(availableServices) - 1; i >= 0; i-- {
		service := availableServices[i]
		if err := c.stopService(service); err != nil {
			return err
		}
	}

	return nil
}

func NewServicesTopCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "top",
			Short: "View resource usage for all Confluent Platform services.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServicesTopCommand

	return c.Command
}

func (c *Command) runServicesTopCommand(_ *cobra.Command, _ []string) error {
	availableServices, err := c.getAvailableServices()
	if err != nil {
		return err
	}

	var pids []int
	for _, service := range availableServices {
		isUp, err := c.isRunning(service)
		if err != nil {
			return err
		}

		if isUp {
			pid, err := c.cc.ReadPid(service)
			if err != nil {
				return err
			}
			pids = append(pids, pid)
		}
	}

	if len(pids) == 0 {
		return errors.New(errors.NoServicesRunningErrorMsg)
	}

	return top(pids)
}

func (c *Command) getConfig(service string) (map[string]string, error) {
	data, err := c.cc.GetDataDir(service)
	if err != nil {
		return map[string]string{}, err
	}

	isCP, err := c.ch.IsConfluentPlatform()
	if err != nil {
		return map[string]string{}, err
	}

	config := make(map[string]string)

	switch service {
	case "connect":
		config["bootstrap.servers"] = fmt.Sprintf("localhost:%d", services["kafka"].port)

		data, err := c.ch.ReadServiceConfig(service)
		if err != nil {
			return map[string]string{}, err
		}

		if path, ok := local.ExtractConfig(data)["plugin.path"].(string); ok {
			full, err := c.ch.GetFile("share", "java")
			if err != nil {
				return map[string]string{}, err
			}
			config["plugin.path"] = strings.ReplaceAll(path, "share/java", full)
		}

		matches, err := c.ch.FindFile("share/java/kafka-connect-replicator/replicator-rest-extension-*.jar")
		if err != nil {
			return map[string]string{}, err
		}
		if len(matches) > 0 {
			file, err := c.ch.GetFile(matches[0])
			if err != nil {
				return map[string]string{}, err
			}

			classpath := fmt.Sprintf("%s:%s", os.Getenv("CLASSPATH"), file)
			classpath = strings.TrimPrefix(classpath, ":")
			if err := os.Setenv("CLASSPATH", classpath); err != nil {
				return map[string]string{}, err
			}

			classes := []string{"io.confluent.connect.replicator.monitoring.ReplicatorMonitoringExtension"}
			if val, ok := local.ExtractConfig(data)["rest.extension.classes"]; ok {
				classes = append(classes, strings.Split(val.(string), ",")...)
			}
			config["rest.extension.classes"] = strings.Join(classes, ",")
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
	case "ksql-server":
		config["kafkastore.connection.url"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		config["ksql.schema.registry.url"] = fmt.Sprintf("http://localhost:%d", services["schema-registry"].port)
		config["state.dir"] = data
	case "schema-registry":
		config["kafkastore.connection.url"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
	case "zookeeper":
		config["dataDir"] = data
	}

	if isCP && types.Contains([]string{"connect", "kafka-rest", "ksql-server", "schema-registry"}, service) {
		config["consumer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor"
		config["producer.interceptor.classes"] = "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor"
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
		return errors.Errorf(errors.TopNotAvailableErrorMsg, runtime.GOOS)
	}

	top.Stdin = os.Stdin
	top.Stdout = os.Stdout
	top.Stderr = os.Stderr

	return top.Run()
}

func (c *Command) getAvailableServices() ([]string, error) {
	isCP, err := c.ch.IsConfluentPlatform()

	var available []string
	for _, service := range orderedServices {
		if isCP || !services[service].isConfluentPlatformOnly {
			available = append(available, service)
		}
	}

	return available, err
}

func (c *Command) notifyConfluentCurrent() error {
	dir, err := c.cc.GetCurrentDir()
	if err != nil {
		return err
	}

	output.Printf(errors.UsingConfluentCurrentMsg, dir)
	return nil
}
