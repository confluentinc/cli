package schema_registry

import (
	"context"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/go-printer"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
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
