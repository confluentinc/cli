package local

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewServiceCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   service + " [command]",
			Short: "Manage the " + service + " service.",
			Args:  cobra.ExactArgs(1),
		},
		cfg, prerunner)

	serviceCommand.AddCommand(NewServiceLogCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceStartCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceStatusCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceStopCommand(service, prerunner, cfg))
	serviceCommand.AddCommand(NewServiceVersionCommand(service, prerunner, cfg))

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
			Short: "Check the status of " + service + ".",
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
			Short: "Stop " + service + ".",
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

func NewServiceVersionCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceVersionCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the version of " + service + ".",
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

	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	isUp, err := isRunning(service, dir)
	if err != nil {
		return err
	}
	if isUp {
		return nil
	}

	command.Printf("Starting %s\n", service)

	confluentHome, err := getConfluentHome()
	if err != nil {
		return err
	}
	src := filepath.Join(confluentHome, "etc", services[service].properties)
	dst := filepath.Join(dir, fmt.Sprintf("%s.properties", service))
	if err := copyFile(src, dst); err != nil {
		return err
	}

	bin := filepath.Join(confluentHome, "bin", services[service].startCommand)
	startCmd := exec.Command(bin, dst)

	log := filepath.Join(dir, fmt.Sprintf("%s.log", service))
	fd, err := os.Create(log)
	if err != nil {
		return err
	}
	startCmd.Stdout = fd
	startCmd.Stderr = fd

	if err := startCmd.Start(); err != nil {
		return err
	}

	pidFile := getPidFile(service, dir)
	if err := writeInt(pidFile, startCmd.Process.Pid); err != nil {
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

	return printStatus(command, service)
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
		return nil
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

	err = process.Signal(syscall.Signal(0))
	return err == nil, nil
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

	x, err := strconv.Atoi(strings.TrimRight(string(data), "\n"))
	if err != nil {
		return 0, err
	}

	return x, nil
}

func writeInt(file string, x int) error {
	data := []byte(fmt.Sprintf("%d\n", x))
	return ioutil.WriteFile(file, data, 0644)
}

func copyFile(src string, dst string) error {
	srcFd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFd.Close()

	dstFd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFd.Close()

	_, err = io.Copy(dstFd, srcFd)
	return err
}
