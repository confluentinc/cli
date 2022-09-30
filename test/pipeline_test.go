package test

import (
	"os"
)

func (s *CLITestSuite) TestSDPipeline() {
	createLinkConfigFile := getCreateLinkConfigFile()
	defer os.Remove(createLinkConfigFile)
	tests := []CLITest{
		{args: "pipeline list --help", fixture: "pipeline/list-help.golden"},
		{args: "pipeline list", fixture: "pipeline/list-pass.golden"},
		{args: "pipeline create --help", fixture: "pipeline/create-help.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345", fixture: "pipeline/create-pass.golden"},
		{args: "pipeline create --name testPipeline --ksql-cluster lksqlc-12345 --description testDescription", fixture: "pipeline/create-pass.golden"},
		{args: "pipeline delete --help", fixture: "pipeline/delete-help.golden"},
		{args: "pipeline delete pipe-12345", fixture: "pipeline/delete-pass.golden"},
		{args: "pipeline activate --help", fixture: "pipeline/activate-help.golden"},
		{args: "pipeline activate pipeline-12345", fixture: "pipeline/activate-pass.golden"},
		{args: "pipeline deactivate --help", fixture: "pipeline/deactivate-help.golden"},
		{args: "pipeline deactivate pipeline-12345", fixture: "pipeline/deactivate-pass.golden"},
		{args: "pipeline deactivate pipeline-12345 --retained-topics topic1", fixture: "pipeline/deactivate-pass.golden"},
		{args: "pipeline update --help", fixture: "pipeline/update-help.golden"},
		{args: "pipeline update pipeline-12345 --name testPipeline", fixture: "pipeline/update-pass.golden"},
		{args: "pipeline update pipeline-12345 --description newDescription", fixture: "pipeline/update-pass.golden"},
		{args: "pipeline update pipeline-12345 --name testPipeline --description newDescription", fixture: "pipeline/update-pass.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		tt.useKafka = "lkc-12345"
		s.runIntegrationTest(tt)
	}
}
