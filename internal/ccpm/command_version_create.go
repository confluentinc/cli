package ccpm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *pluginCommand) newCreateVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom Connect plugin version.",
		Args:  cobra.NoArgs,
		RunE:  c.createVersion,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new version 1.0.0 of a custom connect plugin.",
				Code: "confluent ccpm plugin version create --plugin plugin-123456 --version 1.0.0 --environment env-abcdef --plugin-file datagen.zip --connector-classes 'io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE'",
			},
			examples.Example{
				Text: "Create a new version 2.1.0 of a custom connect plugin with multiple connector classes and optional fields.",
				Code: "confluent ccpm plugin version create --plugin plugin-123456 --version 2.1.0 --environment env-abcdef --plugin-file datagen.zip --connector-classes 'io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE,io.confluent.kafka.connect.sink.SinkConnector:SINK' --sensitive-properties 'passwords,keys,tokens' --documentation-link 'https://github.com/confluentinc/kafka-connect-datagen'",
			},
		),
	}

	cmd.Flags().String("plugin", "", "Plugin ID.")
	cmd.Flags().String("version", "", "Version of the custom Connect plugin (must comply with SemVer).")
	cmd.Flags().String("plugin-file", "", "Custom plugin ZIP or JAR file.")
	cmd.Flags().StringSlice("connector-classes", nil, "A comma-separated list of connector classes in format 'class_name:type' (e.g., 'io.confluent.kafka.connect.source.SourceConnector:SOURCE').")
	cmd.Flags().StringSlice("sensitive-properties", nil, "A comma-separated list of sensitive configuration property names (e.g., 'passwords,keys,tokens').")
	cmd.Flags().String("documentation-link", "", "URL to the plugin documentation (e.g., 'https://docs.confluent.io').")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))
	cobra.CheckErr(cmd.MarkFlagRequired("plugin-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("connector-classes"))
	cobra.CheckErr(cmd.MarkFlagFilename("plugin-file", "zip", "jar"))

	return cmd
}

func (c *pluginCommand) createVersion(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pluginFile, err := cmd.Flags().GetString("plugin-file")
	if err != nil {
		return err
	}

	// Get plugin details to determine cloud provider
	plugin, err := c.V2Client.DescribeCCPMPlugin(pluginId, environment)
	if err != nil {
		return err
	}

	cloud := plugin.Spec.GetCloud()

	connectorClassesFlag, err := cmd.Flags().GetStringSlice("connector-classes")
	if err != nil {
		return err
	}

	// Parse connector classes
	var connectorClasses []ccpmv1.CcpmV1ConnectorClass //nolint:prealloc
	for _, classStr := range connectorClassesFlag {
		className, connectorType, err := parseConnectorClass(classStr)
		if err != nil {
			return err
		}
		connectorClasses = append(connectorClasses, ccpmv1.CcpmV1ConnectorClass{
			ClassName: className,
			Type:      connectorType,
		})
	}

	// Get optional fields
	sensitivePropertiesFlag, err := cmd.Flags().GetStringSlice("sensitive-properties")
	if err != nil {
		return err
	}

	documentationLink, err := cmd.Flags().GetString("documentation-link")
	if err != nil {
		return err
	}

	// Validate file extension
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(pluginFile), "."))
	if extension != "zip" && extension != "jar" {
		return fmt.Errorf(`only file extensions ".jar" and ".zip" are allowed`)
	}

	// Request presigned URL
	presignedUrlRequest := ccpmv1.CcpmV1PresignedUrl{
		ContentFormat: &extension,
		Cloud:         &cloud,
		Environment:   &ccpmv1.EnvScopedObjectReference{Id: environment},
	}

	resp, err := c.V2Client.CreateCCPMPresignedUrl(presignedUrlRequest)
	if err != nil {
		return err
	}

	// Upload file
	if cloud == "AZURE" {
		if err := utils.UploadFileToAzureBlob(resp.GetUploadUrl(),
			pluginFile, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else if cloud == "GCP" {
		if err := utils.UploadFileToGoogleCloudStorage(resp.GetUploadUrl(),
			pluginFile, strings.ToLower(resp.GetContentFormat())); err != nil {
			return err
		}
	} else {
		if err := utils.UploadFile(resp.GetUploadUrl(),
			pluginFile, resp.GetUploadFormData()); err != nil {
			return err
		}
	}

	// Create version request
	uploadSource := ccpmv1.CcpmV1UploadSourcePresignedUrlAsCcpmV1CustomConnectPluginVersionSpecUploadSourceOneOf(
		&ccpmv1.CcpmV1UploadSourcePresignedUrl{
			Location: "PRESIGNED_URL_LOCATION",
			UploadId: resp.GetUploadId(),
		},
	)

	request := ccpmv1.CcpmV1CustomConnectPluginVersion{
		Spec: &ccpmv1.CcpmV1CustomConnectPluginVersionSpec{
			Version:          &version,
			Environment:      &ccpmv1.EnvScopedObjectReference{Id: environment},
			UploadSource:     &uploadSource,
			ConnectorClasses: &connectorClasses,
		},
	}

	// Add optional fields if provided
	if len(sensitivePropertiesFlag) > 0 {
		request.Spec.SensitiveConfigProperties = &sensitivePropertiesFlag
	}
	if documentationLink != "" {
		request.Spec.DocumentationLink = &documentationLink
	}

	pluginResp, err := c.V2Client.DescribeCCPMPlugin(pluginId, environment)
	if err != nil {
		return err
	}

	// Use V2Client to call CCPM API
	pluginVersion, err := c.V2Client.CreateCCPMPluginVersion(pluginId, request)
	if err != nil {
		return err
	}

	return c.printVersionTable(cmd, pluginResp, pluginVersion)
}

// parseConnectorClass parses a connector class string in format "class_name:type"
func parseConnectorClass(classStr string) (string, string, error) {
	// Split by colon to get class name and type
	parts := strings.Split(classStr, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid connector class format: %s. Expected format: 'class_name:type'", classStr)
	}

	className := strings.TrimSpace(parts[0])
	connectorType := strings.TrimSpace(parts[1])

	// Validate connector type
	if connectorType != "SOURCE" && connectorType != "SINK" {
		return "", "", fmt.Errorf("invalid connector type: %s. Must be either 'SOURCE' or 'SINK'", connectorType)
	}

	return className, connectorType, nil
}

func getConnectorClassesString(connectorClasses []ccpmv1.CcpmV1ConnectorClass) string {
	var classes []string //nolint:prealloc
	for _, cc := range connectorClasses {
		className := cc.GetClassName()
		connectorType := cc.GetType()
		if className == "" || connectorType == "" {
			continue // Skip if class name or type is empty
		}
		// Format as "class_name:type"
		formattedClass := fmt.Sprintf("%s:%s", className, strings.ToUpper(connectorType))
		classes = append(classes, formattedClass)
	}
	return strings.Join(classes, ", ")
}
