package flink

import (
	"fmt"
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

type pluginCreateOut struct {
	Name          string `human:"Name" serialized:"name"`
	PluginId      string `human:"Plugin ID" serialized:"plugin_id"`
	VersionId     string `human:"Version ID" serialized:"version_id"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
	RuntimeLang   string `human:"Runtime Language" serialized:"runtime_language"`
	ErrorTrace    string `human:"Error Trace,omitempty" serialized:"error_trace,omitempty"`
}

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink UDF artifact.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createArtifact,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink artifact "my-flink-artifact".`,
				Code: "confluent flink artifact create my-flink-artifact --artifact-file plugin.jar --runtime-language java",
			},
		),
	}

	cmd.Flags().String("artifact-file", "", "Flink artifact JAR/ZIP file.")
	c.addRuntimeLanguage(cmd)
	cmd.Flags().String("description", "", "Description of Flink artifact.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("artifact-file"))
	cobra.CheckErr(cmd.MarkFlagFilename("artifact-file", "jar", "zip"))
	cobra.CheckErr(cmd.MarkFlagRequired("runtime-language"))

	return cmd
}

func (c *command) createArtifact(cmd *cobra.Command, args []string) error {
	displayName := args[0]
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}
	artifactFile, err := cmd.Flags().GetString("artifact-file")
	if err != nil {
		return err
	}
	runtimeLanguage, err := cmd.Flags().GetString("runtime-language")
	if err != nil {
		return err
	}

	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(artifactFile), "."))
	runtimeLanguage = strings.ToLower(runtimeLanguage)
	if extension != "jar" && extension != "zip" {
		return fmt.Errorf(`only ".jar" and ".zip" file extensions are allowed`)
	}
	if extension == "jar" && runtimeLanguage != "java" {
		return fmt.Errorf(`only "java" runtime language is allowed for ".jar" artifacts`)
	} else if extension == "zip" && runtimeLanguage != "python" {
		return fmt.Errorf(`only "python" runtime language is allowed for ".zip" artifacts`)
	}

	for suffix, lang := range ccloudv2.FlinkArtifactRuntimeSuffixes {
		if runtimeLanguage == lang {
			displayName = fmt.Sprintf("%s%s", displayName, suffix)
			break
		}
	}

	request := connectcustompluginv1.ConnectV1PresignedUrlRequest{
		ContentFormat: connectcustompluginv1.PtrString(extension),
	}

	resp, err := c.V2Client.GetPresignedUrl(request)
	if err != nil {
		return err
	}

	if err := utils.UploadFile(resp.GetUploadUrl(), artifactFile, resp.GetUploadFormData()); err != nil {
		return err
	}

	createArtifactRequest := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
		DisplayName:   connectcustompluginv1.PtrString(displayName),
		Description:   connectcustompluginv1.PtrString(description),
		ConnectorType: connectcustompluginv1.PtrString("flink_udf"),
		UploadSource: &connectcustompluginv1.ConnectV1CustomConnectorPluginUploadSourceOneOf{
			ConnectV1UploadSourcePresignedUrl: &connectcustompluginv1.ConnectV1UploadSourcePresignedUrl{
				Location: "PRESIGNED_URL_LOCATION",
				UploadId: resp.GetUploadId(),
			},
		},
	}

	plugin, err := c.V2Client.CreateCustomPlugin(createArtifactRequest)
	if err != nil {
		return err
	}

	displayName = plugin.GetDisplayName()
	runtimeLang, name := getRuntimeLangAndName(displayName, ccloudv2.FlinkArtifactRuntimeSuffixes)

	table := output.NewTable(cmd)
	table.Add(&pluginCreateOut{
		Name:          name,
		PluginId:      plugin.GetId(),
		VersionId:     plugin.GetConnectorClass(),
		ContentFormat: plugin.GetContentFormat(),
		RuntimeLang:   runtimeLang,
	})
	return table.Print()
}

func (c *command) addRuntimeLanguage(cmd *cobra.Command) {
	cmd.Flags().String("runtime-language", "", fmt.Sprintf("Specify the artifact runtime language as %s.", utils.ArrayToCommaDelimitedString(ccloudv2.SupportFlinkArtifactRuntime, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "runtime-language", func(_ *cobra.Command, _ []string) []string { return ccloudv2.SupportFlinkArtifactRuntime })
}
