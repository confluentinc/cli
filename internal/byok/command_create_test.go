package byok

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveKeyVersionFromAzureKeyId(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no version",
			input:    "https://vault-name.vault.azure.net/keys/key-name",
			expected: "https://vault-name.vault.azure.net/keys/key-name",
		},
		{
			name:     "version removed",
			input:    "https://vault-name.vault.azure.net/keys/key-name/00000000000000000000000000000000",
			expected: "https://vault-name.vault.azure.net/keys/key-name",
		},
		{
			name:     "invalid key, valid url",
			input:    "https://thisisnotavalidkey.vault.azure.net/objects0",
			expected: "https://thisisnotavalidkey.vault.azure.net/objects0",
		},
		{
			name:     "invalid key, invalid url",
			input:    "httpsvault.azure.net/objects0",
			expected: "httpsvault.azure.net/objects0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := removeKeyVersionFromAzureKeyId(test.input)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestGcpMetadataCustomRoleName(t *testing.T) {
	t.Run("success, custom role name generated", func(t *testing.T) {
		metadata := gcpPolicyMetadata{
			keyRing: "testKeyRing",
			key:     "testKey",
		}
		customRoleName := metadata.getCustomRoleName()
		assert.Equal(t, "testKeyRing_testKey_custom_kms_role", customRoleName)
	})

	t.Run("success, hyphens replaced", func(t *testing.T) {
		metadata := gcpPolicyMetadata{
			keyRing: "test-key-ring",
			key:     "test-key",
		}
		customRoleName := metadata.getCustomRoleName()
		assert.Equal(t, "test_key_ring_test_key_custom_kms_role", customRoleName)
	})

	t.Run("failure, default role name returned", func(t *testing.T) {
		metadata := gcpPolicyMetadata{
			keyRing: "test&key&ring",
			key:     "test&key",
		}
		customRoleName := metadata.getCustomRoleName()
		assert.Equal(t, "custom_kms_role", customRoleName)
	})
}
