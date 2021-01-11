package test

import (
	"bufio"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	masterKeySecretKeys = []string{"_metadata.master_key.0.salt", "_metadata.symmetric_key.0.created_at"}
)

func (s *CLITestSuite) TestSecretMasterKeyGenerate() {
	secretsFile, err := ioutil.TempFile("", "secret.properties")
	require.NoError(s.T(), err)
	defer os.Remove(secretsFile.Name())

	tests := []CLITest{
		{
			name:	"secret master-key generate",
			args:    "secret master-key generate --passphrase test --local-secrets-file="+secretsFile.Name(),
			contains: "Save the master key. It cannot be retrieved later.",
			wantFunc: checkMasterKeySecretsFile(secretsFile, masterKeySecretKeys),
		},
	}
	for _, test := range tests {
		test.login = "default"
		s.runConfluentTest(test)
	}
}
// Parses through the secrets file to ensure the necessary property entries exists
func checkMasterKeySecretsFile(file *os.File, keys []string) func(t *testing.T){
	return func(t *testing.T) {
		file, err := os.Open(file.Name())
		require.NoError(t, err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if equal := strings.Index(line, "="); equal >= 0 {
				require.Contains(t, keys, line[0:equal-1])
			}
		}
	}
}


var (
	testPropertyEntries        = []string{"testProperty = test", "addProperty = "}
	expectedConfigPropertyKeys = []string{"config.providers", "config.providers.securepass.class"}
	expectedSecretPropertyKeys = []string{"_metadata.symmetric_key.0.created_at", "_metadata.symmetric_key.0.envvar", "_metadata.symmetric_key.0.length",
		"_metadata.symmetric_key.0.iterations", "_metadata.symmetric_key.0.salt", "_metadata.symmetric_key.0.enc"}

	expectedConfigPropertyValues = map[string]string {
		"config.providers" : "securepass",
		"config.providers.securepass.class" : "io.confluent.kafka.security.config.provider.SecurePassConfigProvider",
	}
	expectedSecretPropertyValues = map[string]string {
		"_metadata.symmetric_key.0.created_at" : "",
		"_metadata.symmetric_key.0.envvar" : "CONFLUENT_SECURITY_MASTER_KEY",
		"_metadata.symmetric_key.0.length" : "32",
		"_metadata.symmetric_key.0.iterations" : "1000",
		"_metadata.symmetric_key.0.salt" : "",
		"_metadata.symmetric_key.0.enc" : "",
	}
)

