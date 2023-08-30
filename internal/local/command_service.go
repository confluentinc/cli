package local

import (
	"fmt"
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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/spinner"
)

func NewServiceCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   service,
		Short: fmt.Sprintf("Manage %s.", writeOfficialServiceName(service)),
		Args:  cobra.NoArgs,
	}

	switch service {
	case "zookeeper":
		cmd.Aliases = []string{"zk"}
	case "schema-registry":
		cmd.Aliases = []string{"sr"}
	}

	c := NewLocalCommand(cmd, prerunner)

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
			Short: fmt.Sprintf("Print logs showing %s output.", writeOfficialServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceLogCommand
	c.Command.Flags().BoolP("follow", "f", false, "Log additional output until the command is interrupted.")

	return c.Command
}

func (c *command) runServiceLogCommand(cmd *cobra.Command, _ []string) error {
	service := cmd.Parent().Name()

	exists, err := c.cc.HasLogFile(service)
	if err != nil {
		return err
	}
	if !exists {
		return errors.Errorf(errors.NoLogFoundErrorMsg, writeOfficialServiceName(service), service)
	}

	log, err := c.cc.GetLogFile(service)
	if err != nil {
		return err
	}

	shouldFollow, err := cmd.Flags().GetBool("follow")
	if err != nil {
		return err
	}

	show := exec.Command("cat", log)
	if shouldFollow {
		show = exec.Command("tail", "-f", log)
	}

	show.Stdout = os.Stdout
	show.Stderr = os.Stderr
	return show.Run()
}

func NewServiceStartCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	longDescription := ""
	if service == "kafka" {
		longDescription = fmt.Sprintf("Start %s. For a faster and more lightweight experience, consider using `confluent local kafka start`. In the next major confluent CLI version, this command will be removed and replaced by ongoing support for `confluent local kafka`.", writeOfficialServiceName(service))
	}
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "start",
			Short: fmt.Sprintf("Start %s.", writeOfficialServiceName(service)),
			Long:  longDescription,
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceStartCommand
	c.Command.Flags().StringP("config", "c", "", fmt.Sprintf("Configure %s with a specific properties file.", writeOfficialServiceName(service)))

	return c.Command
}

func (c *command) runServiceStartCommand(cmd *cobra.Command, _ []string) error {
	service := cmd.Parent().Name()

	if err := c.notifyConfluentCurrent(); err != nil {
		return err
	}

	for _, dependency := range services[service].startDependencies {
		if err := c.startService(dependency, ""); err != nil {
			return err
		}
	}

	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	return c.startService(service, config)
}

func NewServiceStatusCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "status",
			Short: fmt.Sprintf("Check if %s is running.", writeOfficialServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceStatusCommand
	return c.Command
}

func (c *command) runServiceStatusCommand(cmd *cobra.Command, _ []string) error {
	service := cmd.Parent().Name()

	if err := c.notifyConfluentCurrent(); err != nil {
		return err
	}

	return c.printStatus(service)
}

func NewServiceStopCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "stop",
			Short: fmt.Sprintf("Stop %s.", writeOfficialServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceStopCommand
	return c.Command
}

func (c *command) runServiceStopCommand(cmd *cobra.Command, _ []string) error {
	service := cmd.Parent().Name()

	if err := c.notifyConfluentCurrent(); err != nil {
		return err
	}

	for _, dependency := range services[service].stopDependencies {
		if err := c.stopService(dependency); err != nil {
			return err
		}
	}

	return c.stopService(service)
}

