package test

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func (s *CLITestSuite) TestSDPipeline() {
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	testPipelineSourceCode := filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "pipeline", "test-pipeline.sql")

	tests := []CLITest{
		{args: "pipeline list --help", fixture: "pipeline/list-help.golden"},
		{args: "pipeline list", fixture: "pipeline/list.golden"},
		{args: "pipeline describe --help", fixture: "pipeline/describe-help.golden"},
		{args: "pipeline describe pipe-12345", fixture: "pipeline/describe-pass.golden"},
		{args: "pipeline save --help", fixture: "pipeline/save-help.golden"},
		{args: "pipeline save pipe-12345 --source-code-file /tmp/test-save.sql", fixture: "pipeline/save.golden"},
		{args: "pipeline create --help", fixture: "pipeline/create-help.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345", fixture: "pipeline/create.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription", fixture: "pipeline/create.golden"},
		{args: fmt.Sprintf("pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription --source-code-file %s", testPipelineSourceCode), fixture: "pipeline/create.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription --secret name1=value1 --secret name2=value2", fixture: "pipeline/create.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription --secret a_really_really_really_long_secret_name_but_not_exceeding_64_yet=value1 --secret name2=value2", fixture: "pipeline/create.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription --secret name1=value-with,and= --secret name2=value-with\"and'", fixture: "pipeline/create.golden"},
		// secret value with space (e.g. name="some value") also works but cannot be integration tested, due to cli_test.runCommand() is splitting these args by space character
		{args: "pipeline delete --help", fixture: "pipeline/delete-help.golden"},
		{args: "pipeline delete pipe-12345", fixture: "pipeline/delete.golden"},
		{args: "pipeline activate --help", fixture: "pipeline/activate-help.golden"},
		{args: "pipeline activate pipeline-12345", fixture: "pipeline/activate.golden"},
		{args: "pipeline deactivate --help", fixture: "pipeline/deactivate-help.golden"},
		{args: "pipeline deactivate pipeline-12345", fixture: "pipeline/deactivate.golden"},
		{args: "pipeline deactivate pipeline-12345 --retained-topics topic1", fixture: "pipeline/deactivate.golden"},
		{args: "pipeline update --help", fixture: "pipeline/update-help.golden"},
		{args: "pipeline update pipeline-12345 --name testPipeline", fixture: "pipeline/update.golden"},
		{args: "pipeline update pipeline-12345 --description newDescription", fixture: "pipeline/update.golden"},
		{args: "pipeline update pipeline-12345 --name testPipeline --description newDescription", fixture: "pipeline/update.golden"},
		{args: fmt.Sprintf("pipeline update pipeline-12345 --source-code-file %s", testPipelineSourceCode), fixture: "pipeline/update.golden"},
		{args: "pipeline update pipeline-12345 --secret name1=value-with,and= --secret name2=value-with\"and' --secret name3=", fixture: "pipeline/update.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		tt.useKafka = "lkc-12345"
		s.runIntegrationTest(tt)
	}
}
