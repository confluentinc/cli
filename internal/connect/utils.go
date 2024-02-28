package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

const (
	maxFileSize = 1024 * 1024 * 1024 // 1GB
)

func getConfig(cmd *cobra.Command) (*map[string]string, error) {
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, err
	}

	options, err := parseConfigFile(configFile)
	if err != nil {
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, configFile, err)
	}

	connectorType := options["confluent.connector.type"]
	if connectorType == "" {
		connectorType = "MANAGED"
	}

	_, nameExists := options["name"]
	_, classExists := options["connector.class"]

	if connectorType != "CUSTOM" && (!nameExists || !classExists) {
		return nil, fmt.Errorf(`required configs "name" and "connector.class" missing from connector config file "%s"`, configFile)
	}

	return &options, nil
}

func parseConfigFile(filename string) (map[string]string, error) {
	jsonFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, filename, err)
	}
	if len(jsonFile) == 0 {
		return nil, fmt.Errorf(`connector config file "%s" is empty`, filename)
	}

	kvPairs := make(map[string]string)
	var options map[string]any

	if err := json.Unmarshal(jsonFile, &options); err != nil {
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, filename, err)
	}

	for key, val := range options {
		if val2, ok := val.(string); ok {
			kvPairs[key] = val2
		} else {
			// We support object-as-a-value only for "config" key.
			if key != "config" {
				return nil, fmt.Errorf(`only string values are permitted for the configuration "%s"`, key)
			}

			configMap, ok := val.(map[string]any)
			if !ok {
				return nil, fmt.Errorf(`value for the configuration "config" is malformed`)
			}

			for configKey, configVal := range configMap {
				value, isString := configVal.(string)
				if !isString {
					return nil, fmt.Errorf(`only string values are permitted for the configuration "%s"`, configKey)
				}
				kvPairs[configKey] = value
			}
		}
	}

	return kvPairs, err
}

func UploadFile(url, filePath string, formFields map[string]any) error {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	if fileSize > maxFileSize {
		return fmt.Errorf("File size exceeds the limit. Maximum allowed size is 1GB. Actual Size %d", fileSize)
	}

	for key, value := range formFields {
		if strValue, ok := value.(string); ok {
			err := writer.WriteField(key, strValue)
			if err != nil {
				return err
			}
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

	client := &http.Client{
		Timeout: 20 * time.Minute,
	}

	// Create the HTTP request
	request, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return err
	}
	// Set the Content-Type header to multipart/form-data
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		responseBody, err := ioutil.ReadAll(response.Body)
		return fmt.Errorf("[Response] %s [Error] %v", string(responseBody), err)
	}

	return nil
}
