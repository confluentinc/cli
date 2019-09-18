package schema_registry

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
)

const (
	SubjectUsage = "Subject of the schema."
)

func GetApiClient(srClient *srsdk.APIClient, ch *pcmd.ConfigHelper) (*srsdk.APIClient, context.Context, error) {
	if srClient != nil {
		// Tests/mocks
		return srClient, nil, nil
	}
	client, ctx, err := SchemaRegistryClient(ch)
	if err != nil {
		return nil, nil, err
	}
	return client, ctx, nil
}

func PrintVersions(versions []int32) {
	titleRow := []string{"Version"}
	var entries [][]string
	for _, version := range versions {
		record := &struct{ Version int32 }{version}
		entries = append(entries, printer.ToRow(record, titleRow))
	}
	printer.RenderCollectionTable(entries, titleRow)
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

func GetSchemaRegistryByAccountId(ctx context.Context, ccClient ccloud.SchemaRegistry, accountId string) (*srv1.SchemaRegistryCluster, error) {
	existingClusters, err := ccClient.GetSchemaRegistryClusters(ctx, &srv1.SchemaRegistryCluster{
		AccountId: accountId,
		Name:      "account schema-registry",
	})
	if err != nil {
		return nil, err
	}
	if len(existingClusters) > 0 {
		return existingClusters[0], nil
	}
	return nil, errors.ErrNoSrEnabled
}

func FormatDescription(description string, cliName string) string {
	return strings.ReplaceAll(description, "{{.CLIName}}", cliName)
}
