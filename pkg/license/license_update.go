package license

import (
	"context"
	"os"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func Update(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, enableColor bool) error {
	jwt, err := readLicenseJwt(cmd)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	opts := &kafkarestv3.UpdateLicenseOpts{DryRun: optional.NewBool(dryRun)}
	result, resp, err := restClient.LicenseV3Api.UpdateLicense(restContext, clusterId, kafkarestv3.UpdateLicenseRequestData{LicenseJwt: jwt}, opts)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	if output.GetFormat(cmd) == output.Human {
		// The server does not always report which category it updated, so fall back to an
		// unqualified message rather than naming a category we were not given.
		switch category := result.UpdatedCategory; {
		case dryRun && category != "":
			output.Printf(enableColor, "Validated license \"%s\". The license was not stored because `--dry-run` was specified.\n", category)
		case dryRun:
			output.Printf(enableColor, "Validated license. The license was not stored because `--dry-run` was specified.\n")
		case category != "":
			output.Printf(enableColor, errors.UpdatedResourceMsg, "license", category)
		default:
			output.Printf(enableColor, "Updated license.\n")
		}
	}

	return PrintUpdateResult(cmd, result)
}

// readLicenseJwt resolves the license from the `--license` flag, which is either the
// JWT-encoded license itself or a path to a file containing it. If the value is an existing
// file, its contents are read; otherwise the value is treated as the license.
func readLicenseJwt(cmd *cobra.Command) (string, error) {
	license, err := cmd.Flags().GetString("license")
	if err != nil {
		return "", err
	}

	if info, err := os.Stat(license); err == nil && !info.IsDir() {
		contents, err := os.ReadFile(license)
		if err != nil {
			return "", err
		}
		license = string(contents)
	}

	license = strings.TrimSpace(license)
	if license == "" {
		return "", errors.NewErrorWithSuggestions(
			"license is empty",
			"Provide a JWT-encoded license, or a path to a file containing one, with `--license`.",
		)
	}

	return license, nil
}
