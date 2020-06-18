package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
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
		// TODO
	case "schema-registry":
		// TODO
	}

	return serviceCommand.Command
}

func NewServiceLogCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceLogCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "log",
			Short: "Print logs for " + service + ".",
			Args:  cobra.NoArgs,
			RunE:  runServiceLogCommand,
		},
		cfg, prerunner)

	return serviceLogCommand.Command
}

func runServiceLogCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	dir, err := getServiceDir(service)
	if err != nil {
		return err
	}

	log := filepath.Join(dir, fmt.Sprintf("%s.log", service))

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
			Short: "Start " + service + ".",
			Args:  cobra.NoArgs,
			RunE:  runServiceStartCommand,
		},
		cfg, prerunner)

	return serviceVersionCommand.Command
}

func runServiceStartCommand(command *cobra.Command, _ []string) error {
	service := command.Parent().Name()

	if err := notifyConfluentCurrent(command); err != nil {
		return err
	}

	for _, dependency := range services[service].startDependencies {
		if err := startService(command, dependency); err != nil {
			return err
		}
	}

	return startService(command, service)
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
	return printStatus(command, service)
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

	if err := notifyConfluentCurrent(command); err != nil {
		return err
	}

	for _, dependency := range services[service].stopDependencies {
		if err := stopService(command, dependency); err != nil {
			return err
		}
	}

	return stopService(command, service)
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
	service := command.Parent().Name()

	dir, err := getServiceDir(service)
	if err != nil {
		return err
	}

	isUp, err := isRunning(service, dir)
	if err != nil {
		return err
	}
	if !isUp {
		return printStatus(command, service)
	}

	pid, err := readInt(getPidFile(service, dir))
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

	version, err := getVersion(service)
	if err != nil {
		return err
	}

	command.Println(version)
	return nil
}

func startService(command *cobra.Command, service string) error {
	dir, err := getServiceDir(service)
	if err != nil {
		return err
	}

	isUp, err := isRunning(service, dir)
	if err != nil {
		return err
	}
	if isUp {
		return printStatus(command, service)
	}

	config := getServiceConfig(service, dir)
	if err := configService(service, dir, config); err != nil {
		return nil
	}

	command.Printf("Starting %s\n", service)

	confluentHome, err := getConfluentHome()
	if err != nil {
		return err
	}

	bin := filepath.Join(confluentHome, "bin", services[service].startCommand)
	arg := filepath.Join(dir, fmt.Sprintf("%s.properties", service))
	start := exec.Command(bin, arg)

	log := filepath.Join(dir, fmt.Sprintf("%s.stdout", service))
	fd, err := os.Create(log)
	if err != nil {
		return err
	}
	start.Stdout = fd
	start.Stderr = fd

	if err := start.Start(); err != nil {
		return err
	}

	pidFile := getPidFile(service, dir)
	if err := writeInt(pidFile, start.Process.Pid); err != nil {
		return err
	}

	for {
		isUp, err := isRunning(service, dir)
		if err != nil {
			return err
		}
		if isUp {
			break
		}
	}

	for {
		isOpen, err := isPortOpen(service)
		if err != nil {
			return err
		}
		if isOpen {
			break
		}
		time.Sleep(time.Second)
	}

	return printStatus(command, service)
}

func configService(service string, dir string, config map[string]string) error {
	dataDir := filepath.Join(dir, "data")
	if service == "ksql-server" {
		dataDir = filepath.Join(dir, "data", "kafka-streams")
	}
	if err := os.MkdirAll(dataDir, 0777); err != nil {
		return err
	}

	confluentHome, err := getConfluentHome()
	if err != nil {
		return err
	}

	src := filepath.Join(confluentHome, "etc", services[service].properties)
	dst := filepath.Join(dir, fmt.Sprintf("%s.properties", service))

	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	for key, val := range config {
		re := regexp.MustCompile(fmt.Sprintf(`(?m)^(#\s)?%s=.+\n`, key))
		line := []byte(fmt.Sprintf("%s=%s\n", key, val))

		if len(re.FindAll(data, -1)) > 0 {
			data = re.ReplaceAll(data, line)
		} else {
			data = append(data, line...)
		}
	}

	return ioutil.WriteFile(dst, data, 0644)
}

func printStatus(command *cobra.Command, service string) error {
	dir, err := getServiceDir(service)
	if err != nil {
		return err
	}

	isUp, err := isRunning(service, dir)
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

func stopService(command *cobra.Command, service string) error {
	dir, err := getServiceDir(service)
	if err != nil {
		return err
	}

	isUp, err := isRunning(service, dir)
	if err != nil {
		return err
	}
	if !isUp {
		return printStatus(command, service)
	}

	command.Printf("Stopping %s\n", service)

	pidFile := getPidFile(service, dir)
	pid, err := readInt(pidFile)
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

	for {
		isUp, err := isRunning(service, dir)
		if err != nil {
			return err
		}
		if !isUp {
			break
		}
	}

	if err := os.Remove(pidFile); err != nil {
		return err
	}

	return printStatus(command, service)
}

func getServiceDir(service string) (string, error) {
	confluentCurrent, err := getConfluentCurrent()
	if err != nil {
		return "", err
	}

	return filepath.Join(confluentCurrent, service), nil
}

func isRunning(service, dir string) (bool, error) {
	pidFile := getPidFile(service, dir)

	if !fileExists(pidFile) {
		return false, nil
	}

	pid, err := readInt(pidFile)
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

func getPidFile(service, dir string) string {
	return filepath.Join(dir, fmt.Sprintf("%s.pid", service))
}

func fileExists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func readInt(file string) (int, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}

	// TODO: Remove \n once the original local command is removed
	x, err := strconv.Atoi(strings.TrimSuffix(string(data), "\n"))
	if err != nil {
		return 0, err
	}

	return x, nil
}

func writeInt(file string, x int) error {
	// TODO: Remove \n once the original local command is removed
	data := []byte(fmt.Sprintf("%d\n", x))
	return ioutil.WriteFile(file, data, 0644)
}
