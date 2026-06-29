package flink

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func TestErrIfCfkManaged(t *testing.T) {
	tests := []struct {
		name                string
		annotations         map[string]string
		expectError         bool
		errorContains       string
		suggestionsContains string
	}{
		{
			name:        "nil annotations are not CFK-managed",
			annotations: nil,
			expectError: false,
		},
		{
			name:        "absent ownership annotation is not CFK-managed",
			annotations: map[string]string{"some.other/annotation": "value"},
			expectError: false,
		},
		{
			name:        "wrong ownership value is not CFK-managed",
			annotations: map[string]string{cfkManagedByAnnotation: "someone-else"},
			expectError: false,
		},
		{
			name: "CFK-managed with namespace and name names the owning custom resource",
			annotations: map[string]string{
				cfkManagedByAnnotation:          cfkManagedByValue,
				cfkManagedByNamespaceAnnotation: "flink-system",
				cfkManagedByNameAnnotation:      "my-statement-cr",
			},
			expectError:         true,
			errorContains:       `Flink SQL statement "my-statement" is managed by Confluent for Kubernetes (CFK)`,
			suggestionsContains: `"flink-system/my-statement-cr"`,
		},
		{
			name: "CFK-managed with only name falls back to the bare custom resource name",
			annotations: map[string]string{
				cfkManagedByAnnotation:     cfkManagedByValue,
				cfkManagedByNameAnnotation: "my-statement-cr",
			},
			expectError:         true,
			suggestionsContains: `"my-statement-cr"`,
		},
		{
			name:                "CFK-managed without owner annotations still blocks and points to kubectl",
			annotations:         map[string]string{cfkManagedByAnnotation: cfkManagedByValue},
			expectError:         true,
			suggestionsContains: "`kubectl edit`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errIfCfkManaged(resource.FlinkStatement, "my-statement", tt.annotations)
			if !tt.expectError {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			if tt.errorContains != "" {
				require.ErrorContains(t, err, tt.errorContains)
			}
			if tt.suggestionsContains != "" {
				errorWithSuggestions, ok := err.(errors.ErrorWithSuggestions)
				require.True(t, ok)
				require.Contains(t, errorWithSuggestions.GetSuggestionsMsg(), tt.suggestionsContains)
			}
		})
	}
}

func TestFlinkApplicationAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		expected map[string]string
	}{
		{
			name:     "metadata without annotations returns nil",
			metadata: map[string]interface{}{"name": "my-app"},
			expected: nil,
		},
		{
			name: "string-valued annotations are extracted",
			metadata: map[string]interface{}{
				"name": "my-app",
				"annotations": map[string]interface{}{
					cfkManagedByAnnotation: cfkManagedByValue,
				},
			},
			expected: map[string]string{cfkManagedByAnnotation: cfkManagedByValue},
		},
		{
			name: "non-string annotation values are skipped",
			metadata: map[string]interface{}{
				"annotations": map[string]interface{}{
					cfkManagedByAnnotation: cfkManagedByValue,
					"replicas":             int64(3),
				},
			},
			expected: map[string]string{cfkManagedByAnnotation: cfkManagedByValue},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := cmfsdk.FlinkApplication{Metadata: tt.metadata}
			require.Equal(t, tt.expected, flinkApplicationAnnotations(application))
		})
	}
}
