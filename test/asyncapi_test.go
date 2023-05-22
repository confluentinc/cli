package test

import (
	"bytes"
	"os"

	"github.com/stretchr/testify/require"

	testserver "github.com/confluentinc/cli/test/test-server"
)

func (s *CLITestSuite) TestAsyncApiExport() {
	tests := []CLITest{
		{args: "asyncapi export", exitCode: 1, fixture: "asyncapi/no-kafka.golden"},
		{args: "asyncapi export", exitCode: 1, useKafka: "lkc-asyncapi", authKafka: "true", fixture: "asyncapi/no-sr-key.golden"},
		{args: "environment use " + testserver.SRApiEnvId, workflow: true},
		// Spec Generated
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET", fixture: "asyncapi/export-success.golden", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --schema-context dev --file asyncapi-with-context.yaml", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --topics topic1 --file asyncapi-topic-specified.yaml", fixture: "asyncapi/export-topic-specified.golden", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --topics topic2 --file asyncapi-no-topics.yaml", fixture: "asyncapi/export-no-topics.golden", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --topics topic* --file asyncapi-topic-specified.yaml", fixture: "asyncapi/export-topic-specified.golden", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --topics no* --file asyncapi-no-topics.yaml", fixture: "asyncapi/export-no-topics.golden", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
	}
	resetConfiguration(s.T(), false)
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}

	for _, filename := range []string{"asyncapi-spec.yaml", "asyncapi-with-context.yaml", "asyncapi-topic-specified.yaml", "asyncapi-no-topics.yaml"} {
		require.FileExists(s.T(), filename)
		defer os.Remove(filename)

		file, err := os.ReadFile(filename)
		require.NoError(s.T(), err)

		testfile, err := os.ReadFile("test/fixtures/output/asyncapi/" + filename)
		require.NoError(s.T(), err)

		index1 := bytes.Index(file, []byte("cluster:"))
		index2 := bytes.Index(file, []byte("confluentSchemaRegistry"))
		file = append(file[:index1], file[index2:]...)
		file = bytes.ReplaceAll(file, []byte("\r"), []byte(""))
		require.Equal(s.T(), string(testfile), string(file))
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
