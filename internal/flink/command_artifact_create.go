package flink

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	flinkartifactv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-artifact/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

var (
	allowedRuntimeLanguages = []string{"python", "java"}
	allowedFileExtensions   = []string{"jar", "zip"}
	allowedDocLinkPattern   = "^$|^(http://|https://).+"
)

type artifactCreateOut struct {
	Id                string `human:"ID" serialized:"id"`
	Name              string `human:"Name" serialized:"name"`
	Version           string `human:"Version" serialized:"version"`
	Cloud             string `human:"Cloud" serialized:"cloud"`
	Region            string `human:"Region" serialized:"region"`
	Environment       string `human:"Environment" serialized:"environment"`
	ContentFormat     string `human:"Content Format" serialized:"content_format"`
	Description       string `human:"Description" serialized:"description"`
	DocumentationLink string `human:"Documentation Link" serialized:"documentation_link"`
	ErrorTrace        string `human:"Error Trace,omitempty" serialized:"error_trace,omitempty"`
}

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <unique-name>",
		Short: "Create a Flink UDF artifact.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createArtifact,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink artifact "my-flink-artifact".`,
				Code: "confluent flink artifact create my-flink-artifact --artifact-file artifact.jar --cloud aws --region us-west-2 --environment env-123456",
			},
			examples.Example{
				Text: `Create Flink artifact "flink-java-artifact".`,
				Code: "confluent flink artifact create my-flink-artifact --artifact-file artifact.jar --cloud aws --region us-west-2 --environment env-123456 --description flinkJavaScalar",
			},
		),
	}

	cmd.Flags().String("artifact-file", "", "Flink artifact JAR file or ZIP file.")
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("runtime-language", "java", fmt.Sprintf("Specify the Flink artifact runtime language as %s.", utils.ArrayToCommaDelimitedString(allowedRuntimeLanguages, "or")))
	cmd.Flags().String("description", "", "Specify the Flink artifact description.")
	cmd.Flags().String("documentation-link", "", "Specify the Flink artifact documentation link.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("artifact-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagFilename("artifact-file", "zip", "jar"))

	return cmd
}

func isValidDockLink(docLink string) bool {
	re := regexp.MustCompile(allowedDocLinkPattern)
	return re.MatchString(docLink)
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
	if _, err = c.V2Client.GetOrgEnvironment(environment); err != nil {
		return fmt.Errorf("environment '%s' not found", environment)
	}
	runtimeLanguage, err := cmd.Flags().GetString("runtime-language")
	if err != nil {
		return err
	}
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}
	documentationLink, err := cmd.Flags().GetString("documentation-link")
	if err != nil {
		return err
	}

	if !isValidDockLink(documentationLink) {
		return fmt.Errorf("invalid documentation link format, must be empty or start with http:// or https://")
	}

	extension := strings.TrimPrefix(filepath.Ext(artifactFile), ".")
	if !slices.Contains(allowedFileExtensions, strings.ToLower(extension)) {
		return fmt.Errorf("only extensions allowed for `--artifact-file` are %s", utils.ArrayToCommaDelimitedString(allowedFileExtensions, "and"))
	}

	request := flinkartifactv1.ArtifactV1PresignedUrlRequest{
		ContentFormat: flinkartifactv1.PtrString(extension),
		Cloud:         flinkartifactv1.PtrString(cloud),
		Region:        flinkartifactv1.PtrString(region),
		Environment:   flinkartifactv1.PtrString(environment),
	}

	resp, err := c.V2Client.GetFlinkPresignedUrl(request)
	if err != nil {
		return err
	}

	if strings.ToLower(cloud) == "azure" {
		if err := utils.UploadFileToAzureBlob(resp.GetUploadUrl(), artifactFile, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else if strings.ToLower(cloud) == "gcp" {
		if err := utils.UploadFileToGoogleCloudStorage(resp.GetUploadUrl(), artifactFile, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else {
		if err := utils.UploadFile(resp.GetUploadUrl(), artifactFile, resp.GetUploadFormData()); err != nil {
			return err
		}
	}

	createArtifactRequest := flinkartifactv1.InlineObject{
		DisplayName:       displayName,
		Cloud:             cloud,
		Region:            region,
		Environment:       environment,
		Description:       flinkartifactv1.PtrString(description),
		DocumentationLink: flinkartifactv1.PtrString(documentationLink),
		UploadSource: flinkartifactv1.InlineObjectUploadSourceOneOf{
			ArtifactV1UploadSourcePresignedUrl: &flinkartifactv1.ArtifactV1UploadSourcePresignedUrl{
				Location: flinkartifactv1.PtrString("PRESIGNED_URL_LOCATION"),
				UploadId: flinkartifactv1.PtrString(resp.GetUploadId()),
			},
		},
		RuntimeLanguage: flinkartifactv1.PtrString(runtimeLanguage),
	}

	artifact, err := c.V2Client.CreateFlinkArtifact(createArtifactRequest)
	if err != nil {
		return err
	}

	var artifactVersion = ""
	if versions := artifact.GetVersions(); len(versions) > 0 {
		artifactVersion = versions[0].GetVersion()
	}

	table := output.NewTable(cmd)
	table.Add(&artifactCreateOut{
		Name:              artifact.GetDisplayName(),
		Id:                artifact.GetId(),
		Version:           artifactVersion,
		Cloud:             cloud,
		Region:            region,
		Environment:       environment,
		ContentFormat:     artifact.GetContentFormat(),
		Description:       artifact.GetDescription(),
		DocumentationLink: artifact.GetDocumentationLink(),
	})
	return table.Print()
}
