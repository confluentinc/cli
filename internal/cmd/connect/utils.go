package connect

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	config         = "config"
	name           = "name"
	connectorClass = "connector.class"
)

func getConfig(cmd *cobra.Command) (*map[string]string, error) {
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, err
	}

	options, err := parseConfigFile(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse config %s", configFile)
	}

	_, nameExists := options[name]
	_, classExists := options[connectorClass]
	if !nameExists || !classExists {
		return nil, errors.Errorf(errors.MissingRequiredConfigsErrorMsg, configFile)
	}

	return &options, nil
}

func parseConfigFile(fileName string) (map[string]string, error) {
	jsonFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read config file %s", fileName)
	}
	if len(jsonFile) == 0 {
		return nil, errors.Errorf(errors.EmptyConfigFileErrorMsg, fileName)
	}

	kvPairs := make(map[string]string)
	var options map[string]any

	if err := json.Unmarshal(jsonFile, &options); err != nil {
		return nil, errors.Wrapf(err, errors.ParseConfigErrorMsg, fileName)
	}

	for key, val := range options {
		if val2, ok := val.(string); ok {
			kvPairs[key] = val2
		} else {
			// We support object-as-a-value only for "config" key.
			if key != config {
				return nil, errors.Errorf(`only string values are permitted for the configuration "%s"`, key)
			}

			configMap, ok := val.(map[string]any)
			if !ok {
				return nil, errors.Errorf(`value for the configuration "%s" is malformed`, config)
			}

			for configKey, configVal := range configMap {
				value, isString := configVal.(string)
				if !isString {
					return nil, errors.Errorf(`only string values are permitted for the configuration "%s"`, configKey)
				}
				kvPairs[configKey] = value
			}
		}
	}

	return kvPairs, err
}
