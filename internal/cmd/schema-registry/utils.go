package schemaregistry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	SubjectUsage              = "Subject of the schema."
	OnPremAuthenticationMsg   = "--ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>"
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

func convertMapToString(m map[string]string) string {
	pairs := make([]string, 0, len(m))
	for key, value := range m {
		pairs = append(pairs, fmt.Sprintf("%s=\"%s\"", key, value))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "\n")
}

func getPackageDisplayName(packageName string) string {
	return packageDisplayNameMapping[packageName]
}

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
	return utils.ArrayToCommaDelimitedString(packageDisplayNames)
}

func addPackageFlag(cmd *cobra.Command, defaultPackage string) {
	cmd.Flags().String("package", defaultPackage, fmt.Sprintf("Specify the type of Stream Governance package as %s.", getCommaDelimitedPackagesString()))
}
