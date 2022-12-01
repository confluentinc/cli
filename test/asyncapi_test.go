package test

import (
	"fmt"
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
	}
	resetConfiguration(s.T(), false)
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
	s.FileExistsf("./asyncapi-spec.yaml", "Spec file not generated.")
	file, err := os.ReadFile("asyncapi-spec.yaml")
	if err != nil {
		s.Errorf(err, "Cannot read file asyncapi-spec.yaml")
	}
	testfile, _ := os.ReadFile("test/fixtures/output/asyncapi/asyncapi-spec.yaml")
	testfile1 := strings.ReplaceAll(string(testfile), "\r", "")
	index1 := strings.Index(string(file), "cluster:")
	index2 := strings.Index(string(file), "confluentSchemaRegistry")
	file1 := string(file[:index1]) + string(file[index2:])
	file1 = strings.ReplaceAll(file1, "\r", "")
	file1 = strings.ReplaceAll(file1, " ", "")
	if strings.Compare(file1, testfile1) != 0 {
		fmt.Println(file1)
		s.Error(nil, "spec generated does not match the template output file")
	}
	resetConfiguration(s.T(), false)
}
