package test

import (
	"fmt"
	"io/ioutil"
	"strings"

	testserver "github.com/confluentinc/cli/test/test-server"
)

func (s *CLITestSuite) TestAsyncApiExport() {
	tests := []CLITest{
		//No Kafka
		{args: "asyncapi export", wantErrCode: 1},
		//No SR Key Set up
		{args: "asyncapi export", fixture: "asyncapi/1.golden", useKafka: "lkc-asyncapi", authKafka: "true"},
		{args: "environment use " + testserver.SRApiEnvId, wantErrCode: 0, workflow: true},
		//Spec Generated
		{args: "asyncapi export --api-key ASYNCAPIKEY --api-secret ASYNCAPISECRET", fixture: "asyncapi/3.golden", useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
		//With examples
		{args: "asyncapi export --api-key ASYNCAPIKEY --api-secret ASYNCAPISECRET --consume-examples true --file asyncapi-withExamples.yaml", wantErrCode: 0, useKafka: "lkc-asyncapi", authKafka: "true", workflow: true},
	}

	resetConfiguration(s.T())
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
	s.FileExistsf("./asyncapi-spec.yaml", "Spec file not generated.")
	file, err := ioutil.ReadFile("asyncapi-spec.yaml")
	if err != nil {
		s.Errorf(err, "Cannot read file asyncapi-spec.yaml")
	}
	testfile, _ := ioutil.ReadFile("test/fixtures/output/asyncapi/asyncapi-spec.yaml")
	testfile1 := strings.ReplaceAll(string(testfile), "\r", "")
	index1 := strings.Index(string(file), "prod-schemaRegistry:")
	index2 := strings.Index(string(file), "confluentSchemaRegistry")
	file1 := string(file[:index1]) + string(file[index2:])
	file1 = strings.ReplaceAll(file1, "\r", "")
	if strings.Compare(strings.TrimSpace(testfile1), strings.TrimSpace(file1)) != 0 {
		var err2 error
		fmt.Println([]byte(testfile1))
		fmt.Println([]byte(file1))
		s.Errorf(err2, "Spec generated does not match the template output file.")
	}
	resetConfiguration(s.T())

}
