package local

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/spf13/cobra"
)

var (
	services = []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
	}
	confluentPlatformServices = []string{
		"control-center",
	}
	allServices = append(services, confluentPlatformServices...)
)

func NewServicesCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	servicesCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "services [command]",
			Short: "Manage all Confluent Platform services.",
			Args:  cobra.MinimumNArgs(1),
		},
		cfg, prerunner)

	servicesCommand.AddCommand(NewServicesListCommand(prerunner, cfg))
	for _, service := range allServices {
		servicesCommand.AddCommand(NewServiceCommand(service, prerunner, cfg))
	}

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

func NewServiceCommand(service string, prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	serviceCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   service + " [command]",
			Short: "Manage the " + service + " service.",
			Args:  cobra.ExactArgs(1),
		},
		cfg, prerunner)

	serviceCommand.AddCommand(NewServiceVersionCommand(service, prerunner, cfg))

	return serviceCommand.Command
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

func runServiceVersionCommand(command *cobra.Command, args []string) error {
	service := command.Parent().Name()

	isValid, err := isValidService(service)
	if err != nil {
		return err
	}
	if !isValid {
		return fmt.Errorf("unknown service: %s", service)
	}

	version, err := getVersion(service)
	if err != nil {
		return err
	}

	command.Println(version)
	return nil
}

func getAvailableServices() ([]string, error) {
	isCP, err := isConfluentPlatform()
	if err != nil {
		return []string{}, err
	}

	if isCP {
		return allServices, nil
	}

	return services, nil
}

func isValidService(service string) (bool, error) {
	availableServices, err := getAvailableServices()
	if err != nil {
		return false, err
	}

	for _, validService := range availableServices {
		if service == validService {
			return true, nil
		}
	}
	return false, nil
}
