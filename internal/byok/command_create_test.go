package byok

import (
	"testing"

	"github.com/stretchr/testify/require"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"
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
	tests := []struct {
		name     string
		keyRing  string
		key      string
		expected string
	}{
		{
			name:     "success, custom role name generated",
			keyRing:  "testKeyRing",
			key:      "testKey",
			expected: "testKeyRing_testKey_custom_kms_role",
		},
		{
			name:     "success, hyphens replaced",
			keyRing:  "test-key-ring",
			key:      "test-key",
			expected: "test_key_ring_test_key_custom_kms_role",
		},
		{
			name:     "failure, unsupported characters return default",
			keyRing:  "test&key&ring",
			key:      "test&key",
			expected: "custom_kms_role",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := gcpPolicyMetadata{
				keyRing: test.keyRing,
				key:     test.key,
			}
			actual := metadata.getCustomRoleName()
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPolicyCommand(t *testing.T) {
	awsRoles := []string{"arn:aws:iam::123456789012:role/test-role"}
	azureKeyID := "https://vault-name.vault.azure.net/keys/key-name"
	azureAppID := "00000000-0000-0000-0000-000000000000"
	gcpKeyID := "projects/proj/locations/loc/keyRings/ring/cryptoKeys/key/cryptoKeyVersions/1"
	gcpGroup := "group@example.com"

	tests := []struct {
		name     string
		key      byokv1.ByokV1Key
		wantErr  bool
		contains string
	}{
		{
			name: "aws with valid role",
			key: byokv1.ByokV1Key{
				Key: &byokv1.ByokV1KeyKeyOneOf{
					ByokV1AwsKey: &byokv1.ByokV1AwsKey{
						Roles: &awsRoles,
					},
				},
			},
			wantErr:  false,
			contains: "arn:aws:iam",
		},
		{
			name: "azure with valid input",
			key: byokv1.ByokV1Key{
				Key: &byokv1.ByokV1KeyKeyOneOf{
					ByokV1AzureKey: &byokv1.ByokV1AzureKey{
						KeyId:         azureKeyID,
						ApplicationId: &azureAppID,
					},
				},
			},
			wantErr:  false,
			contains: "az role assignment create",
		},
		{
			name: "gcp with valid key",
			key: byokv1.ByokV1Key{
				Key: &byokv1.ByokV1KeyKeyOneOf{
					ByokV1GcpKey: &byokv1.ByokV1GcpKey{
						KeyId:         gcpKeyID,
						SecurityGroup: &gcpGroup,
					},
				},
			},
			wantErr:  false,
			contains: "group@example.com",
		},

		{
			name: "unknown key type returns empty string",
			key:  byokv1.ByokV1Key{Key: &byokv1.ByokV1KeyKeyOneOf{}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := getPolicyCommand(test.key)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if test.contains != "" {
					require.Contains(t, actual, test.contains)
				}
			}
		})
	}
}
