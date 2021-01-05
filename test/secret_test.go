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
	testPropertyEntry = []string{"testProperty = test"}
)

func (s *CLITestSuite) TestSecretFile() {
	config, err := ioutil.TempFile("", "config.*.properties")
	require.NoError(s.T(), err)
	err = ioutil.WriteFile(config.Name(), []byte(testPropertyEntry[0]), 0644)
	require.NoError(s.T(), err)

	secrets, err := ioutil.TempFile("", "secrets.*.properties")
	require.NoError(s.T(), err)

	output, err := ioutil.TempFile("", "output.*.properties")
	require.NoError(s.T(), err)

	defer os.Remove(config.Name())
	defer os.Remove(secrets.Name())
	defer os.Remove(output.Name())

	expectedPropertiesForConfig := map[string]string {
		"testProperty" : `${securepass:`+secrets.Name()+`:`+filepath.Base(config.Name())+`/testProperty}`,
		"config.providers" : "securepass",
		"config.providers.securepass.class" : "io.confluent.kafka.security.config.provider.SecurePassConfigProvider",
	}
	expectedPropertiesForSecrets := map[string]string {
		"_metadata.symmetric_key.0.created_at" : "2021-01-04 15:55:08.792508 -0800 PST m=+0.035750432",
		"_metadata.symmetric_key.0.envvar" : "CONFLUENT_SECURITY_MASTER_KEY",
		"_metadata.symmetric_key.0.length" : "32",
		"_metadata.symmetric_key.0.iterations" : "1000",
		"_metadata.symmetric_key.0.salt" : "Hlj1KM4mbIy6LoRwKniYtB5e+yQ3CoUucqh70NGYUfw=",
		"_metadata.symmetric_key.0.enc" : `ENC[AES/CBC/PKCS5Padding,data:YUNiJYhbo0XXKGV3S9oH8OcVMLQH6Rc6uiDvUWO3OWKRJYSJT8dUO3wdsFVqkNpM,iv:/oct96zlQcuy3sgupLE+yQ==,type:str]`,
		"secrets.properties/testProperty" : `ENC[AES/CBC/PKCS5Padding,data:phUW0EYbYpjIV9fbjH37/w==,iv:Ab2ZBOcofe3tOC/DKfYG2w==,type:str]`,
	}

	tests := []CLITest{
		{
			name:	"secret file encrypt",
			env:	[]string{fmt.Sprintf("%s=H0TKtEk7kxBHqxyXtQx0WJaiqsOlaQ89yXZLguEa2yI=", secret.ConfluentKeyEnvVar)},
			args: "secret file encrypt --config-file="+config.Name()+" --local-secrets-file="+ secrets.Name()+" --remote-secrets-file="+ secrets.Name()+" --config testProperty",
			wantFunc: checkEncryptedPropertyValues(config, expectedPropertiesForConfig, secrets, expectedPropertiesForSecrets),
		},
		{
			name:	"secret file decrypt",
			env:	[]string{fmt.Sprintf("%s=H0TKtEk7kxBHqxyXtQx0WJaiqsOlaQ89yXZLguEa2yI=", secret.ConfluentKeyEnvVar)},
			args: "secret file decrypt --config-file="+config.Name()+" --local-secrets-file="+ secrets.Name()+" --output-file="+ output.Name()+" --config testProperty",
			wantFunc: checkDecryptedPropertyValues(output, testPropertyEntry),
		},
	}
	for _, test := range tests {
		test.login = "default"
		s.runConfluentTest(test)
	}
}

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
		}
	}
}

func checkEncryptedPropertyValues(config *os.File, expectedPropertiesForConfig map[string]string, secrets *os.File, expectedPropertiesForSecrets map[string]string) func(t *testing.T) {
	return func(t *testing.T) {
		configFile, err := os.Open(config.Name())
		require.NoError(t, err)
		defer configFile.Close()
		scanner := bufio.NewScanner(configFile)
		for scanner.Scan() {
			line := scanner.Text()
			if equal := strings.Index(line, "="); equal >= 0 {
				require.NotEmpty(t, expectedPropertiesForConfig[line[0:equal-1]])
				require.Equal(t, expectedPropertiesForConfig[line[0:equal-1]], line[equal+2:])
			}
		}

		secretsFile, err := os.Open(secrets.Name())
		require.NoError(t, err)
		defer secretsFile.Close()
		secretsScanner := bufio.NewScanner(secretsFile)
		for secretsScanner.Scan() {
			line := scanner.Text()
			if equal := strings.Index(line, "="); equal >= 0 {
				require.NotEmpty(t, expectedPropertiesForSecrets[line[0:equal-1]])
				require.Equal(t, expectedPropertiesForSecrets[line[0:equal-1]], line[equal+2:])
			}
		}
	}
}
