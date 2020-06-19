package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"time"

	"github.com/fatih/color"
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
			Short: "Start " + service + ".",
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

	config, err := getConfig(ch, cc, service)
	if err != nil {
		return err
	}

	if err := configService(ch, cc, service, config); err != nil {
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

func startProcess(ch local.ConfluentHome, cc local.ConfluentCurrent, service string) error {
	scriptFile, err := ch.GetScriptFile(service)
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

	for {
		isUp, err := isRunning(cc, service)
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

	return nil
}

func configService(ch local.ConfluentHome, cc local.ConfluentCurrent, service string, config map[string]string) error {
	data, err := ch.GetConfig(service)
	if err != nil {
		return err
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

	return cc.SetConfig(service, data)
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

	for {
		isUp, err := isRunning(cc, service)
		if err != nil {
			return err
		}
		if !isUp {
			break
		}
	}

	if err := cc.RemovePidFile(service); err != nil {
		return err
	}

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
