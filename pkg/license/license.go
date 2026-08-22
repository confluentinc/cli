package license

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type licenseOut struct {
	Category          string `human:"Category" json:"category" yaml:"category"`
	CategoryShortName string `human:"Short Name" json:"category_short_name" yaml:"category_short_name"`
	LicenseType       string `human:"Type" json:"license_type" yaml:"license_type"`
	Status            string `human:"Status" json:"status" yaml:"status"`
	ExpiresAt         string `human:"Expires At,omitempty" json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
	Audience          string `human:"Audience,omitempty" json:"audience,omitempty" yaml:"audience,omitempty"`
	ClusterId         string `human:"Cluster ID" json:"cluster_id" yaml:"cluster_id"`
	TopicName         string `human:"Topic Name,omitempty" json:"topic_name,omitempty" yaml:"topic_name,omitempty"`
	LicenseJwt        string `human:"License JWT" json:"license_jwt" yaml:"license_jwt"`
}

// updateLicenseOut mirrors the REST UpdateLicenseResponseData shape: a top-level
// updated_category wrapping the resulting licenses.
type updateLicenseOut struct {
	UpdatedCategory string        `json:"updated_category,omitempty" yaml:"updated_category,omitempty"`
	Licenses        []*licenseOut `json:"licenses" yaml:"licenses"`
}

func newLicenseOut(license kafkarestv3.LicenseData) *licenseOut {
	var expiresAt string
	if !license.ExpiresAt.IsZero() {
		expiresAt = license.ExpiresAt.Format(time.RFC3339)
	}

	return &licenseOut{
		Category:          license.Category,
		CategoryShortName: license.CategoryShortName,
		LicenseType:       license.LicenseType,
		Status:            license.Status,
		ExpiresAt:         expiresAt,
		Audience:          license.Audience,
		ClusterId:         license.ClusterId,
		TopicName:         license.TopicName,
		LicenseJwt:        license.LicenseJwt,
	}
}

func toLicenseOuts(licenses []kafkarestv3.LicenseData) []*licenseOut {
	outs := make([]*licenseOut, len(licenses))
	for i, license := range licenses {
		outs[i] = newLicenseOut(license)
	}
	return outs
}

func List(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string) error {
	licenses, resp, err := restClient.LicenseV3Api.GetLicenses(restContext, clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, o := range toLicenseOuts(licenses.Data) {
		list.Add(o)
	}
	return list.Print()
}

func Update(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, enableColor bool) error {
	license, err := readLicense(cmd)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	opts := &kafkarestv3.UpdateLicenseOpts{DryRun: optional.NewBool(dryRun)}
	result, resp, err := restClient.LicenseV3Api.UpdateLicense(restContext, clusterId, kafkarestv3.UpdateLicenseRequestData{LicenseJwt: license}, opts)
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

	if output.GetFormat(cmd) != output.Human {
		return output.SerializedOutput(cmd, &updateLicenseOut{
			UpdatedCategory: result.UpdatedCategory,
			Licenses:        toLicenseOuts(result.Licenses),
		})
	}

	list := output.NewList(cmd)
	for _, o := range toLicenseOuts(result.Licenses) {
		list.Add(o)
	}
	return list.Print()
}

// readLicense resolves the license from the `--license` flag, which is either the
// JWT-encoded license itself or a path to a file containing it.
func readLicense(cmd *cobra.Command) (string, error) {
	value, err := cmd.Flags().GetString("license")
	if err != nil {
		return "", err
	}
	return resolveLicense(value)
}

// resolveLicense returns the license JWT from value. If value is an existing file, its
// contents are read; otherwise value is treated as the license itself. The result is
// trimmed and must be non-empty.
func resolveLicense(value string) (string, error) {
	if info, err := os.Stat(value); err == nil && !info.IsDir() {
		contents, err := os.ReadFile(value)
		if err != nil {
			return "", errors.NewErrorWithSuggestions(
				err.Error(),
				"Ensure the file passed to `--license` exists and is readable.",
			)
		}
		value = string(contents)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.NewErrorWithSuggestions(
			"license is empty",
			"Provide a JWT-encoded license, or a path to a file containing one, with `--license`.",
		)
	}

	return value, nil
}