func (s *CLITestSuite) TestSecretFile() {
	config, err := ioutil.TempFile("", "config.*.properties")
	require.NoError(s.T(), err)
	err = ioutil.WriteFile(config.Name(), []byte(testPropertyEntries[0]), 0644)
	require.NoError(s.T(), err)

	secrets, err := ioutil.TempFile("", "secrets.*.properties")
	require.NoError(s.T(), err)

	output, err := ioutil.TempFile("", "output.*.properties")
	require.NoError(s.T(), err)

	defer os.Remove(config.Name())
	defer os.Remove(secrets.Name())
	defer os.Remove(output.Name())

	encryptTestConfigPropertyKeys := append([]string{"testProperty"}, expectedConfigPropertyKeys...)
	encryptTestConfigPropertyValues := copyMap(expectedConfigPropertyValues)
	encryptTestConfigPropertyValues["testProperty"] = `${securepass:`+secrets.Name()+`:`+filepath.Base(config.Name())+`/testProperty}`
	encryptTestSecretPropertyKeys := append(expectedSecretPropertyKeys, filepath.Base(config.Name())+`/testProperty`)
	encryptTestSecretPropertyValues := copyMap(expectedSecretPropertyValues)
	encryptTestSecretPropertyValues[filepath.Base(config.Name())+`/testProperty`] = ""

	addTestConfigPropertyKeys := append(encryptTestConfigPropertyKeys, "addProperty")
	addTestConfigPropertyValues := copyMap(encryptTestConfigPropertyValues)
	addTestConfigPropertyValues["addProperty"] = `${securepass:`+secrets.Name()+`:`+filepath.Base(config.Name())+`/addProperty}`
	addTestSecretPropertyKeys := append(encryptTestSecretPropertyKeys, filepath.Base(config.Name())+`/addProperty`)
	addTestSecretPropertyValues := copyMap(encryptTestSecretPropertyValues)
	addTestSecretPropertyValues[filepath.Base(config.Name())+`/addProperty`] = ""

	tests := []CLITest{
		{
			name:	"secret file encrypt",
			env:	[]string{fmt.Sprintf("%s=H0TKtEk7kxBHqxyXtQx0WJaiqsOlaQ89yXZLguEa2yI=", secret.ConfluentKeyEnvVar)},
			args: "secret file encrypt --config-file="+config.Name()+" --local-secrets-file="+ secrets.Name()+" --remote-secrets-file="+ secrets.Name()+" --config testProperty",
			wantFunc: checkEncryptedPropertyValues(config, encryptTestConfigPropertyKeys, encryptTestConfigPropertyValues, secrets, encryptTestSecretPropertyKeys, encryptTestSecretPropertyValues),
		},
		{
			name:	"secret file decrypt",
			env:	[]string{fmt.Sprintf("%s=H0TKtEk7kxBHqxyXtQx0WJaiqsOlaQ89yXZLguEa2yI=", secret.ConfluentKeyEnvVar)},
			args: "secret file decrypt --config-file="+config.Name()+" --local-secrets-file="+ secrets.Name()+" --output-file="+ output.Name()+" --config testProperty",
			wantFunc: checkDecryptedPropertyValues(output, testPropertyEntries[:1]),
		},
		{
			name:	"secret file add",
			env:	[]string{fmt.Sprintf("%s=H0TKtEk7kxBHqxyXtQx0WJaiqsOlaQ89yXZLguEa2yI=", secret.ConfluentKeyEnvVar)},
			args: "secret file add --config-file="+config.Name()+" --local-secrets-file="+ secrets.Name()+" --remote-secrets-file="+ secrets.Name()+" --config addProperty",
			wantFunc: checkEncryptedPropertyValues(config, addTestConfigPropertyKeys, addTestConfigPropertyValues, secrets, addTestSecretPropertyKeys, addTestSecretPropertyValues),
		},
		{
			name:	"secret file decrypt after add",
			env:	[]string{fmt.Sprintf("%s=H0TKtEk7kxBHqxyXtQx0WJaiqsOlaQ89yXZLguEa2yI=", secret.ConfluentKeyEnvVar)},
			args: "secret file decrypt --config-file="+config.Name()+" --local-secrets-file="+ secrets.Name()+" --output-file="+ output.Name()+" --config testProperty,addProperty",
			wantFunc: checkDecryptedPropertyValues(output, testPropertyEntries),
		},
	}
	for _, test := range tests {
		test.login = "default"
		s.runConfluentTest(test)
	}
}
// Returns a copy of the original map
func copyMap(original map[string]string) map[string]string {
	copy := make(map[string]string)
	for entry, val := range original {
		copy[entry] = val
	}
	return copy
}
// Parse output file to check decrypted property entries + values
func checkDecryptedPropertyValues(output *os.File, entry []string) func(t *testing.T) {
	return func(t *testing.T) {
		outputFile, err := os.Open(output.Name())
		require.NoError(t, err)
		defer outputFile.Close()
		scanner := bufio.NewScanner(outputFile)
		index := 0
		for scanner.Scan() {
			if index >= len(entry) {
				t.Fatal("index out of bounds; too many entries in decrypted file")
			}
			require.Equal(t, entry[index], scanner.Text())
			index++
		}
	}
}
// Parse files to see that the config and secret property files have the correct entries
func checkEncryptedPropertyValues(config *os.File, configPropertyKeys []string, expectedConfigPropertyValues map[string]string, secrets *os.File, secretPropertyKeys []string, expectedSecretPropertyValues map[string]string) func(t *testing.T) {
	return func(t *testing.T) {
		configFile, err := os.Open(config.Name())
		require.NoError(t, err)
		defer configFile.Close()
		configScanner := bufio.NewScanner(configFile)

		for _, entry := range configPropertyKeys {
			require.True(t, configScanner.Scan())
			line := configScanner.Text()
			equal := strings.Index(line, "=")
			require.Equal(t, entry, line[0:equal-1])
			if entry == "testProperty" || entry == "addProperty" { // to compensate for OS-specific paths
				require.Equal(t, expectedConfigPropertyValues[entry], filepath.ToSlash(line[equal+2:]))
			} else {
				require.Equal(t, expectedConfigPropertyValues[entry], line[equal+2:])
			}
		}

		secretsFile, err := os.Open(secrets.Name())
		require.NoError(t, err)
		defer secretsFile.Close()
		secretsScanner := bufio.NewScanner(secretsFile)

		for _, entry := range secretPropertyKeys {
			require.True(t, secretsScanner.Scan())
			line := secretsScanner.Text()
			equal := strings.Index(line, "=")
			require.Equal(t, entry, line[0:equal-1])
			if expectedSecretPropertyValues[entry] != "" {	// empty strings indicate entry values that are non-deterministic
				require.Equal(t, expectedSecretPropertyValues[entry], line[equal+2:])
			}
		}
	}
}
