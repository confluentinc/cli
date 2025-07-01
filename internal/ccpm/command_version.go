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
	Id                string `human:"ID" serialized:"id"`
	Version           string `human:"Version" serialized:"version"`
	ContentFormat     string `human:"Content Format" serialized:"content_format"`
	DocumentationLink string `human:"Documentation Link" serialized:"documentation_link"`
}

func newVersionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage Custom Connect Plugin Versions.",
	}

	c := &versionCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func (c *versionCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <plugin-id>",
		Short: "Create a Custom Connect Plugin Version.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}

	cmd.Flags().String("version", "", "Version of the Custom Connect Plugin (must comply with SemVer).")
	cmd.Flags().String("environment", "", "Environment ID.")
	cmd.Flags().String("upload-id", "", "Upload ID returned by the presigned URL API.")
	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("environment")
	cmd.MarkFlagRequired("upload-id")

	return cmd
}

func (c *versionCommand) create(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

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

	spec, _ := pluginVersion.GetSpecOk()
	output.Printf(c.Config.EnableColor, "Created Custom Connect Plugin Version \"%s\" with ID \"%s\".\n", spec.GetVersion(), pluginVersion.GetId())

	return nil
}

func (c *versionCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <plugin-id> <version-id>",
		Short: "Describe a Custom Connect Plugin Version.",
		Args:  cobra.ExactArgs(2),
		RunE:  c.describe,
	}

	return cmd
}

func (c *versionCommand) describe(cmd *cobra.Command, args []string) error {
	pluginId := args[0]
	versionId := args[1]

	// Use V2Client to call CCPM API
	version, err := c.V2Client.DescribeCCPMPluginVersion(pluginId, versionId)
	if err != nil {
		return err
	}

	// Display version details
	spec, _ := version.GetSpecOk()
	output.Printf(c.Config.EnableColor, "ID: %s\n", version.GetId())
	output.Printf(c.Config.EnableColor, "Version: %s\n", spec.GetVersion())
	output.Printf(c.Config.EnableColor, "Content Format: %s\n", spec.GetContentFormat())
	output.Printf(c.Config.EnableColor, "Documentation Link: %s\n", spec.GetDocumentationLink())

	return nil
}

func (c *versionCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <plugin-id> <version-id>",
		Short: "Delete a Custom Connect Plugin Version.",
		Args:  cobra.ExactArgs(2),
		RunE:  c.delete,
	}

	return cmd
}

func (c *versionCommand) delete(cmd *cobra.Command, args []string) error {
	pluginId := args[0]
	versionId := args[1]

	// Use V2Client to call CCPM API
	err := c.V2Client.DeleteCCPMPluginVersion(pluginId, versionId)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Deleted Custom Connect Plugin Version \"%s\".\n", versionId)

	return nil
}

func (c *versionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <plugin-id>",
		Short: "List Custom Connect Plugin Versions.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.list,
	}

	return cmd
}

func (c *versionCommand) list(cmd *cobra.Command, args []string) error {
	pluginId := args[0]

	// Use V2Client to call CCPM API
	versions, err := c.V2Client.ListCCPMPluginVersions(pluginId)
	if err != nil {
		return err
	}

	// Display results in table format
	table := output.NewTable(cmd)
	for _, version := range versions.GetData() {
		spec, _ := version.GetSpecOk()
		table.Add(&versionOut{
			Id:                version.GetId(),
			Version:           spec.GetVersion(),
			ContentFormat:     spec.GetContentFormat(),
			DocumentationLink: spec.GetDocumentationLink(),
		})
	}

	return table.Print()
}
