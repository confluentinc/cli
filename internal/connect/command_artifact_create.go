package connect

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	camv1 "github.com/confluentinc/ccloud-sdk-go-v2/cam/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

var (
	allowedFileExtensions = []string{"jar", "zip"}
)

func (c *artifactCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Connect artifact.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createArtifact,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Connect artifact "my-connect-artifact".`,
				Code: "confluent connect artifact create my-connect-artifact --artifact-file artifact.jar --cloud aws --environment env-abc123 --description \"This is my new Connect artifact\"",
			},
		),
	}

	cmd.Flags().String("artifact-file", "", "Connect artifact JAR file or ZIP file.")
	pcmd.AddCloudAwsAzureFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("description", "", "Specify the Connect artifact description.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("artifact-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagFilename("artifact-file", "zip", "jar"))

	return cmd
}

func (c *artifactCommand) createArtifact(cmd *cobra.Command, args []string) error {
	displayName := args[0]
	artifactFile, err := cmd.Flags().GetString("artifact-file")
	if err != nil {
		return err
	}
	cloud, err := cmd.Flags().GetString("cloud")
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
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	extension := strings.TrimPrefix(filepath.Ext(artifactFile), ".")
	if !slices.Contains(allowedFileExtensions, strings.ToLower(extension)) {
		return fmt.Errorf("only extensions allowed for `--artifact-file` are %s", utils.ArrayToCommaDelimitedString(allowedFileExtensions, "and"))
	}

	request := camv1.CamV1PresignedUrlRequest{
		ContentFormat: camv1.PtrString(extension),
		Cloud:         camv1.PtrString(cloud),
		Environment:   camv1.PtrString(environment),
	}

	supportedClouds := []string{"aws", "azure"}
	if !slices.Contains(supportedClouds, strings.ToLower(cloud)) {
		return fmt.Errorf("only clouds supported are `AWS` and `AZURE`")
	}

	resp, err := c.V2Client.GetArtifactPresignedUrl(request)
	if err != nil {
		return err
	}

	if strings.ToLower(cloud) == "azure" {
		if err := utils.UploadFileToAzureBlob(resp.GetUploadUrl(), artifactFile, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else {
		if err := utils.UploadFile(resp.GetUploadUrl(), artifactFile, resp.GetUploadFormData()); err != nil {
			return err
		}
	}

	createArtifactRequest := camv1.CamV1ConnectArtifact{
		Spec: &camv1.CamV1ConnectArtifactSpec{
			DisplayName: displayName,
			Cloud:       cloud,
			Environment: environment,
			Description: camv1.PtrString(description),
			UploadSource: &camv1.CamV1ConnectArtifactSpecUploadSourceOneOf{
				CamV1UploadSourcePresignedUrl: &camv1.CamV1UploadSourcePresignedUrl{
					Location: "PRESIGNED_URL_LOCATION",
					UploadId: resp.GetUploadId(),
				},
			},
		},
	}

	artifact, err := c.V2Client.CreateConnectArtifact(createArtifactRequest)
	if err != nil {
		return err
	}

	return printArtifactTable(cmd, artifact)
}
