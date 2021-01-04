package test

import (
	"bufio"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func (s *CLITestSuite) TestSecretMasterKeyGenerate() {
	secretsFile, err := ioutil.TempFile("", "secret.json")
	require.NoError(s.T(), err)
	defer os.Remove(secretsFile.Name())
	keys := []string{"_metadata.master_key.0.salt", "_metadata.symmetric_key.0.created_at"}
	tests := []CLITest{
		{
			args:    "secret master-key generate --passphrase test --local-secrets-file="+secretsFile.Name(),
			contains: "Save the master key. It cannot be retrieved later.",
			wantFunc: func(t *testing.T) {
				file, err := os.Open(secretsFile.Name())
				require.NoError(t, err)
				defer file.Close()
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					if equal := strings.Index(line, "="); equal >= 0 {
						require.Contains(t, keys, line[0:equal-1])
					}
				}
			},
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runConfluentTest(test)
	}
}
