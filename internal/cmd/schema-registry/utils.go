package schemaregistry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/confluentinc/go-printer"
)

const (
	SubjectUsage            = "Subject of the schema."
	OnPremAuthenticationMsg = "--ca-location <ca-file-location> --sr-endpoint <schema-registry-endpoint>"
)

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
