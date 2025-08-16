package local

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/local"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type Service struct {
	startDependencies       []string
	stopDependencies        []string
	port                    int
	isConfluentPlatformOnly bool
	envPrefix               string
	versionConstraints      string
}

var (
	services = map[string]*Service{
		"connect": {
			startDependencies: []string{
				"zookeeper",
				"kraft-controller",
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
				"prometheus",
				"alertmanager",
			},
			stopDependencies:        []string{},
			port:                    9021,
			isConfluentPlatformOnly: true,
			envPrefix:               "CONTROL_CENTER",
		},
		"kafka": {
			startDependencies: []string{
				"zookeeper",
				"kraft-controller",
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
				"kraft-controller",
				"kafka",
				"schema-registry",
			},
			stopDependencies:        []string{},
			port:                    8082,
			isConfluentPlatformOnly: false,
			envPrefix:               "KAFKAREST",
		},
		"kraft-controller": {
			startDependencies: []string{},
			stopDependencies: []string{
				"ksql-server",
				"connect",
				"kafka-rest",
				"schema-registry",
				"kafka",
			},
			port:                    9093,
			isConfluentPlatformOnly: false,
			envPrefix:               "SAVED_KAFKA",
			versionConstraints:      ">= 8.0",
		},
		"ksql-server": {
			startDependencies: []string{
				"zookeeper",
				"kraft-controller",
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
				"kraft-controller",
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
			versionConstraints:      "< 8.0",
		},
		"prometheus": {
			startDependencies: []string{},
			stopDependencies: []string{
				"control-center",
			},
			port:                    9090,
			isConfluentPlatformOnly: false,
			envPrefix:               "PROMETHEUS",
			versionConstraints:      ">= 8.0",
		},
		"alertmanager": {
			startDependencies: []string{},
			stopDependencies: []string{
				"control-center",
			},
			port:                    9098,
			isConfluentPlatformOnly: false,
			envPrefix:               "ALERTMANAGER",
			versionConstraints:      ">= 8.0",
		},
	}

	orderedServices = []string{
		"zookeeper",
		"kraft-controller",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
		"prometheus",
		"alertmanager",
		"control-center",
	}
)

func NewServicesCommand(cfg *config.Config, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "services",
			Short: "Manage Confluent Platform services.",
			Args:  cobra.NoArgs,
		}, prerunner)

	availableServices, _ := c.getAvailableServices()
	if cfg.IsTest {
		availableServices = orderedServices
	}

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

func (c *command) runServicesListCommand(_ *cobra.Command, _ []string) error {
	services, err := c.getAvailableServices()
	if err != nil {
		return err
	}

	sort.Strings(services)

	serviceNames := make([]string, len(services))
	for i, service := range services {
		serviceNames[i] = writeServiceName(service)
	}

	output.Println(c.Config.EnableColor, "Available Services:")
	output.Println(c.Config.EnableColor, local.BuildTabbedList(serviceNames))
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
					Text: "Start Apache Kafka® and its dependency:",
					Code: "confluent local services kafka start",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runServicesStartCommand

	return c.Command
}

