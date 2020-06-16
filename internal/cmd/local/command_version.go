package local

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/local"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
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

func runVersionCommand(command *cobra.Command, _ []string) error {
	ch := local.NewConfluentHomeManager()

	isCP, err := ch.IsConfluentPlatform()
	if err != nil {
		return err
	}

	flavor := "Confluent Community Software"
	if isCP {
		flavor = "Confluent Platform"
	}

	version, err := getVersion(ch, flavor)
	if err != nil {
		return err
	}

	cmd.Printf(command, "%s: %s\n", flavor, version)
	return nil
}

// Get the version number of a service based on a trusted file
func getVersion(ch local.ConfluentHome, service string) (string, error) {
	pattern, ok := versionFiles[service]
	if !ok {
		pattern = versionFiles["Confluent Platform"]
	}

	matches, err := ch.FindFile(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("could not find %s in CONFLUENT_HOME", pattern)
	}

	versionFile := matches[0]
	x := strings.Split(pattern, "*")
	prefix, suffix := x[0], x[1]
	version := versionFile[len(prefix) : len(versionFile)-len(suffix)]

	return version, nil
}
