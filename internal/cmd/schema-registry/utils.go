package schemaregistry

import (
	"bytes"
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/confluentinc/go-printer"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/version"
)

const (
	SubjectUsage = "Subject of the schema."
)

func GetApiClient(cmd *cobra.Command, srClient *srsdk.APIClient, cfg *cmd.DynamicConfig, ver *version.Version) (*srsdk.APIClient, context.Context, error) {
	if srClient != nil {
		// Tests/mocks
		return srClient, nil, nil
	}
	return getSchemaRegistryClient(cmd, cfg, ver)
}

func PrintVersions(versions []int32) {
	titleRow := []string{"Version"}
	var entries [][]string
	for _, v := range versions {
		record := &struct{ Version int32 }{v}
		entries = append(entries, printer.ToRow(record, titleRow))
	}
	printer.RenderCollectionTable(entries, titleRow)
}

func convertMapToString(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

func toMap(configs []string) (map[string]string, error) {
	configMap := make(map[string]string)
	for _, cfg := range configs {
		pair := strings.SplitN(cfg, "=", 2)
		if len(pair) < 2 {
			return nil, fmt.Errorf(errors.ConfigurationFormErrorMsg)
		}
		configMap[pair[0]] = pair[1]
	}
	return configMap, nil
}

func readConfigsFromFile(configFile string) (map[string]string, error) {
	if configFile == "" {
		return map[string]string{}, nil
	}

	configContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Create config map from the file
	var configs []string
	for _, s := range strings.Split(string(configContents), "\n") {
		// Filter out blank lines
		spaceTrimmed := strings.TrimSpace(s)
		if s != "" && spaceTrimmed[0] != '#' {
			configs = append(configs, spaceTrimmed)
		}
	}

	return toMap(configs)
}

func RequireSubjectFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	_ = cmd.MarkFlagRequired("subject")
}

func getServiceProviderFromUrl(url string) string {
	if url == "" {
		return ""
	}
	//Endpoint url is of the form https://psrc-<id>.<location>.<service-provider>.<devel/stag/prod/env>.cpdev.cloud
	stringSlice := strings.Split(url, ".")
	if len(stringSlice) != 6 {
		return ""
	}
	return strings.Trim(stringSlice[2], ".")
}
