package schemaregistry

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/go-printer"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/version"
)

const (
	SubjectUsage            = "Subject of the schema."
	OnPremAuthenticationMsg = "--ca-location <ca-file-location> --sr-endpoint <schema-registry-endpoint>"
)

func getApiClient(cmd *cobra.Command, srClient *srsdk.APIClient, cfg *dynamicconfig.DynamicConfig, ver *version.Version) (*srsdk.APIClient, context.Context, error) {
	if srClient != nil {
		// Tests/mocks
		return srClient, nil, nil
	}
	return getSchemaRegistryClient(cmd, cfg, ver, "", "")
}

func GetAPIClientWithAPIKey(cmd *cobra.Command, srClient *srsdk.APIClient, cfg *dynamicconfig.DynamicConfig, ver *version.Version, srAPIKey string, srAPISecret string) (*srsdk.APIClient, context.Context, error) {
	if srClient != nil {
		// Tests/mocks
		return srClient, nil, nil
	}
	return getSchemaRegistryClient(cmd, cfg, ver, srAPIKey, srAPISecret)
}

func GetSrApiClientWithToken(cmd *cobra.Command, srClient *srsdk.APIClient, ver *version.Version, mdsToken string) (*srsdk.APIClient, context.Context, error) {
	if srClient != nil {
		// Tests/mocks
		return srClient, nil, nil
	}
	return getSchemaRegistryClientWithToken(cmd, ver, mdsToken)
}

func printVersions(versions []int32) {
	titleRow := []string{"Version"}
	var entries [][]string
	for _, v := range versions {
		record := &struct{ Version int32 }{v}
		entries = append(entries, printer.ToRow(record, titleRow))
	}
	printer.RenderCollectionTable(entries, titleRow)
}

func convertMapToString(m map[string]string) string {
	pairs := make([]string, 0, len(m))
	for key, value := range m {
		pairs = append(pairs, fmt.Sprintf("%s=\"%s\"", key, value))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "\n")
}

func getServiceProviderFromUrl(url string) string {
	if url == "" {
		return ""
	}
	// Endpoint URL is of the form https://psrc-<id>.<location>.<service-provider>.<devel/stag/prod/env>.cpdev.cloud
	stringSlice := strings.Split(url, ".")
	if len(stringSlice) != 6 {
		return ""
	}
	return strings.Trim(stringSlice[2], ".")
}
