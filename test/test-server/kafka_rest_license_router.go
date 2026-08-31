package testserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cpkafkarestv3 "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

// mockLicenseJwt is sized like a real license JWT (~680 characters) so the goldens cover
// how a long, space-free token renders in the list table.
var mockLicenseJwt = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9." +
	strings.Repeat("p", 300) + "." + strings.Repeat("s", 344)

// mockLicenses returns a valid two-slot state: a CP-for-CC license (dated expiry, backing
// topic) and the Developer license (never expires, so expires_at is omitted; no backing
// topic, so topic_name is omitted). This exercises both the present and absent field cases.
func mockLicenses(clusterId string) []cpkafkarestv3.LicenseData {
	return []cpkafkarestv3.LicenseData{
		{
			Kind:              "KafkaLicense",
			ClusterId:         clusterId,
			Category:          "Customer-Managed Confluent Platform for Confluent Cloud",
			CategoryShortName: "cp_for_cc",
			LicenseType:       "ENTERPRISE",
			Status:            "ACTIVE",
			ExpiresAt:         time.Date(2026, time.October, 30, 7, 0, 0, 0, time.UTC),
			Audience:          "test-audience",
			TopicName:         "_confluent-command",
			LicenseJwt:        mockLicenseJwt,
		},
		{
			Kind:              "KafkaLicense",
			ClusterId:         clusterId,
			Category:          "Confluent Platform Developer License",
			CategoryShortName: "cp_developer",
			LicenseType:       "FREE_TIER",
			Status:            "ACTIVE",
			Audience:          "free tier",
			LicenseJwt:        mockLicenseJwt,
		},
	}
}

// licenseToWire mirrors how the REST server serializes a license: `expires_at` is omitted
// entirely when the license never expires. Encoding the SDK struct directly would not do
// that -- `omitempty` has no effect on time.Time, since a struct is never "empty" to
// encoding/json -- so the mock builds the wire form explicitly.
func licenseToWire(license cpkafkarestv3.LicenseData) map[string]any {
	wire := map[string]any{
		"kind":                license.Kind,
		"cluster_id":          license.ClusterId,
		"category":            license.Category,
		"category_short_name": license.CategoryShortName,
		"license_type":        license.LicenseType,
		"status":              license.Status,
		"license_jwt":         license.LicenseJwt,
	}
	if license.Audience != "" {
		wire["audience"] = license.Audience
	}
	if license.TopicName != "" {
		wire["topic_name"] = license.TopicName
	}
	if !license.ExpiresAt.IsZero() {
		wire["expires_at"] = license.ExpiresAt.Format(time.RFC3339)
	}
	return wire
}

func licensesToWire(licenses []cpkafkarestv3.LicenseData) []map[string]any {
	wire := make([]map[string]any, len(licenses))
	for i, license := range licenses {
		wire[i] = licenseToWire(license)
	}
	return wire
}

// Handler for: "/kafka/v3/clusters/{cluster_id}/licenses"
func handleKafkaRestLicenses(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterId := mux.Vars(r)["cluster_id"]

		switch r.Method {
		case http.MethodGet:
			err := json.NewEncoder(w).Encode(map[string]any{
				"kind": "KafkaLicenseList",
				"data": licensesToWire(mockLicenses(clusterId)),
			})
			require.NoError(t, err)
		case http.MethodPut:
			// The CLI always sends dry_run as an explicit query parameter.
			require.Contains(t, []string{"true", "false"}, r.URL.Query().Get("dry_run"))

			var req cpkafkarestv3.UpdateLicenseRequestData
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			require.NotEmpty(t, req.LicenseJwt)

			err := json.NewEncoder(w).Encode(map[string]any{
				"kind":             "KafkaLicenseUpdateResult",
				"updated_category": "Customer-Managed Confluent Platform for Confluent Cloud",
				"licenses":         licensesToWire(mockLicenses(clusterId)),
			})
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
