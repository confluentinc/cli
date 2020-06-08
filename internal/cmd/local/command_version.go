package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
)

var versionFiles = map[string]string{
	"Confluent Platform":           "/share/java/kafka-connect-replicator/connect-replicator-*.jar",
	"Confluent Community Software": "/share/java/confluent-common/common-config-*.jar",
	"kafka":                        "/share/java/kafka/kafka-clients-*.jar",
	"zookeeper":                    "/share/java/kafka/zookeeper-*.jar",
}

func NewVersionCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	versionCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version [service]",
			Short: "Print the Confluent Platform version, or the individual version of a service.",
			Args:  cobra.MaximumNArgs(1),
			RunE:  executeVersionCommand,
		},
		cfg, prerunner)

	return versionCommand.Command
}

func executeVersionCommand(command *cobra.Command, args []string) error {
	if len(args) > 0 {
		service := args[0]

		version, err := getVersion(service)
		if err != nil {
			return err
		}

		cmd.Println(command, version)
		return nil
	}

	flavor := "Confluent Platform"
	version, err := getVersion(flavor)
	if err != nil {
		flavor = "Confluent Community Software"
		version, err = getVersion(flavor)
		if err != nil {
			return fmt.Errorf("could not find Confluent Platform version")
		}
	}

	cmd.Printf(command, "%s: %s\n", flavor, version)
	return nil
}

// Get the version number of a flavor or service based on a trusted file
func getVersion(service string) (string, error) {
	confluentHome := os.Getenv("CONFLUENT_HOME")
	if confluentHome == "" {
		return "", fmt.Errorf("set environment variable CONFLUENT_HOME")
	}

	versionFile, ok := versionFiles[service]
	if !ok {
		return "", fmt.Errorf("unknown service: %s", service)
	}

	pattern := filepath.Join(confluentHome, versionFile)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("could not get version")
	}

	x := strings.Split(pattern, "*")
	prefix, suffix := x[0], x[1]

	file := matches[0]
	version := file[len(prefix) : len(file)-len(suffix)]

	return version, nil
}
