package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/spinner"
)

func NewServiceCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   service,
			Short: fmt.Sprintf("Manage the %s service.", writeServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.AddCommand(NewServiceLogCommand(service, prerunner))
	c.AddCommand(NewServiceStartCommand(service, prerunner))
	c.AddCommand(NewServiceStatusCommand(service, prerunner))
	c.AddCommand(NewServiceStopCommand(service, prerunner))
	c.AddCommand(NewServiceTopCommand(service, prerunner))
	c.AddCommand(NewServiceVersionCommand(service, prerunner))

	switch service {
	case "connect":
		c.AddCommand(NewConnectConnectorCommand(prerunner))
		c.AddCommand(NewConnectPluginCommand(prerunner))
	case "kafka":
		c.AddCommand(NewKafkaConsumeCommand(prerunner))
		c.AddCommand(NewKafkaProduceCommand(prerunner))
	case "schema-registry":
		c.AddCommand(NewSchemaRegistryACLCommand(prerunner))
	}

	return c.Command
}

func NewServiceLogCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "log",
			Short: fmt.Sprintf("Print logs for %s.", writeServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceLogCommand
	return c.Command
}

func (c *Command) runServiceLogCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	log, err := c.cc.GetLogFile(service)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(log)
	if err != nil {
		return fmt.Errorf("no log found: to run %s, use \"confluent local services %s start\"", writeServiceName(service), service)
	}
	command.Print(string(data))

	return nil
}

func NewServiceStartCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "start",
			Short: fmt.Sprintf("Start the %s service.", writeServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceStartCommand
	c.Command.Flags().StringP("config", "c", "", fmt.Sprintf("Configure %s with a specific properties file.", service))

	return c.Command
}

func (c *Command) runServiceStartCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	if err := c.notifyConfluentCurrent(command); err != nil {
		return err
	}

	for _, dependency := range services[service].startDependencies {
		if err := c.startService(command, dependency, ""); err != nil {
			return err
		}
	}

	configFile, err := command.Flags().GetString("config")
	if err != nil {
		return err
	}

	return c.startService(command, service, configFile)
}

func NewServiceStatusCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "status",
			Short: fmt.Sprintf("Check the status of %s.", writeServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceStatusCommand
	return c.Command
}

func (c *Command) runServiceStatusCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	if err := c.notifyConfluentCurrent(command); err != nil {
		return err
	}

	return c.printStatus(command, service)
}

func NewServiceStopCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "stop",
			Short: fmt.Sprintf("Stop the %s service.", writeServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceStopCommand
	return c.Command
}

func (c *Command) runServiceStopCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	if err := c.notifyConfluentCurrent(command); err != nil {
		return err
	}

	for _, dependency := range services[service].stopDependencies {
		if err := c.stopService(command, dependency); err != nil {
			return err
		}
	}

	return c.stopService(command, service)
}

func NewServiceTopCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "top",
			Short: fmt.Sprintf("Monitor %s processes.", writeServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceTopCommand
	return c.Command
}

func (c *Command) runServiceTopCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	isUp, err := c.isRunning(service)
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus(command, service)
	}

	pid, err := c.cc.ReadPid(service)
	if err != nil {
		return err
	}

	return top([]int{pid})
}

func NewServiceVersionCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "version",
			Short: fmt.Sprintf("Print the version of %s.", writeServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceVersionCommand

	return c.Command
}

func (c *Command) runServiceVersionCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	ver, err := c.ch.GetVersion(service)
	if err != nil {
		return err
	}

	command.Println(ver)
	return nil
}

func (c *Command) startService(command *cobra.Command, service string, configFile string) error {
	isUp, err := c.isRunning(service)
	if err != nil {
		return err
	}
	if isUp {
		return c.printStatus(command, service)
	}

	if err := checkService(service); err != nil {
		return err
	}

	if err := c.configService(service, configFile); err != nil {
		return err
	}

	command.Printf("Starting %s\n", writeServiceName(service))

	spin := spinner.New()
	spin.Start()
	err = c.startProcess(service)
	spin.Stop()
	if err != nil {
		return err
	}

	return c.printStatus(command, service)
}

func checkService(service string) error {
	if err := checkOSVersion(); err != nil {
		return err
	}

	if err := checkJavaVersion(service); err != nil {
		return err
	}

	return nil
}

func (c *Command) configService(service string, configFile string) error {
	port, err := c.ch.ReadServicePort(service)
	if err != nil {
		if err.Error() != "no port specified" {
			return err
		}
	} else {
		services[service].port = port
	}

	var data []byte
	if configFile == "" {
		data, err = c.ch.ReadServiceConfig(service)
	} else {
		data, err = ioutil.ReadFile(configFile)
	}
	if err != nil {
		return err
	}

	config, err := c.getConfig(service)
	if err != nil {
		return err
	}

	data = injectConfig(data, config)

	if err := c.cc.WriteConfig(service, data); err != nil {
		return err
	}

	logs, err := c.cc.GetLogsDir(service)
	if err != nil {
		return err
	}
	if err := os.Setenv("LOG_DIR", logs); err != nil {
		return err
	}

	if err := setServiceEnvs(service); err != nil {
		return err
	}

	return nil
}

