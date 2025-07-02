package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *versionCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Custom Connect Plugin Version.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new version 1.0.0 of a custom connect plugin using an upload ID from a presigned URL.",
				Code: "confluent ccpm plugin version create --plugin plugin-123456 --version 1.0.0 --environment env-abcdef --upload-id upload-789012",
			},
			examples.Example{
				Text: "Create a new version 2.1.0 of a custom connect plugin with a patch version.",
				Code: "confluent ccpm plugin version create --plugin plugin-123456 --version 2.1.0 --environment env-abcdef --upload-id upload-345678",
			},
		),
	}

	cmd.Flags().String("plugin", "", "Plugin ID.")
	cmd.Flags().String("version", "", "Version of the Custom Connect Plugin (must comply with SemVer).")
	cmd.Flags().String("environment", "", "Environment ID.")
	cmd.Flags().String("upload-id", "", "Upload ID returned by the presigned URL API.")
	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("upload-id"))

	return cmd
}

func (c *versionCommand) create(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	uploadId, err := cmd.Flags().GetString("upload-id")
	if err != nil {
		return err
	}

	// Create version request
	uploadSource := ccpmv1.CcpmV1UploadSourcePresignedUrlAsCcpmV1CustomConnectPluginVersionSpecUploadSourceOneOf(
		&ccpmv1.CcpmV1UploadSourcePresignedUrl{
			UploadId: uploadId,
		},
	)

	request := ccpmv1.CcpmV1CustomConnectPluginVersion{
		Spec: &ccpmv1.CcpmV1CustomConnectPluginVersionSpec{
			Version:      &version,
			Environment:  &ccpmv1.EnvScopedObjectReference{Id: environment},
			UploadSource: &uploadSource,
		},
	}

	// Use V2Client to call CCPM API
	pluginVersion, err := c.V2Client.CreateCCPMPluginVersion(pluginId, request)
	if err != nil {
		return err
	}
	return c.printVersionTable(cmd, pluginVersion)
}