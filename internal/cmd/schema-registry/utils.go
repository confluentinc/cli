package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	SubjectUsage              = "Subject of the schema."
	OnPremAuthenticationMsg   = "--ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>"
	essentialsPackage         = "essentials"
	advancedPackage           = "advanced"
	privateNetworkType        = "private"
	publicNetworkType         = "public"
	essentialsPackageInternal = "free"
	advancedPackageInternal   = "paid"
)

var packageDisplayNameMapping = map[string]string{
	essentialsPackageInternal: essentialsPackage,
	advancedPackageInternal:   advancedPackage,
}

var packageDisplayNames = []string{essentialsPackage, advancedPackage}
var networkTypeDisplayNames = []string{privateNetworkType, publicNetworkType}

func getPackageInternalName(inputPackageDisplayName string) (string, error) {
	inputPackageDisplayName = strings.ToLower(inputPackageDisplayName)
	for internalName, displayName := range packageDisplayNameMapping {
		if displayName == inputPackageDisplayName {
			return internalName, nil
		}
	}

	return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SRInvalidPackageTypeErrorMsg, inputPackageDisplayName),
		fmt.Sprintf(errors.SRInvalidPackageSuggestions, getCommaDelimitedPackagesString()))
}

func getNetworkTypeInternal(inputNetworkType string) (ccloudv1.SchemaRegistryNetworkType, error) {
	inputNetworkType = strings.ToLower(inputNetworkType)
	if inputNetworkType == publicNetworkType {
		return ccloudv1.SchemaRegistryNetworkType_SR_PUBLIC, nil
	} else if inputNetworkType == privateNetworkType {
		return ccloudv1.SchemaRegistryNetworkType_SR_PRIVATE, nil
	} else if inputNetworkType == "" {
		return ccloudv1.SchemaRegistryNetworkType_SR_UNKNOWN, nil
	}
	return ccloudv1.SchemaRegistryNetworkType_SR_PUBLIC,
		errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SRInvalidNetworkTypeErrorMsg, inputNetworkType),
			fmt.Sprintf(errors.SRInvalidNetworkTypeSuggestions, getCommaDelimitedNetworkTypeString()))
}

func getCommaDelimitedPackagesString() string {
	return utils.ArrayToCommaDelimitedString(packageDisplayNames, "or")
}

func addPackageFlag(cmd *cobra.Command, defaultPackage string) {
	cmd.Flags().String("package", defaultPackage, fmt.Sprintf("Specify the type of Stream Governance package as %s.", getCommaDelimitedPackagesString()))
}

func getCommaDelimitedNetworkTypeString() string {
	return utils.ArrayToCommaDelimitedString(networkTypeDisplayNames, "or")
}

func addNetworkTypeFlag(cmd *cobra.Command, defaultNetworkType string) {
	cmd.Flags().String("networkType", defaultNetworkType, fmt.Sprintf("Specify the network type as %s.", getCommaDelimitedNetworkTypeString()))
}