func injectConfig(data []byte, config map[string]string) []byte {
	for key, val := range config {
		re := regexp.MustCompile(fmt.Sprintf(`(?m)^(#\s)?%s=.+\n`, key))
		line := []byte(fmt.Sprintf("%s=%s\n", key, val))

		matches := re.FindAll(data, -1)
		switch len(matches) {
		case 0:
			data = append(data, line...)
		case 1:
			data = re.ReplaceAll(data, line)
		default:
			re := regexp.MustCompile(fmt.Sprintf(`(?m)^%s=.+\n`, key))
			data = re.ReplaceAll(data, line)
		}
	}

	return data
}

func (c *Command) startProcess(service string) error {
	scriptFile, err := c.ch.GetServiceScript("start", service)
	if err != nil {
		return err
	}

	configFile, err := c.cc.GetConfigFile(service)
	if err != nil {
		return err
	}

	start := exec.Command(scriptFile, configFile)

	logFile, err := c.cc.GetLogFile(service)
	if err != nil {
		return err
	}

	fd, err := os.Create(logFile)
	if err != nil {
		return err
	}
	start.Stdout = fd
	start.Stderr = fd

	if err := start.Start(); err != nil {
		return err
	}

	if err := c.cc.WritePid(service, start.Process.Pid); err != nil {
		return err
	}

	errors := make(chan error)

	up := make(chan bool)
	go func() {
		for {
			isUp, err := c.isRunning(service)
			if err != nil {
				errors <- err
			}
			if isUp {
				up <- isUp
			}
		}
	}()
	select {
	case <-up:
		break
	case err := <-errors:
		return err
	case <-time.After(time.Second):
		return fmt.Errorf("%s failed to start", writeServiceName(service))
	}

	open := make(chan bool)
	go func() {
		for {
			isOpen, err := isPortOpen(service)
			if err != nil {
				errors <- err
			}
			if isOpen {
				open <- isOpen
			}
			time.Sleep(time.Second)
		}
	}()
	select {
	case <-open:
		break
	case err := <-errors:
		return err
	case <-time.After(90 * time.Second):
		return fmt.Errorf("%s failed to start", writeServiceName(service))
	}

	return nil
}

func (c *Command) stopService(command *cobra.Command, service string) error {
	isUp, err := c.isRunning(service)
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus(command, service)
	}

	command.Printf("Stopping %s\n", writeServiceName(service))

	spin := spinner.New()
	spin.Start()
	err = c.stopProcess(service)
	spin.Stop()
	if err != nil {
		return err
	}

	return c.printStatus(command, service)
}

func (c *Command) stopProcess(service string) error {
	scriptFile, err := c.ch.GetServiceScript("stop", service)
	if err != nil {
		return err
	}

	if scriptFile == "" {
		pid, err := c.cc.ReadPid(service)
		if err != nil {
			return err
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}

		if err := process.Kill(); err != nil {
			return err
		}
	} else {
		stop := exec.Command(scriptFile)
		if err := stop.Start(); err != nil {
			return err
		}
	}

	errors := make(chan error)
	up := make(chan bool)
	go func() {
		for {
			isUp, err := c.isRunning(service)
			if err != nil {
				errors <- err
			}
			if !isUp {
				up <- isUp
			}
		}
	}()
	select {
	case <-up:
		break
	case err := <-errors:
		return err
	case <-time.After(10 * time.Second):
		if err := c.killProcess(service); err != nil {
			return err
		}
	}

	if err := c.cc.RemovePidFile(service); err != nil {
		return err
	}

	return nil
}

func (c *Command) killProcess(service string) error {
	pid, err := c.cc.ReadPid(service)
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.SIGKILL); err != nil {
		return err
	}

	errors := make(chan error)
	up := make(chan bool)
	go func() {
		for {
			isUp, err := c.isRunning(service)
			if err != nil {
				errors <- err
			}
			if !isUp {
				up <- isUp
			}
		}
	}()
	select {
	case <-up:
		return nil
	case err := <-errors:
		return err
	case <-time.After(time.Second):
		return fmt.Errorf("%s failed to stop", writeServiceName(service))
	}
}

func (c *Command) printStatus(command *cobra.Command, service string) error {
	isUp, err := c.isRunning(service)
	if err != nil {
		return err
	}

	status := color.RedString("DOWN")
	if isUp {
		status = color.GreenString("UP")
	}

	command.Printf("%s is [%s]\n", writeServiceName(service), status)
	return nil
}

