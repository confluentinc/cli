package connect

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	CONFIG          = "config"
	NAME            = "name"
	CONNECTOR_CLASS = "connector.class"
)

func getConfig(cmd *cobra.Command) (*map[string]string, error) {
	fileName, err := cmd.Flags().GetString(CONFIG)
	if err != nil {
		return nil, errors.Wrap(err, "error reading --config as string")
	}
	var options map[string]string
	options, err = parseConfigFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse config %s", fileName)
	}
	_, nameExists := options[NAME]
	_, classExists := options[CONNECTOR_CLASS]
	if !nameExists || !classExists {
		return nil, errors.Errorf(errors.MissingRequiredConfigsErrorMsg, fileName)
	}
	return &options, nil
}

func parseConfigFile(fileName string) (map[string]string, error) {
	jsonFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read config file %s", fileName)
	}
	if len(jsonFile) == 0 {
		return nil, errors.Errorf(errors.EmptyConfigFileErrorMsg, fileName)
	}

	kvPairs := make(map[string]string)
	var options map[string]interface{}

	err = json.Unmarshal(jsonFile, &options)

	if err != nil {
		return nil, errors.Wrapf(err, errors.ParseConfigErrorMsg, fileName)
	}
	for key, val := range options {
		if val2, ok := val.(string); ok {
			kvPairs[key] = val2
		} else {
			//We support object-as-a-value only for "config" key.
			if key != CONFIG {
				return nil, errors.Errorf("Only string value is permitted for the configuration : %s", key)
			}
			configMap, ok := val.(map[string]interface{})
			if !ok {
				return nil, errors.Errorf("Value for the configuration : %s is malformed", CONFIG)
			}
			for configKey, configVal := range configMap {
				value, isString := configVal.(string)
				if !isString {
					return nil, errors.Errorf("Only string value is permitted for the configuration : %s", configKey)
				}
				kvPairs[configKey] = value
			}
		}
	}

	return kvPairs, err
}
