package test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func (s *CLITestSuite) TestSDPipeline() {
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	testPipelineSourceCode := filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "pipeline", "test-pipeline.sql")
	testOutputFile, _ := os.CreateTemp("", "test-save.sql")

	tests := []CLITest{
		{args: "pipeline list --help", fixture: "pipeline/list-help.golden"},
		{args: "pipeline list", fixture: "pipeline/list.golden"},
		{args: "pipeline describe --help", fixture: "pipeline/describe-help.golden"},
		{args: "pipeline describe pipe-12345", fixture: "pipeline/describe-pass.golden"},
		{args: "pipeline save --help", fixture: "pipeline/save-help.golden"},
		{args: fmt.Sprintf("pipeline save pipe-12345 --sql-file %s", testOutputFile.Name()), fixture: "pipeline/save.golden", regex: true},
		{args: "pipeline create --help", fixture: "pipeline/create-help.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345", fixture: "pipeline/create.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription", fixture: "pipeline/create.golden"},
		{args: fmt.Sprintf("pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription --sql-file %s", testPipelineSourceCode), fixture: "pipeline/create.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription --secret name1=value1 --secret name2=value-with,and= --secret name3=value-with\"and' --secret a_really_really_really_really_really_really_really_really_really_really_really_really_long_secret_name_but_not_exceeding_128_yet=value", fixture: "pipeline/create-with-secret-names.golden"},
		// secret value with space (e.g. name="some value") also works but cannot be integration tested, due to cli_test.runCommand() is splitting these args by space character
		{args: "pipeline delete --help", fixture: "pipeline/delete-help.golden"},
		{args: "pipeline delete pipe-12345 --force", fixture: "pipeline/delete.golden"},
		{args: "pipeline delete pipe-12345", input: "testPipeline\n", fixture: "pipeline/delete-prompt.golden"},
		{args: "pipeline activate --help", fixture: "pipeline/activate-help.golden"},
		{args: "pipeline activate pipeline-12345", fixture: "pipeline/activate.golden"},
		{args: "pipeline deactivate --help", fixture: "pipeline/deactivate-help.golden"},
		{args: "pipeline deactivate pipeline-12345", fixture: "pipeline/deactivate.golden"},
		{args: "pipeline deactivate pipeline-12345 --retained-topics topic1", fixture: "pipeline/deactivate.golden"},
		{args: "pipeline update --help", fixture: "pipeline/update-help.golden"},
		{args: "pipeline update pipeline-12345 --name newName --description newDescription", fixture: "pipeline/update-with-new-name-and-desc.golden"},
		{args: "pipeline update pipeline-12345 --name newName --description newDescription -o json", fixture: "pipeline/update-with-json-output.golden"},
		{args: "pipeline update pipeline-12345 --activation-privilege", fixture: "pipeline/update-activation-privilege.golden"},
		{args: "pipeline update pipeline-12345 --activation-privilege=true", fixture: "pipeline/update-activation-privilege.golden"},
		{args: "pipeline update pipeline-12345 --activation-privilege=false", fixture: "pipeline/update.golden"},
		{args: fmt.Sprintf("pipeline update pipeline-12345 --sql-file %s", testPipelineSourceCode), fixture: "pipeline/update.golden"},
		{args: "pipeline update pipeline-12345 --secret name1=value-with,and= --secret name2=value-with\"and' --secret name3=", fixture: "pipeline/update-with-secret-names.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		tt.useKafka = "lkc-12345"
		s.runIntegrationTest(tt)
	}
}
