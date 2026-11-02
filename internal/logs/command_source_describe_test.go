package logs

import (
	"testing"

	"github.com/stretchr/testify/require"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"
)

func TestRenderSource(t *testing.T) {
	source := &loggingv1.LoggingV1Source{
		Name:                    "cluster-health",
		DisplayName:             "Cluster health events",
		ResourceType:            "kafka",
		IdPrefix:                loggingv1.PtrString("lkc"),
		CrnElements:             []string{"organization", "environment", "cloud-cluster", "kafka"},
		AllowedScopes:           []string{"ORGANIZATION", "ENVIRONMENT", "RESOURCE"},
		FullTextSearchSupported: true,
		FullTextSearchFields:    &[]string{"message"},
		MaxLookbackDays:         3,
		SearchFields: []loggingv1.LoggingV1SearchFieldDescriptor{
			{Name: "id", Type: "string"},
			{Name: "level", Type: "enum", AllowedValues: &[]string{"INFO", "WARN", "ERROR"}},
		},
		ReturnedFields: []loggingv1.LoggingV1ReturnedFieldDescriptor{
			{Name: "id", Type: "string", DefaultColumn: false, Description: loggingv1.PtrString("Unique identifier.")},
			{Name: "timestamp", Type: "instant", DefaultColumn: true, Description: loggingv1.PtrString("Event time.")},
			{Name: "level", Type: "enum", DefaultColumn: true},
		},
	}

	expected := `Log source:        cluster-health
Display name:      Cluster health events
Resource:          kafka (lkc- IDs; crn: organization/environment/cloud-cluster/kafka)
Scopes:            organization, environment, resource
Max lookback:      3d
Full-text search:  yes (--search-text over: message)

Search fields (--search field=value):
  id     string
  level  enum    INFO, WARN, ERROR

Returned fields (each log record):
  timestamp  instant  *  Event time.
  level      enum     *
  id         string      Unique identifier.

  * = default column in human output
`
	require.Equal(t, expected, renderSource(source))

	minimal := renderSource(&loggingv1.LoggingV1Source{Name: "x", DisplayName: "X", ResourceType: "thing", AllowedScopes: []string{"RESOURCE"}})
	require.Contains(t, minimal, "Resource:          thing\n")
	require.Contains(t, minimal, "Full-text search:  no\n")
	require.NotContains(t, minimal, "Search fields")
	require.NotContains(t, minimal, "Returned fields")
}
