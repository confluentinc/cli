package connect

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/dghubble/sling"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func getConfig(cmd *cobra.Command) (*map[string]string, error) {
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, err
	}

	options, err := parseConfigFile(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, errors.UnableToReadConfigurationFileErrorMsg, configFile)
	}

	connectorType := options["confluent.connector.type"]
	if connectorType == "" {
		connectorType = "MANAGED"
	}

	_, nameExists := options["name"]
	_, classExists := options["connector.class"]

	if connectorType != "CUSTOM" && (!nameExists || !classExists) {
		return nil, errors.Errorf(errors.MissingRequiredConfigsErrorMsg, configFile)
	}

	return &options, nil
}

func parseConfigFile(filename string) (map[string]string, error) {
	jsonFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, errors.UnableToReadConfigurationFileErrorMsg, filename)
	}
	if len(jsonFile) == 0 {
		return nil, errors.Errorf(errors.EmptyConfigFileErrorMsg, filename)
	}

	kvPairs := make(map[string]string)
	var options map[string]any

	if err := json.Unmarshal(jsonFile, &options); err != nil {
		return nil, errors.Wrapf(err, errors.UnableToReadConfigurationFileErrorMsg, filename)
	}

	for key, val := range options {
		if val2, ok := val.(string); ok {
			kvPairs[key] = val2
		} else {
			// We support object-as-a-value only for "config" key.
			if key != "config" {
				return nil, errors.Errorf(`only string values are permitted for the configuration "%s"`, key)
			}

			configMap, ok := val.(map[string]any)
			if !ok {
				return nil, errors.Errorf(`value for the configuration "config" is malformed`)
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

func uploadFile(url, filePath string, formFields map[string]any) error {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	for key, value := range formFields {
		if strValue, ok := value.(string); ok {
			_ = writer.WriteField(key, strValue)
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	_, err = sling.New().Base(url).Set("Content-Type", writer.FormDataContentType()).Post("").Body(&buffer).ReceiveSuccess(nil)
	if err != nil {
		return err
	}

	return nil
}