func (c *Command) isRunning(service string) (bool, error) {
	hasPidFile, err := c.cc.HasPidFile(service)
	if err != nil {
		return false, err
	}
	if !hasPidFile {
		return false, nil
	}

	pid, err := c.cc.ReadPid(service)
	if err != nil {
		return false, err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}

	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}

	return true, nil
}

func isPortOpen(service string) (bool, error) {
	addr := fmt.Sprintf(":%d", services[service].port)
	out, err := exec.Command("lsof", "-i", addr).Output()
	if err != nil {
		return false, nil
	}
	return len(out) > 0, nil
}

func setServiceEnvs(service string) error {
	serviceEnvFormats := map[string]string{
		"KAFKA_LOG4J_OPTS":           "%s_LOG4J_OPTS",
		"EXTRA_ARGS":                 "%s_EXTRA_ARGS",
		"KAFKA_HEAP_OPTS":            "%s_HEAP_OPTS",
		"KAFKA_JVM_PERFORMANCE_OPTS": "%s_JVM_PERFORMANCE_OPTS",
		"KAFKA_GC_LOG_OPTS":          "%s_GC_LOG_OPTS",
		"KAFKA_JMX_OPTS":             "%s_JMX_OPTS",
		"KAFKA_DEBUG":                "%s_DEBUG",
		"KAFKA_OPTS":                 "%s_OPTS",
		"CLASSPATH":                  "%s_CLASSPATH",
		"JMX_PORT":                   "%s_JMX_PORT",
	}

	for _, envFormat := range serviceEnvFormats {
		env := fmt.Sprintf(envFormat, "KAFKA")
		savedEnv := fmt.Sprintf("SAVED_%s", env)
		if os.Getenv(savedEnv) == "" {
			val := os.Getenv(env)
			if val != "" {
				if err := os.Setenv(savedEnv, val); err != nil {
					return err
				}
			}
		}
	}

	prefix := services[service].envPrefix
	for env, envFormat := range serviceEnvFormats {
		val := os.Getenv(fmt.Sprintf(envFormat, prefix))
		if val != "" {
			if err := os.Setenv(env, val); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkOSVersion() error {
	// CLI-84: Require macOS version >= 10.13
	if runtime.GOOS == "darwin" {
		osVersion, err := exec.Command("sw_vers", "-productVersion").Output()
		if err != nil {
			return err
		}

		v, err := version.NewSemver(strings.TrimSuffix(string(osVersion), "\n"))
		if err != nil {
			return err
		}

		v10_13, _ := version.NewSemver("10.13")
		if v.Compare(v10_13) < 0 {
			return fmt.Errorf("macOS version >= 10.13 is required (detected: %s)", osVersion)
		}
	}
	return nil
}

func checkJavaVersion(service string) error {
	java := filepath.Join(os.Getenv("JAVA_HOME"), "/bin/java")
	if os.Getenv("JAVA_HOME") == "" {
		out, err := exec.Command("which", "java").Output()
		if err != nil {
			return err
		}
		java = strings.TrimSuffix(string(out), "\n")
		if java == "java not found" {
			return fmt.Errorf("could not find java executable, please install java or set JAVA_HOME")
		}
	}

	data, err := exec.Command(java, "-version").CombinedOutput()
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`.+ version "([\d._]+)"`)
	javaVersion := string(re.FindSubmatch(data)[1])

	isValid, err := isValidJavaVersion(service, javaVersion)
	if err != nil {
		return err
	}
	if !isValid {
		return fmt.Errorf("the Confluent CLI requires Java version 1.8 or 1.11.\nSee https://docs.confluent.io/current/installation/versions-interoperability.html\nIf you have multiple versions of Java installed, you may need to set JAVA_HOME to the version you want Confluent to use.")
	}

	return nil
}

func isValidJavaVersion(service, javaVersion string) (bool, error) {
	// 1.8.0_152 -> 8.0_152 -> 8.0
	javaVersion = strings.TrimPrefix(javaVersion, "1.")
	javaVersion = strings.Split(javaVersion, "_")[0]

	v, err := version.NewSemver(javaVersion)
	if err != nil {
		return false, err
	}

	v8, _ := version.NewSemver("8")
	v9, _ := version.NewSemver("9")
	v11, _ := version.NewSemver("11")
	if v.Compare(v8) < 0 || v.Compare(v9) >= 0 && v.Compare(v11) < 0 {
		return false, nil
	}

	if service == "zookeeper" || service == "kafka" {
		return true, nil
	}

	v12, _ := version.NewSemver("12")
	if v.Compare(v12) >= 0 {
		return false, nil
	}

	return true, nil
}

func writeServiceName(service string) string {
	switch service {
	case "kafka-rest":
		return "Kafka REST"
	case "ksql-server":
		return "KSQL Server"
	}

	return strings.Title(strings.ReplaceAll(service, "-", " "))
}
