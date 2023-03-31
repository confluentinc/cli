package test

import (
	"os"
	"strings"

	testserver "github.com/confluentinc/cli/test/test-server"
	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestAsyncApiExport() {
	tests := []CLITest{
		{args: "asyncapi export", exitCode: 1, fixture: "asyncapi/no-kafka.golden"},
		{args: "asyncapi export", exitCode: 1, useKafka: "lkc-asyncapi", authKafka: "true", fixture: "asyncapi/no-sr-key.golden"},
		{args: "environment use " + testserver.SRApiEnvId, workflow: true},
		// Spec Generated
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET", fixture: "asyncapi/export-success.golden", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --schema-context dev --file asyncapi-with-context.yaml", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
	}
	resetConfiguration(s.T(), false)
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
	fileNames := []string{"asyncapi-spec.yaml", "asyncapi-with-context.yaml"}
	for _, fileName := range fileNames {
		defer os.Remove(fileName)
		require.FileExists(s.T(), fileName)
		file, err := os.ReadFile(fileName)
		require.NoError(s.T(), err)
		testfile, err := os.ReadFile("test/fixtures/output/asyncapi/" + fileName)
		require.NoError(s.T(), err)
		index1 := strings.Index(string(file), "cluster:")
		index2 := strings.Index(string(file), "confluentSchemaRegistry")
		file1 := string(file[:index1]) + string(file[index2:])
		file1 = strings.ReplaceAll(file1, "\r", "")
		require.Equal(s.T(), string(testfile), file)
	}
}

func (s *CLITestSuite) TestAsyncApiImport() {
	tests := []CLITest{
		{args: "asyncapi import", fixture: "asyncapi/import-err-no-file.golden", exitCode: 1},
		{args: "asyncapi import --file=./test/fixtures/input/asyncapi/asyncapi-spec.yaml", exitCode: 1, fixture: "asyncapi/no-kafka.golden"},
		{args: "asyncapi import --file=./test/fixtures/input/asyncapi/asyncapi-spec.yaml", exitCode: 1, useKafka: "lkc-asyncapi", authKafka: "true", fixture: "asyncapi/no-sr-key.golden"},
	}
	resetConfiguration(s.T(), false)
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAsyncApiImportWithWorkflow() {
	tests := []CLITest{
		{args: "environment use " + testserver.SRApiEnvId, workflow: true},
		{args: "asyncapi import --file=./test/fixtures/input/asyncapi/asyncapi-spec.yaml --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true, fixture: "asyncapi/import-no-overwrite.golden"},
		{args: "asyncapi import --file=./test/fixtures/input/asyncapi/asyncapi-spec.yaml --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --overwrite", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true, fixture: "asyncapi/import-with-overwrite.golden"},
		{args: "asyncapi import --file=./test/fixtures/input/asyncapi/asyncapi-with-context.yaml --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --overwrite", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true, fixture: "asyncapi/import-no-channels.golden"},
		{args: "asyncapi import --file=./test/fixtures/input/asyncapi/asyncapi-create-topic.yaml --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --overwrite", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true, fixture: "asyncapi/import-create-topic.golden"},
	}
	resetConfiguration(s.T(), false)
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
