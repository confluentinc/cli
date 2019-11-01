package connector

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func FormatDescription(description string, cliName string) string {
	return strings.ReplaceAll(description, "{{.CLIName}}", cliName)
}

func getConfig(cmd *cobra.Command) (map[string]string, error) {
	filename, err := cmd.Flags().GetString("config")

	if err != nil {
		return nil, errors.Wrap(err, "error reading --config as string")
	}
	options := connectv1.Connector{}.UserConfigs
	//if strings.HasPrefix(filename, "@") {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read config file %s", filename)
	}
	fmt.Print(yamlFile)
	err = yaml.Unmarshal(yamlFile, options)
	//} else {
	//	err = yaml.Unmarshal([]byte(filename), options)
	//}

	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse config %s", filename)
	}
	return options, nil
}
