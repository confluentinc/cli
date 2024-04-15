package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseApplicationOptionsFromSlices(t *testing.T) {
	tests := []struct {
		name               string
		configKeys         []string
		configValues       []string
		expectedAppOptions *ApplicationOptions
		expectedError      bool
	}{
		{
			name:          "TestUnequalSliceLengths",
			configKeys:    []string{"key1"},
			configValues:  []string{},
			expectedError: true,
		},
		{
			name:          "TestInvalidKey",
			configKeys:    []string{"key1"},
			configValues:  []string{"value1"},
			expectedError: true,
		},
		{
			name:          "TestUnsupportedType",
			configKeys:    []string{"Context"},
			configValues:  []string{"unsupportedType"},
			expectedError: true,
		},
		{
			name:          "TestBoolParsingError",
			configKeys:    []string{"UnsafeTrace"},
			configValues:  []string{"notABool"},
			expectedError: true,
		},
		{
			name:         "TestParseAppOptionsSuccessfully",
			configKeys:   []string{"UnsafeTrace", "UserAgent", "EnvironmentId", "EnvironmentName", "OrganizationId", "Database", "ComputePoolId", "ServiceAccountId", "Verbose", "LSPBaseURL", "GatewayURL"},
			configValues: []string{"true", "test", "env-123", "test-env", "org-123", "test-database", "lfcp-123", "sa-123", "true", "localhost:8080", "localhost:8000"},
			expectedAppOptions: &ApplicationOptions{
				UnsafeTrace:      true,
				UserAgent:        "test",
				EnvironmentId:    "env-123",
				EnvironmentName:  "test-env",
				OrganizationId:   "org-123",
				Database:         "test-database",
				ComputePoolId:    "lfcp-123",
				ServiceAccountId: "sa-123",
				Verbose:          true,
				LSPBaseURL:       "localhost:8080",
				GatewayURL:       "localhost:8000",
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			appOptions, err := ParseApplicationOptionsFromSlices(test.configKeys, test.configValues)

			if test.expectedError {
				require.Nil(t, appOptions)
				require.Error(t, err)
			} else {
				require.Equal(t, test.expectedAppOptions, appOptions)
				require.Nil(t, err)
			}
		})
	}
}
