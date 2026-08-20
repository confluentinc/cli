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

// readLicenseJwt resolves the license JWT from exactly one of `--license-jwt` or `--license-file`.
func readLicenseJwt(cmd *cobra.Command) (string, error) {
	jwt, err := cmd.Flags().GetString("license-jwt")
	if err != nil {
		return "", err
	}

	path, err := cmd.Flags().GetString("license-file")
	if err != nil {
		return "", err
	}

	if path == "" {
		return jwt, nil
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	jwt = strings.TrimSpace(string(contents))
	if jwt == "" {
		return "", errors.NewErrorWithSuggestions(
			"license file is empty",
			"Ensure the file specified by `--license-file` contains a JWT-encoded license.",
		)
	}

	return jwt, nil
}