func NewServiceTopCommand(service string, prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "top",
			Short: fmt.Sprintf("View resource usage for %s.", writeOfficialServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceTopCommand
	return c.Command
}

func (c *command) runServiceTopCommand(cmd *cobra.Command, _ []string) error {
	service := cmd.Parent().Name()

	isUp, err := c.isRunning(service)
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus(service)
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
			Short: fmt.Sprintf("Print the current version of %s.", writeOfficialServiceName(service)),
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runServiceVersionCommand

	return c.Command
}

func (c *command) runServiceVersionCommand(cmd *cobra.Command, _ []string) error {
	service := cmd.Parent().Name()

	ver, err := c.ch.GetVersion(service)
	if err != nil {
		return err
	}

	output.Println(ver)
	return nil
}

func (c *command) startService(service, configFile string) error {
	isUp, err := c.isRunning(service)
	if err != nil {
		return err
	}
	if isUp {
		return c.printStatus(service)
	}

	if err := c.checkService(service); err != nil {
		return err
	}

	if err := c.configService(service, configFile); err != nil {
		return err
	}

	output.Printf(errors.StartingServiceMsg, writeServiceName(service))

	spin := spinner.New()
	spin.Start()
	err = c.startProcess(service)
	spin.Stop()
	if err != nil {
		return err
	}

	return c.printStatus(service)
}

func (c *command) checkService(service string) error {
	if err := c.checkOSVersion(); err != nil {
		return err
	}

	if err := c.checkJavaVersion(service); err != nil {
		return err
	}

	return nil
}

func (c *command) configService(service, configFile string) error {
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
		data, err = os.ReadFile(configFile)
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
	// If there is existing config data, and we are going to inject
	// at least one thing, then ensure we put a newline before
	// injecting any of our new content so as to not corrupt the config file.
	// We don't need to do this if the last character is already a newline.
	if len(config) > 0 && len(data) > 0 && string(data[len(data)-1:]) != "\n" {
		data = append(data, []byte("\n")...)
	}

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

func (c *command) startProcess(service string) error {
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

	errorsChan := make(chan error)

	up := make(chan bool)
	go func() {
		for {
			isUp, err := c.isRunning(service)
			if err != nil {
				errorsChan <- err
			}
			if isUp {
				up <- isUp
			}
		}
	}()
	select {
	case <-up:
		break
	case err := <-errorsChan:
		return err
	case <-time.After(time.Second):
		return errors.Errorf(errors.FailedToStartErrorMsg, writeServiceName(service))
	}

	open := make(chan bool)
	go func() {
		for {
			if isPortOpen(service) {
				open <- true
			}
			time.Sleep(time.Second)
		}
	}()
	select {
	case <-open:
		break
	case <-time.After(90 * time.Second):
		return errors.Errorf(errors.FailedToStartErrorMsg, writeServiceName(service))
	}

	return nil
}

func (c *command) stopService(service string) error {
	isUp, err := c.isRunning(service)
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus(service)
	}

	output.Printf(errors.StoppingServiceMsg, writeServiceName(service))

	spin := spinner.New()
	spin.Start()
	err = c.stopProcess(service)
	spin.Stop()
	if err != nil {
		return err
	}

	return c.printStatus(service)
}

func (c *command) stopProcess(service string) error {
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

	errs := make(chan error)
	up := make(chan bool)
	go func() {
		for {
			isUp, err := c.isRunning(service)
			if err != nil {
				errs <- err
			}
			if !isUp {
				up <- isUp
			}
		}
	}()
	select {
	case <-up:
		break
	case err := <-errs:
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

func (c *command) killProcess(service string) error {
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

	errorsChan := make(chan error)
	up := make(chan bool)
	go func() {
		for {
			isUp, err := c.isRunning(service)
			if err != nil {
				errorsChan <- err
			}
			if !isUp {
				up <- isUp
			}
		}
	}()
	select {
	case <-up:
		return nil
	case err := <-errorsChan:
		return err
	case <-time.After(time.Second):
		return errors.Errorf(errors.FailedToStopErrorMsg, writeServiceName(service))
	}
}

func (c *command) printStatus(service string) error {
	isUp, err := c.isRunning(service)
	if err != nil {
		return err
	}

	status := color.RedString("DOWN")
	if isUp {
		status = color.GreenString("UP")
	}

	output.Printf(errors.ServiceStatusMsg, writeServiceName(service), status)
	return nil
}

func (c *command) isRunning(service string) (bool, error) {
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

	return process.Signal(syscall.Signal(0)) == nil, nil
}

func isPortOpen(service string) bool {
	if _, err := os.Stat("/etc/redhat-release"); err == nil { // check to see if it's a RedHat OS (i.e. CentOS) which doesn't have `lsof`
		out, err := exec.Command("ss", "-lptn", fmt.Sprintf("( sport = :%d )", services[service].port)).Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), "LISTEN") // LISTEN is the state of the process; can't just check if len > 0 bc headers are always printed
	} else {
		addr := fmt.Sprintf(":%d", services[service].port)
		out, err := exec.Command("lsof", "-i", addr).Output()
		if err != nil {
			return false
		}
		return len(out) > 0
	}
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

func (c *command) checkOSVersion() error {
	if runtime.GOOS == "darwin" {
		required, _ := version.NewSemver("10.13")
		// CLI-584 CP 6.0 now requires at least 10.14
		above, err := c.ch.IsAtLeastVersion("6.0")
		if err != nil {
			return err
		}
		if above {
			required, _ = version.NewSemver("10.14")
		}

		osVersion, err := exec.Command("sw_vers", "-productVersion").Output()
		if err != nil {
			return err
		}

		v, err := version.NewSemver(strings.TrimSuffix(string(osVersion), "\n"))
		if err != nil {
			return err
		}

		if v.Compare(required) < 0 {
			return fmt.Errorf(errors.MacVersionErrorMsg, required.String(), osVersion)
		}
	}
	return nil
}

func (c *command) checkJavaVersion(service string) error {
	java := filepath.Join(os.Getenv("JAVA_HOME"), "/bin/java")
	if os.Getenv("JAVA_HOME") == "" {
		out, err := exec.Command("which", "java").Output()
		if err != nil {
			return err
		}
		java = strings.TrimSuffix(string(out), "\n")
		if java == "java not found" {
			return errors.New(errors.JavaExecNotFondErrorMsg)
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
		return errors.New(errors.JavaRequirementErrorMsg)
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

func writeOfficialServiceName(service string) string {
	switch service {
	case "kafka":
		return "Apache Kafka®"
	case "zookeeper":
		return "Apache ZooKeeper™"
	default:
		return writeServiceName(service)
	}
}

func writeServiceName(service string) string {
	switch service {
	case "kafka-rest":
		return "Kafka REST"
	case "ksql-server":
		return "ksqlDB Server"
	case "zookeeper":
		return "ZooKeeper"
	default:
		service = strings.ReplaceAll(service, "-", " ")
		return cases.Title(language.Und).String(service)
	}
}
