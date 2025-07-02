package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type versionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type versionOut struct {
	Id                        string   `human:"ID" serialized:"id"`
	Version                   string   `human:"Version" serialized:"version"`
	ContentFormat             string   `human:"Content Format" serialized:"content_format"`
	DocumentationLink         string   `human:"Documentation Link" serialized:"documentation_link"`
	SensitiveConfigProperties []string `human:"Sensitive Config Properties" serialized:"sensitive_config_properties"`
	Phase                     string   `human:"Phase" serialized:"phase"`
	ErrorMessage              string   `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
	Environment               string   `human:"Environment" serialized:"environment"`
}

func (c *pluginCommand) newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage Custom Connect Plugin Versions.",
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func (c *versionCommand) printVersionTable(cmd *cobra.Command, version ccpmv1.CcpmV1CustomConnectPluginVersion) error {
	table := output.NewTable(cmd)
	table.Add(&versionOut{
		Id:                        *version.Id,
		Version:                   version.Spec.GetVersion(),
		ContentFormat:             version.Spec.GetContentFormat(),
		DocumentationLink:         version.Spec.GetDocumentationLink(),
		SensitiveConfigProperties: version.Spec.GetSensitiveConfigProperties(),
		Phase:                     version.Status.GetPhase(),
		ErrorMessage:              version.Status.GetErrorMessage(),
		Environment:               version.Spec.GetEnvironment().Id,
	})
	return table.Print()
}
