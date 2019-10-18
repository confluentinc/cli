package connector

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ghodss/yaml"

	"github.com/confluentinc/cli/internal/pkg/errors"
	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
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
	if strings.HasPrefix(filename,"@") {
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to read config file %s", filename)
		}

		err = yaml.Unmarshal(yamlFile, options)
	} else {
		err = yaml.Unmarshal([]byte(filename), options)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse config %s", filename)
	}
	return options, nil
}

func (c *command) describeFromId(cmd *cobra.Command, connectorID string) (*connectv1.Connector, error) {

	connector,err := c.client.Describe(context.Background(), &connectv1.Connector{Id: connectorID, AccountId: c.config.Auth.Account.Id} )
	if err!=nil {
		return nil, err
	}
	return connector,nil
}

