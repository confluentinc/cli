package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type presignedUrlCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newPresignedUrlCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "presigned-url",
		Short: "Manage presigned URLs for Custom Connect Plugin uploads.",
	}

	c := &presignedUrlCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())

	return cmd
}

func (c *presignedUrlCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Request a presigned upload URL for a new Custom Connect Plugin.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
	}

	cmd.Flags().String("content-format", "", "Content format of the Custom Connect Plugin archive (ZIP, JAR).")
	cmd.Flags().String("cloud", "", "Cloud provider where the Custom Connect Plugin archive is uploaded (AWS, GCP, AZURE).")
	cmd.Flags().String("environment", "", "Environment ID.")
	cmd.MarkFlagRequired("content-format")
	cmd.MarkFlagRequired("cloud")
	cmd.MarkFlagRequired("environment")

	return cmd
}

func (c *presignedUrlCommand) create(cmd *cobra.Command, args []string) error {
	contentFormat, err := cmd.Flags().GetString("content-format")
	if err != nil {
		return err
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	// Create presigned URL request
	request := ccpmv1.CcpmV1PresignedUrl{
		ContentFormat: &contentFormat,
		Cloud:         &cloud,
		Environment:   &ccpmv1.EnvScopedObjectReference{Id: environment},
	}

	// Use V2Client to call CCPM API
	presignedUrl, err := c.V2Client.CreateCCPMPresignedUrl(request)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Presigned URL created successfully.\n")
	output.Printf(c.Config.EnableColor, "Upload ID: %s\n", presignedUrl.GetUploadId())
	output.Printf(c.Config.EnableColor, "Upload URL: %s\n", presignedUrl.GetUploadUrl())

	return nil
}
