package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	subjectUsage              = "Subject of the schema."
	onPremAuthenticationMsg   = "--ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>"
	essentialsPackage         = "essentials"
	advancedPackage           = "advanced"
	essentialsPackageInternal = "free"
	advancedPackageInternal   = "paid"
)

var packageDisplayNameMapping = map[string]string{
	essentialsPackageInternal: essentialsPackage,
	advancedPackageInternal:   advancedPackage,
}

var packageDisplayNames = []string{essentialsPackage, advancedPackage}

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

func getCommaDelimitedPackagesString() string {
	return utils.ArrayToCommaDelimitedString(packageDisplayNames, "or")
}

func addPackageFlag(cmd *cobra.Command, defaultPackage string) {
	cmd.Flags().String("package", defaultPackage, fmt.Sprintf("Specify the type of Stream Governance package as %s.", getCommaDelimitedPackagesString()))
}
