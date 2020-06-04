package local

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
)

var versionFiles = map[string]string{
	"Confluent Platform":           "share/java/kafka-connect-replicator/connect-replicator-*.jar",
	"Confluent Community Software": "share/java/confluent-common/common-config-*.jar",
	"kafka":                        "share/java/kafka/kafka-clients-*.jar",
	"zookeeper":                    "share/java/kafka/zookeeper-*.jar",
}

func NewVersionCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	versionCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the Confluent Platform version.",
			Args:  cobra.NoArgs,
			RunE:  runVersionCommand,
		},
		cfg, prerunner)

	return versionCommand.Command
}

func runVersionCommand(command *cobra.Command, args []string) error {
	isCP, err := isConfluentPlatform()
	if err != nil {
		return err
	}

	flavor := "Confluent Community Software"
	if isCP {
		flavor = "Confluent Platform"
	}

	version, err := getVersion(flavor)
	if err != nil {
		return err
	}

	cmd.Printf(command, "%s: %s\n", flavor, version)
	return nil
}

// Get the version number of a service based on a trusted file
func getVersion(service string) (string, error) {
	versionFilePattern, ok := versionFiles[service]
	if !ok {
		versionFilePattern = versionFiles["Confluent Platform"]
	}

	matches, err := findConfluentFile(versionFilePattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("could not find %s inside CONFLUENT_HOME", versionFilePattern)
	}

	versionFile := matches[0]
	x := strings.Split(versionFilePattern, "*")
	prefix, suffix := x[0], x[1]
	version := versionFile[len(prefix) : len(versionFile)-len(suffix)]

	return version, nil
}

