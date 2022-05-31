package test

import (
	testserver "github.com/confluentinc/cli/test/test-server"
	"io/ioutil"
	"os"
	str "strings"
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
		//time.Sleep(60 * time.Second)
	}
	s.FileExistsf("./asyncapi-spec.yaml", "Spec file not generated.")
	file, err := os.ReadFile("asyncapi-spec.yaml")
	if err != nil {
		s.Errorf(err, "Cannot read file asyncapi-spec.yaml")
	}
	testfile, _ := ioutil.ReadFile("test/fixtures/output/asyncapi/asyncapi-spec.yaml")
	//fmt.Println(string(testfile))
	index1 := str.Index(string(file), "prod-schemaRegistry:")
	index2 := str.Index(string(file), "confluentSchemaRegistry")
	file1 := string(file[:index1]) + string(file[index2:])
	//fmt.Println(file1)

	if str.Compare(string(testfile), file1) != 0 {
		var err2 error
		s.Errorf(err2, "Spec generated does not match the template output file.")
	}
	resetConfiguration(s.T())

}
