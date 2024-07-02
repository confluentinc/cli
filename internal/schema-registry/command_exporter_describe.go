package schemaregistry

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type exporterOut struct {
	Name          string            `human:"Name" serialized:"name"`
	Subjects      string            `human:"Subjects" serialized:"subjects"`
	SubjectFormat string            `human:"Subject Format" serialized:"subject_format"`
	ContextType   string            `human:"Context Type" serialized:"context_type"`
	Context       string            `human:"Context" serialized:"context"`
	Config        map[string]string `human:"Config" serialized:"config"`
}

func (c *command) newExporterDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterDescribe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	info, err := client.GetExporterInfo(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&exporterOut{
		Name:          info.GetName(),
		Subjects:      strings.Join(info.GetSubjects(), ", "),
		SubjectFormat: info.GetSubjectRenameFormat(),
		ContextType:   info.GetContextType(),
		Context:       info.GetContext(),
		Config:        info.GetConfig(),
	})
	return table.Print()
}
