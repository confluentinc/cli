package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const (
	subjectUsage              = "Subject of the schema."
	onPremAuthenticationMsg   = "--certificate-authority-path <certification-authority-location> --schema-registry-endpoint <schema-registry-endpoint>"
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

	return "", errors.NewErrorWithSuggestions(
		fmt.Sprintf(`"%s" is an invalid package type`, inputPackageDisplayName),
		fmt.Sprintf("Allowed values for `--package` flag are: %s.", getCommaDelimitedPackagesString()),
	)
}

func getCommaDelimitedPackagesString() string {
	return utils.ArrayToCommaDelimitedString(packageDisplayNames, "or")
}
