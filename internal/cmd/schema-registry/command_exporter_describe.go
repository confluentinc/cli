package schemaregistry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
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

func (c *command) newExporterDescribeCommand(cfg *v1.Config) *cobra.Command {
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
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	info, err := client.GetExporterInfo(args[0])
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
		pairs = append(pairs, fmt.Sprintf(`%s="%s"`, key, value))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "\n")
}
