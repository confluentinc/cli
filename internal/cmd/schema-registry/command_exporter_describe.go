package schemaregistry

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	describeInfoLabels            = []string{"Name", "Subjects", "SubjectFormat", "ContextType", "Context", "Config"}
	describeInfoHumanRenames      = map[string]string{"SubjectFormat": "Subject Format", "ContextType": "Context Type", "Config": "Remote Schema Registry Configs"}
	describeInfoStructuredRenames = map[string]string{"Name": "name", "Subjects": "subjects", "SubjectFormat": "subject_format", "ContextType": "context_type", "Context": "context", "Config": "config"}
)

func (c *exporterCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the information of the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) describe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	info, _, err := srClient.DefaultApi.GetExporterInfo(ctx, args[0])
	if err != nil {
		return err
	}

	data := &exporterInfoDisplay{
		Name:          info.Name,
		Subjects:      strings.Join(info.Subjects, ", "),
		SubjectFormat: info.SubjectRenameFormat,
		ContextType:   info.ContextType,
		Context:       info.Context,
		Config:        convertMapToString(info.Config),
	}
	return output.DescribeObject(cmd, data, describeInfoLabels, describeInfoHumanRenames, describeInfoStructuredRenames)
}
