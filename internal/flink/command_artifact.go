package flink

import (
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/spf13/cobra"
	"strings"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type flinkArtifactOut struct {
	Name          string `human:"Name" serialized:"name" `
	PluginId      string `human:"Plugin ID" serialized:"plugin_id" `
	VersionId     string `human:"Version ID" serialized:"version_id" `
	ContentFormat string `human:"Content Format" serialized:"content_format" `
	RuntimeLang   string `human:"Runtime Language" serialized:"runtime_language" `
}

func (c *command) newArtifactCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "artifact",
		Short:       "Manage Flink UDF artifacts.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func getRuntimeLangAndName(displayName string, suffixes map[string]string) (string, string) {
	for suffix, lang := range suffixes {
		if strings.HasSuffix(displayName, suffix) {
			return lang, strings.TrimSuffix(displayName, suffix)
		}
	}
	return "java", displayName // default values
}

func printTable(cmd *cobra.Command, plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin) error {
	displayName := plugin.GetDisplayName()
	runtimeLang, name := getRuntimeLangAndName(displayName, ccloudv2.FlinkArtifactRuntimeSuffixes)

	table := output.NewTable(cmd)
	table.Add(&flinkArtifactOut{
		Name:          name,
		PluginId:      plugin.GetId(),
		VersionId:     plugin.GetConnectorClass(),
		ContentFormat: plugin.GetContentFormat(),
		RuntimeLang:   runtimeLang,
	})

	return table.Print()
}
