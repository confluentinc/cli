package license

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type out struct {
	Category string `human:"Category" json:"category" yaml:"category"`
	// The short name is a REST transport detail -- it exists so the API has a URL-safe key
	// for `self` links. Operators read the display category, so it is serialized only.
	CategoryShortName string `human:"-" json:"category_short_name" yaml:"category_short_name"`
	LicenseType       string `human:"Type" json:"license_type" yaml:"license_type"`
	Status            string `human:"Status" json:"status" yaml:"status"`
	ExpiresAt         string `human:"Expires At,omitempty" json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
	Audience          string `human:"Audience,omitempty" json:"audience,omitempty" yaml:"audience,omitempty"`
	ClusterId         string `human:"-" json:"cluster_id" yaml:"cluster_id"`
	// TopicName is the internal Kafka topic the license is stored in, e.g.
	// `_confluent-command`. The API omits it when it does not apply, so it is passed through
	// as-is and dropped from output when absent.
	TopicName  string `human:"Topic Name,omitempty" json:"topic_name,omitempty" yaml:"topic_name,omitempty"`
	LicenseJwt string `human:"License JWT" json:"license_jwt" yaml:"license_jwt"`
}

// updateOut mirrors the REST UpdateLicenseResponseData shape: a top-level updated_category
// wrapping the resulting licenses.
type updateOut struct {
	UpdatedCategory string `json:"updated_category,omitempty" yaml:"updated_category,omitempty"`
	Licenses        []*out `json:"licenses" yaml:"licenses"`
}

// newOut renders a license for display.
func newOut(license kafkarestv3.LicenseData) *out {
	var expiration string
	if !license.ExpiresAt.IsZero() {
		expiration = license.ExpiresAt.Format(time.RFC3339)
	}

	return &out{
		Category:          license.Category,
		CategoryShortName: license.CategoryShortName,
		LicenseType:       license.LicenseType,
		Status:            license.Status,
		ExpiresAt:         expiration,
		Audience:          license.Audience,
		ClusterId:         license.ClusterId,
		TopicName:         license.TopicName,
		LicenseJwt:        license.LicenseJwt,
	}
}

func toOuts(licenses []kafkarestv3.LicenseData) []*out {
	outs := make([]*out, len(licenses))
	for i, license := range licenses {
		outs[i] = newOut(license)
	}
	return outs
}

func List(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string) error {
	licenses, resp, err := restClient.LicenseV3Api.GetLicenses(restContext, clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, o := range toOuts(licenses.Data) {
		list.Add(o)
	}
	return list.Print()
}

// PrintUpdateResult renders an update response.
func PrintUpdateResult(cmd *cobra.Command, result kafkarestv3.UpdateLicenseResponseData) error {
	if output.GetFormat(cmd) != output.Human {
		return output.SerializedOutput(cmd, &updateOut{
			UpdatedCategory: result.UpdatedCategory,
			Licenses:        toOuts(result.Licenses),
		})
	}

	list := output.NewList(cmd)
	for _, o := range toOuts(result.Licenses) {
		list.Add(o)
	}
	return list.Print()
}