func (c *command) runServicesStartCommand(_ *cobra.Command, _ []string) error {
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

func (c *command) runServicesStatusCommand(_ *cobra.Command, _ []string) error {
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

func (c *command) runServicesStopCommand(_ *cobra.Command, _ []string) error {
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

func (c *command) runServicesTopCommand(_ *cobra.Command, _ []string) error {
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
		return fmt.Errorf("no services running")
	}

	return top(pids)
}

func (c *command) getConfig(service string) (map[string]string, error) {
	data, err := c.cc.GetDataDir(service)
	if err != nil {
		return map[string]string{}, err
	}

	isCP, err := c.ch.IsConfluentPlatform()
	if err != nil {
		return map[string]string{}, err
	}

	zookeeperMode, err := c.isZookeeperMode()
	if err != nil {
		return map[string]string{}, err
	}

	config := make(map[string]string)

	switch service {
	case "connect":
		config["bootstrap.servers"] = fmt.Sprintf("localhost:%d", services["kafka"].port)

		data, err := c.ch.ReadServiceConfig(service, zookeeperMode)
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
		if c.isC3(service) {
			dir := os.Getenv("CONTROL_CENTER_HOME")
			file, _ := os.ReadFile(dir + "/etc/confluent-control-center/control-center-local.properties")
			configs := local.ExtractConfig(file)
			alertmanager := configs["confluent.controlcenter.alertmanager.config.file"]
			prometheus := configs["confluent.controlcenter.prometheus.rules.file"]
			config["confluent.controlcenter.alertmanager.config.file"] = dir + "/" + alertmanager.(string)
			config["confluent.controlcenter.prometheus.rules.file"] = dir + "/" + prometheus.(string)
		}
	case "kafka":
		if zookeeperMode {
			config["log.dirs"] = data
		} else {
			config["log.dirs"] = filepath.Join(data, "kraft-broker-logs")
		}
		if isCP {
			config["metric.reporters"] = "io.confluent.metrics.reporter.ConfluentMetricsReporter"
			config["confluent.metrics.reporter.bootstrap.servers"] = fmt.Sprintf("localhost:%d", services["kafka"].port)
			config["confluent.metrics.reporter.topic.replicas"] = "1"
		}
	case "kafka-rest":
		config["schema.registry.url"] = fmt.Sprintf("http://localhost:%d", services["schema-registry"].port)
		if zookeeperMode {
			config["zookeeper.connect"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		}
	case "kraft-controller":
		config["log.dirs"] = filepath.Join(data, "kraft-controller-logs")
		if isCP {
			config["metric.reporters"] = "io.confluent.metrics.reporter.ConfluentMetricsReporter"
			config["confluent.metrics.reporter.bootstrap.servers"] = fmt.Sprintf("localhost:%d", services["kafka"].port)
			config["confluent.metrics.reporter.topic.replicas"] = "1"
		}
	case "ksql-server":
		if zookeeperMode {
			config["kafkastore.connection.url"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		}
		config["ksql.schema.registry.url"] = fmt.Sprintf("http://localhost:%d", services["schema-registry"].port)
		config["state.dir"] = data
	case "schema-registry":
		if zookeeperMode {
			config["kafkastore.connection.url"] = fmt.Sprintf("localhost:%d", services["zookeeper"].port)
		}
	case "zookeeper":
		config["dataDir"] = data
	}

	if isCP && slices.Contains([]string{"connect", "kafka-rest", "ksql-server", "schema-registry"}, service) && zookeeperMode {
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
		return fmt.Errorf("`top` command not available on platform: %s", runtime.GOOS)
	}

	top.Stdin = os.Stdin
	top.Stdout = os.Stdout
	top.Stderr = os.Stderr

	return top.Run()
}

func (c *command) getAvailableServices() ([]string, error) {
	var errs *multierror.Error
	isCP, err := c.ch.IsConfluentPlatform()
	errs = multierror.Append(err)

	var available []string
	for _, service := range orderedServices {
		compatible, err := c.isCompatibleService(service)
		errs = multierror.Append(err)
		if !compatible {
			continue
		}

		if isCP || !services[service].isConfluentPlatformOnly {
			available = append(available, service)
		}
	}

	return available, errs.ErrorOrNil()
}

func (c *command) isCompatibleService(service string) (bool, error) {
	if services[service].versionConstraints == "" {
		return true, nil
	}

	confluentVersion, err := c.ch.GetConfluentVersion()
	if err != nil {
		return false, err
	}

	constraints, err := version.NewConstraint(services[service].versionConstraints)
	if err != nil {
		return false, err
	}

	ver, err := version.NewVersion(confluentVersion)
	if err != nil {
		return false, err
	}

	return constraints.Check(ver.Core()), nil
}

func (c *command) notifyConfluentCurrent() error {
	dir, err := c.cc.GetCurrentDir()
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Using CONFLUENT_CURRENT: %s\n", dir)
	return nil
}

func (c *command) isZookeeperMode() (bool, error) {
	availableServices, err := c.getAvailableServices()
	if err != nil {
		return false, err
	}

	return slices.Contains(availableServices, "zookeeper"), nil
}
