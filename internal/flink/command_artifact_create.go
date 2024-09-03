package flink

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	flinkartifactv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-artifact/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

var (
	allowedRuntimeLanguages = []string{"python", "java"}
	allowedFileExtensions   = []string{"jar", "zip"}
)

type artifactCreateOut struct {
	Id            string `human:"ID" serialized:"id"`
	Name          string `human:"Name" serialized:"name"`
	Version       string `human:"Version" serialized:"version"`
	Class         string `human:"Class" serialized:"class"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
	Cloud         string `human:"Cloud" serialized:"cloud"`
	Region        string `human:"Region" serialized:"region"`
	Environment   string `human:"Environment" serialized:"environment"`
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
				Code: "confluent flink artifact create my-flink-artifact --artifact-file plugin.jar --cloud aws --region us-west-2 --environment env-123456",
			},
		),
	}

	cmd.Flags().String("artifact-file", "", "Flink artifact JAR file or ZIP file.")
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("runtime-language", "java", fmt.Sprintf("Specify the Flink artifact runtime language as %s.", utils.ArrayToCommaDelimitedString(allowedRuntimeLanguages, "or")))
	cmd.Flags().String("description", "", "Description of Flink artifact.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("artifact-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagFilename("artifact-file", "zip", "jar"))

	return cmd
}

func (c *command) createArtifact(cmd *cobra.Command, args []string) error {
	displayName := args[0]
	artifactFile, err := cmd.Flags().GetString("artifact-file")
	if err != nil {
		return err
	}
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	runtimeLanguage, err := cmd.Flags().GetString("runtime-language")
	if err != nil {
		return err
	}
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	extension := strings.TrimPrefix(filepath.Ext(artifactFile), ".")
	if !slices.Contains(allowedFileExtensions, strings.ToLower(extension)) {
		return fmt.Errorf("only extensions allowed for `--artifact-file` are %s", utils.ArrayToCommaDelimitedString(allowedFileExtensions, "and"))
	}

	request := flinkartifactv1.ArtifactV1PresignedUrlRequest{
		ContentFormat: flinkartifactv1.PtrString(extension),
		Cloud:         flinkartifactv1.PtrString(cloud),
		Region:        flinkartifactv1.PtrString(region),
	}

	resp, err := c.V2Client.GetFlinkPresignedUrl(request)
	if err != nil {
		return err
	}

	if err := utils.UploadFile(resp.GetUploadUrl(), artifactFile, resp.GetUploadFormData()); err != nil {
		return err
	}

	createArtifactRequest := flinkartifactv1.InlineObject{
		DisplayName: displayName,
		Cloud:       cloud,
		Region:      region,
		Description: flinkartifactv1.PtrString(description),
		UploadSource: flinkartifactv1.InlineObjectUploadSourceOneOf{
			ArtifactV1UploadSourcePresignedUrl: &flinkartifactv1.ArtifactV1UploadSourcePresignedUrl{
				Location: flinkartifactv1.PtrString("PRESIGNED_URL_LOCATION"),
				UploadId: flinkartifactv1.PtrString(resp.GetUploadId()),
			},
		},
		RuntimeLanguage: flinkartifactv1.PtrString(runtimeLanguage),
	}

	plugin, err := c.V2Client.CreateFlinkArtifact(createArtifactRequest)
	if err != nil {
		return err
	}

	var pluginVersion = ""
	if len(plugin.GetVersions()) > 0 {
		pluginVersion = (*plugin.Versions)[0].GetVersion()
	}

	table := output.NewTable(cmd)
	table.Add(&artifactCreateOut{
		Name:          plugin.GetDisplayName(),
		Id:            plugin.GetId(),
		Version:       pluginVersion,
		Class:         plugin.GetClass(),
		Cloud:         cloud,
		Region:        region,
		Environment:   environment,
		ContentFormat: plugin.GetContentFormat(),
	})
	return table.Print()
}
