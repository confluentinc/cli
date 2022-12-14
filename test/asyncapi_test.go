package test

import (
	"os"
	"strings"

	testserver "github.com/confluentinc/cli/test/test-server"
)

func (s *CLITestSuite) TestAsyncApiExport() {
	tests := []CLITest{
		// No Kafka
		{args: "asyncapi export", wantErrCode: 1},
		// No SR Key Set up
		{args: "asyncapi export", wantErrCode: 1, useKafka: "lkc-asyncapi", authKafka: "true"},
		{args: "environment use " + testserver.SRApiEnvId, wantErrCode: 0, workflow: true},
		// Spec Generated
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET", fixture: "asyncapi/1.golden", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
		{args: "asyncapi export --schema-registry-api-key ASYNCAPIKEY --schema-registry-api-secret ASYNCAPISECRET --schema-context dev --file=asyncapi-with-context.yaml", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
	}
	resetConfiguration(s.T(), false)
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
	fileNames := []string{"asyncapi-spec.yaml", "asyncapi-with-context.yaml"}
	for _, fileName := range fileNames {
		defer os.Remove(fileName)
		s.FileExistsf("./"+fileName, "Spec file not generated.")
		file, err := os.ReadFile(fileName)
		if err != nil {
			s.Errorf(err, "Cannot read file %s", fileName)
		}
		testfile, _ := os.ReadFile("test/fixtures/output/asyncapi/" + fileName)
		index1 := strings.Index(string(file), "cluster:")
		index2 := strings.Index(string(file), "confluentSchemaRegistry")
		file1 := string(file[:index1]) + string(file[index2:])
		file1 = strings.ReplaceAll(file1, "\r", "")
		if strings.Compare(file1, string(testfile)) != 0 {
			s.Error(nil, "spec generated does not match the template output file")
		}
	}

	resetConfiguration(s.T(), false)
}
