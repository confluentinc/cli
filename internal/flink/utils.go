package flink

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	perrors "github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

const cmfEnvironmentConfigPath = "~/.cmf/environment"

// getEnvironment returns the environment from the flag or the config file.
func getEnvironment(cmd *cobra.Command) (string, error) {
	envEmptyErr := perrors.NewErrorWithSuggestions("environment is required.", "Specify an environment with `--environment` or set the default environment using `confluent flink environment use`.")

	// Check if someone has passed the environment flag
	environment, err := cmd.Flags().GetString("environment")
	if err == nil && environment != "" {
		return environment, nil
	}
	// check if the environment is set in the config file
	cmfConfigFilePath := expandHomeDir(cmfEnvironmentConfigPath)
	if _, err := os.Stat(cmfConfigFilePath); os.IsNotExist(err) {
		// Don't return the "ErrNotExist" error as it's not relevant, just return the empty environment error.
		return "", envEmptyErr
	}
	data, err := os.ReadFile(cmfConfigFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read environment from config file %s: %w", cmfConfigFilePath, err)
	}

	trimmedEnv := strings.TrimSpace(string(data))
	if trimmedEnv == "" {
		return "", envEmptyErr
	}
	output.Printf(false, "Using environment from config file: %s\n", trimmedEnv)
	return trimmedEnv, nil
}

func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}

	return path
}

// Returns the next page number and whether we need to fetch more pages or not.
func extractPageOptions(receivedItemsLength int, currentPageNumber int) (int, bool) {
	if receivedItemsLength == 0 {
		return currentPageNumber, true
	}
	return currentPageNumber + 1, false
}

// Creates a rich error message from the HTTP response and the SDK error if possible.
func parseSdkError(httpResp *http.Response, sdkErr error) error {
	// If there's an error, and the httpResp is populated, it may contain a more detailed error message.
	// If there's nothing in the response body, we'll return the status.
	if sdkErr != nil && httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
			respBody, parseError := io.ReadAll(httpResp.Body)
			trimmedBody := strings.TrimSpace(string(respBody))
			if parseError == nil && len(trimmedBody) > 0 {
				return errors.New(trimmedBody)
			} else if httpResp.Status != "" {
				return errors.New(httpResp.Status)
			}
		}
	}
	// In case we can't parse the body, or if there's no body at all, return the original error.
	return sdkErr
}
