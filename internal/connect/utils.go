package connect

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func getConfigAndOffsets(cmd *cobra.Command, isUpdate bool) (*map[string]string, *[]map[string]any, error) {
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, nil, err
	}

	options, offsets, err := parseConfigFile(configFile, isUpdate)
	if err != nil {
		return nil, nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, configFile, err)
	}

	connectorType := options["confluent.connector.type"]
	if connectorType == "" {
		connectorType = "MANAGED"
	}

	_, nameExists := options["name"]
	_, classExists := options["connector.class"]

	if connectorType != "CUSTOM" && (!nameExists || !classExists) {
		return nil, nil, fmt.Errorf(`required configs "name" and "connector.class" missing from connector config file "%s"`, configFile)
	}

	return &options, &offsets, nil
}

func getConfig(cmd *cobra.Command, isUpdate bool) (*map[string]string, error) {

	options, _, err := getConfigAndOffsets(cmd, isUpdate)

	return options, err
}

func parseConfigFile(filename string, isUpdate bool) (map[string]string, []map[string]any, error) {
	jsonFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, filename, err)
	}
	if len(jsonFile) == 0 {
		return nil, nil, fmt.Errorf(`connector config file "%s" is empty`, filename)
	}

	kvPairs := make(map[string]string)
	var options map[string]any
	var offsets []map[string]any

	if err := json.Unmarshal(jsonFile, &options); err != nil {
		return nil, nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, filename, err)
	}

	for key, val := range options {
		if val2, ok := val.(string); ok {
			kvPairs[key] = val2
		} else if key == "config" {
			configMap, ok := val.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf(`value for the configuration "config" is malformed`)
			}

			for configKey, configVal := range configMap {
				value, isString := configVal.(string)
				if !isString {
					return nil, nil, fmt.Errorf(`only string values are permitted for the configuration "%s"`, configKey)
				}
				kvPairs[configKey] = value
			}
		} else if key == "offsets" {
			if isUpdate {
				return nil, nil, fmt.Errorf("offsets are not allowed in configuration file for `confluent connect cluster update`")
			}
			var request *[]map[string]any
			valBytes, err := json.Marshal(val)
			if err != nil {
				return nil, nil, fmt.Errorf(`error while marshalling offsets, value for the configuration "offsets" is malformed`)
			}
			if err := json.Unmarshal(valBytes, &request); err != nil {
				return nil, nil, fmt.Errorf(`error while unmarshalling offsets, value for the configuration "offsets" is malformed`)
			}

			offsets = *request
		} else {
			return nil, nil, fmt.Errorf(`only string values are permitted for the configuration "%s"`, key)
		}
	}

	return kvPairs, offsets, err
}
