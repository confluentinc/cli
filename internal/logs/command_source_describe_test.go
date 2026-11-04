package logs

import (
	"testing"

	"github.com/stretchr/testify/require"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"
)

var describeTestSource = &loggingv1.LoggingV1Source{
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

func TestNewSourceDescribeOut(t *testing.T) {
	require.Equal(t, &sourceDescribeOut{
		Name:            "cluster-health",
		DisplayName:     "Cluster health events",
		Resource:        "kafka",
		IdPrefix:        "lkc",
		CrnElements:     "organization/environment/cloud-cluster/kafka",
		Scopes:          "ORGANIZATION, ENVIRONMENT, RESOURCE",
		MaxLookbackDays: 3,
		FullTextSearch:  true,
		FullTextFields:  "message",
	}, newSourceDescribeOut(describeTestSource))

	// A source with no ID prefix, CRN elements or full-text search leaves those values empty.
	minimal := newSourceDescribeOut(&loggingv1.LoggingV1Source{Name: "x", DisplayName: "X", ResourceType: "thing", AllowedScopes: []string{"RESOURCE"}})
	require.Equal(t, &sourceDescribeOut{Name: "x", DisplayName: "X", Resource: "thing", Scopes: "RESOURCE"}, minimal)
}

func TestNewSearchFieldOuts(t *testing.T) {
	require.Equal(t, []*searchFieldOut{
		{Name: "id", Type: "string"},
		{Name: "level", Type: "enum", AllowedValues: "INFO, WARN, ERROR"},
	}, newSearchFieldOuts(describeTestSource))

	require.Empty(t, newSearchFieldOuts(&loggingv1.LoggingV1Source{}))
}

func TestNewReturnedFieldOuts(t *testing.T) {
	// The default columns come first, and each group keeps the API's order.
	require.Equal(t, []*returnedFieldOut{
		{Name: "timestamp", Type: "instant", Default: true, Description: "Event time."},
		{Name: "level", Type: "enum", Default: true},
		{Name: "id", Type: "string", Description: "Unique identifier."},
	}, newReturnedFieldOuts(describeTestSource))

	require.Empty(t, newReturnedFieldOuts(&loggingv1.LoggingV1Source{}))
}
