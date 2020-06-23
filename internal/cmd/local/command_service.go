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
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/local"
	"github.com/confluentinc/cli/internal/pkg/spinner"
)

func NewServiceCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   fmt.Sprintf("%s [command]", service),
			Short: fmt.Sprintf("Manage the %s service.", service),
			Args:  cobra.ExactArgs(1),
		},
		cfg, prerunner)

	serviceCommand.AddCommand(NewServiceLogCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceStartCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceStatusCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceStopCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceTopCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceVersionCommand(service, prerunner, cfg))

	switch service {
	case "connect":
		serviceCommand.AddCommand(NewConnectConnectorCommand(prerunner, cfg))
		serviceCommand.AddCommand(NewConnectPluginCommand(prerunner, cfg))
	case "kafka":
		serviceCommand.AddCommand(NewKafkaConsumeCommand(prerunner, cfg))
		serviceCommand.AddCommand(NewKafkaProduceCommand(prerunner, cfg))
	case "schema-registry":
		serviceCommand.AddCommand(NewSchemaRegistryACLCommand(prerunner, cfg))
	}

	return serviceCommand.Command
}

func NewServiceLogCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceLogCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "log",
			Short: fmt.Sprintf("Print logs for %s.", service),
			Args:  cobra.NoArgs,
			RunE:  runServiceLogCommand,
		},
		cfg, prerunner)

	return serviceLogCommand.Command
}

func runServiceLogCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	cc := local.NewConfluentCurrentManager()

	log, err := cc.GetLogFile(service)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(log)
	if err != nil {
		return err
	}
	command.Print(string(data))

	return nil
}

func NewServiceStartCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceVersionCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "start",
			Short: fmt.Sprintf("Start %s.", service),
			Args:  cobra.NoArgs,
			RunE:  runServiceStartCommand,
		},
		cfg, prerunner)

	return serviceVersionCommand.Command
}

func runServiceStartCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	cc := local.NewConfluentCurrentManager()

	if err := notifyConfluentCurrent(command, cc); err != nil {
		return err
	}

	ch := local.NewConfluentHomeManager()

	for _, dependency := range services[service].startDependencies {
		if err := startService(command, ch, cc, dependency); err != nil {
			return err
		}
	}

	return startService(command, ch, cc, service)
}

func NewServiceStatusCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceVersionCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "status",
			Short: fmt.Sprintf("Check the status of %s.", service),
			Args:  cobra.NoArgs,
			RunE:  runServiceStatusCommand,
		},
		cfg, prerunner)

	return serviceVersionCommand.Command
}

func runServiceStatusCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	cc := local.NewConfluentCurrentManager()

	return printStatus(command, cc, service)
}

func NewServiceStopCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceVersionCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "stop",
			Short: fmt.Sprintf("Stop %s.", service),
			Args:  cobra.NoArgs,
			RunE:  runServiceStopCommand,
		},
		cfg, prerunner)

	return serviceVersionCommand.Command
}

func runServiceStopCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	cc := local.NewConfluentCurrentManager()

	if err := notifyConfluentCurrent(command, cc); err != nil {
		return err
	}

	for _, dependency := range services[service].stopDependencies {
		if err := stopService(command, cc, dependency); err != nil {
			return err
		}
	}

	return stopService(command, cc, service)
}

func NewServiceTopCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceTopCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "top",
			Short: fmt.Sprintf("Monitor %s processes.", service),
			Args:  cobra.NoArgs,
			RunE:  runServiceTopCommand,
		},
		cfg, prerunner)

	return serviceTopCommand.Command
}

func runServiceTopCommand(command *cobra.Command, _ []string) error {
	cc := local.NewConfluentCurrentManager()

	service := command.Parent().Name()

	isUp, err := isRunning(cc, service)
	if err != nil {
		return err
	}
	if !isUp {
		return printStatus(command, cc, service)
	}

	pid, err := cc.GetPid(service)
	if err != nil {
		return err
	}

	return top([]int{pid})
}

func NewServiceVersionCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceVersionCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version",
			Short: fmt.Sprintf("Print the version of %s.", service),
			Args:  cobra.NoArgs,
			RunE:  runServiceVersionCommand,
		},
		cfg, prerunner)

	return serviceVersionCommand.Command
}

func runServiceVersionCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	ch := local.NewConfluentHomeManager()

	version, err := ch.GetVersion(service)
	if err != nil {
		return err
	}

	command.Println(version)
	return nil
}

func startService(command *cobra.Command, ch local.ConfluentHome, cc local.ConfluentCurrent, service string) error {
	isUp, err := isRunning(cc, service)
	if err != nil {
		return err
	}
	if isUp {
		return printStatus(command, cc, service)
	}

	if err := checkService(service); err != nil {
		return err
	}

	if err := configService(ch, cc, service); err != nil {
		return err
	}

	command.Printf("Starting %s\n", service)

	spin := spinner.New()
	spin.Start()
	err = startProcess(ch, cc, service)
	spin.Stop()
	if err != nil {
		return err
	}

	return printStatus(command, cc, service)
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

func configService(ch local.ConfluentHome, cc local.ConfluentCurrent, service string) error {
	port, err := ch.GetServicePort(service)
	if err != nil {
		if err.Error() != "no port specified" {
			return err
		}
	} else {
		services[service].port = port
	}

	data, err := ch.GetServiceConfig(service)
	if err != nil {
		return err
	}

	config, err := getConfig(ch, cc, service)
	if err != nil {
		return err
	}

	data = injectConfig(data, config)

	if err := cc.SetConfig(service, data); err != nil {
		return err
	}

	logs, err := cc.GetLogsDir(service)
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

func startProcess(ch local.ConfluentHome, cc local.ConfluentCurrent, service string) error {
	scriptFile, err := ch.GetServiceStartScript(service)
	if err != nil {
		return err
	}

	configFile, err := cc.GetConfigFile(service)
	if err != nil {
		return err
	}

	start := exec.Command(scriptFile, configFile)

	logFile, err := cc.GetLogFile(service)
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

	if err := cc.SetPid(service, start.Process.Pid); err != nil {
		return err
	}

	errors := make(chan error)

	up := make(chan bool)
	go func() {
		for {
			isUp, err := isRunning(cc, service)
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
		return fmt.Errorf("%s failed to start", service)
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
	case <-time.After(60 * time.Second):
		return fmt.Errorf("%s failed to start", service)
	}

	return nil
}

func stopService(command *cobra.Command, cc local.ConfluentCurrent, service string) error {
	isUp, err := isRunning(cc, service)
	if err != nil {
		return err
	}
	if !isUp {
		return printStatus(command, cc, service)
	}

	command.Printf("Stopping %s\n", service)

	spin := spinner.New()
	spin.Start()
	err = stopProcess(cc, service)
	spin.Stop()
	if err != nil {
		return err
	}

	return printStatus(command, cc, service)
}

func stopProcess(cc local.ConfluentCurrent, service string) error {
	pid, err := cc.GetPid(service)
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

	errors := make(chan error)

	up := make(chan bool)
	go func() {
		for {
			isUp, err := isRunning(cc, service)
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
	case <-time.After(time.Second):
		return fmt.Errorf("%s failed to stop", service)
	}

	if err := cc.RemovePidFile(service); err != nil {
		return err
	}

	return nil
}

func printStatus(command *cobra.Command, cc local.ConfluentCurrent, service string) error {
	isUp, err := isRunning(cc, service)
	if err != nil {
		return err
	}

	status := color.RedString("DOWN")
	if isUp {
		status = color.GreenString("UP")
	}

	command.Printf("%s is [%s]\n", service, status)
	return nil
}

func isRunning(cc local.ConfluentCurrent, service string) (bool, error) {
	hasPidFile, err := cc.HasPidFile(service)
	if err != nil {
		return false, err
	}
	if !hasPidFile {
		return false, nil
	}

	pid, err := cc.GetPid(service)
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
