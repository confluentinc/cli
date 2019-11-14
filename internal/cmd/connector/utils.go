package connector

import (
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func FormatDescription(description string, cliName string) string {
	return strings.ReplaceAll(description, "{{.CLIName}}", cliName)
}

func getConfig(cmd *cobra.Command) (*map[string]string, error) {
	filename, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, errors.Wrap(err, "error reading --config as string")
	}
	var options map[string]string
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read config file %s", filename)
	}
	err = yaml.Unmarshal(yamlFile, &options)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse config %s", filename)
	}
	if len(options) == 0 {
		return nil, errors.Wrapf(err, "empty config file %s", filename)
	}
	_, nameExists := options["name"]
	_, classExists := options["connector.class"]
	if !nameExists || !classExists {
		return nil, errors.Wrapf(err, "%s does not contain required properties name and connector.class", filename)
	}
	return &options, nil
}
