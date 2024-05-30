package flink

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

var (
	allowedRuntimeLanguages = map[string]any{
		"python": struct{}{},
		"java":   struct{}{},
	}
	allowedFileExtensions = map[string]any{
		"zip": struct{}{},
		"jar": "java",
	}
)

type pluginCreateOut struct {
	Name          string `human:"Name" serialized:"name"`
	PluginId      string `human:"Plugin ID" serialized:"plugin_id"`
	VersionId     string `human:"Version ID" serialized:"version_id"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
	ErrorTrace    string `human:"Error Trace,omitempty" serialized:"error_trace,omitempty"`
}

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Short:   "Create a Flink UDF artifact.",
		Args:    cobra.ExactArgs(1),
		RunE:    c.createArtifact,
		PreRunE: c.validateCreateArtifact,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink artifact "my-flink-artifact".`,
				Code: "confluent flink artifact create my-flink-artifact --artifact-file plugin.jar",
			},
		),
	}

	cmd.Flags().String("artifact-file", "", "Flink artifact file zip, jar or py.")
	cmd.Flags().String("runtime-lang", "java", "Flink artifact language runtime python/java.")
	cmd.Flags().String("description", "", "Description of Flink artifact.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("artifact-file"))
	cobra.CheckErr(cmd.MarkFlagFilename("artifact-file", "zip", "jar", "py"))

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
	runtimeLang, err := cmd.Flags().GetString("runtime-lang")
	if err != nil {
		return err
	}

	extension := strings.TrimPrefix(filepath.Ext(artifactFile), ".")

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
		RuntimeLanguage: connectcustompluginv1.PtrString(runtimeLang),
	}

	plugin, err := c.V2Client.CreateCustomPlugin(createArtifactRequest)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&pluginCreateOut{
		Name:          plugin.GetDisplayName(),
		PluginId:      plugin.GetId(),
		VersionId:     plugin.GetConnectorClass(),
		ContentFormat: plugin.GetContentFormat(),
	})
	return table.Print()
}

func (c *command) validateCreateArtifact(cmd *cobra.Command, args []string) error {
	// validate --runtime-lang param
	runtimeLang := "java"
	if cmd.Flags().Changed("runtime-lang") {
		rl, err := cmd.Flags().GetString("runtime-lang")
		if err != nil {
			return err
		}
		if _, ok := allowedRuntimeLanguages[rl]; !ok {
			return fmt.Errorf("invalid value for --runtime-lang %s. Allowed values are %v", rl, utils.ArrayToCommaDelimitedString(maps.Keys(allowedRuntimeLanguages), "or"))
		}
		runtimeLang = rl
	}

	// validate extension for --artifact-file
	artifactFile, err := cmd.Flags().GetString("artifact-file")
	if err != nil {
		return err
	}
	extension := strings.TrimPrefix(filepath.Ext(artifactFile), ".")
	if extension != "zip" {
		requiredLang, ok := allowedFileExtensions[extension]
		if !ok {
			return fmt.Errorf("only extensions are allowed for --artifact-file are %v", utils.ArrayToCommaDelimitedString(maps.Keys(allowedFileExtensions), "or"))
		}
		if requiredLang != runtimeLang {
			return fmt.Errorf("provided value for --runtime-lang %v and --artifact-file extension %v are mis matched", runtimeLang, extension)
		}
	}
	return nil
}
