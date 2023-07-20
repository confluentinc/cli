package schemaregistry

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type exporterOut struct {
	Name          string `human:"Name" serialized:"name"`
	Subjects      string `human:"Subjects" serialized:"subjects"`
	SubjectFormat string `human:"Subject Format" serialized:"subject_format"`
	ContextType   string `human:"Context Type" serialized:"context_type"`
	Context       string `human:"Context" serialized:"context"`
	Config        string `human:"Config" serialized:"config"`
}

func (c *command) newExporterDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterDescribe,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.Config, c.Version)
	if err != nil {
		return err
	}

	return describeExporter(cmd, args[0], srClient, ctx)
}

func describeExporter(cmd *cobra.Command, name string, srClient *srsdk.APIClient, ctx context.Context) error {
	info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, name)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&exporterOut{
		Name:          info.Name,
		Subjects:      strings.Join(info.Subjects, ", "),
		SubjectFormat: info.SubjectRenameFormat,
		ContextType:   info.ContextType,
		Context:       info.Context,
		Config:        convertMapToString(info.Config),
	})
	return table.Print()
}

func convertMapToString(m map[string]string) string {
	pairs := make([]string, 0, len(m))
	for key, value := range m {
		pairs = append(pairs, fmt.Sprintf("%s=\"%s\"", key, value))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "\n")
}
