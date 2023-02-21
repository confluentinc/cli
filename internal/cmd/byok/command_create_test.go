package byok

import (
	"testing"
)

func TestRemoveKeyVersionFromAzureKeyId(t *testing.T) {

	// table driven tests
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
			if actual != test.expected {
				t.Errorf("%s: Expected %s, got %s", test.name, test.expected, actual)
			}
		})
	}
}
