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
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

var (
	allowedFileExtensions = []string{"jar", "zip"}
)

type artifactCreateOut struct {
	Id            string `human:"ID" serialized:"id"`
	Name          string `human:"Name" serialized:"name"`
	Cloud         string `human:"Cloud" serialized:"cloud"`
	Environment   string `human:"Environment" serialized:"environment"`
	Description   string `human:"Description" serialized:"description"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
}

func (c *artifactCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Connect artifact.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createArtifact,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create connect artifact "my-connect-artifact".`,
				Code: "confluent connect artifact create my-connect-artifact --artifact-file artifact.jar --cloud aws --environment env-abc123 --description newArtifact",
			},
		),
	}

	cmd.Flags().String("artifact-file", "", "Connect artifact JAR file or ZIP file.")
	pcmd.AddCloudAwsFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("description", "", "Specify the connect artifact description.")
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

	resp, err := c.V2Client.GetArtifactPresignedUrl(request)
	if err != nil {
		return err
	}

	if strings.ToLower(cloud) == "aws" {
		if err := utils.UploadFile(resp.GetUploadUrl(), artifactFile, resp.GetUploadFormData()); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("only cloud supported is `AWS`")
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

	table := output.NewTable(cmd)
	table.Add(&artifactCreateOut{
		Name:          artifact.Spec.GetDisplayName(),
		Id:            artifact.GetId(),
		Cloud:         cloud,
		Environment:   environment,
		ContentFormat: artifact.Spec.GetContentFormat(),
		Description:   artifact.Spec.GetDescription(),
	})

	return table.Print()
}
