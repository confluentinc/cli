package test

func (s *CLITestSuite) TestKafkaCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "kafka cluster --help", fixture: "kafka-cluster-help.golden"},
		{args: "environment use a-595", fixture: "kafka0.golden", wantErrCode: 0},
		{args: "kafka cluster list", fixture: "kafka6.golden", wantErrCode: 0},
		{args: "kafka cluster list -o json", fixture: "kafka7.golden", wantErrCode: 0},
		{args: "kafka cluster list -o yaml", fixture: "kafka8.golden", wantErrCode: 0},

		{args: "kafka cluster create", fixture: "kafka1.golden", wantErrCode: 1},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone", fixture: "kafka2.golden", wantErrCode: 0},
		{args: "kafka cluster create my-failed-cluster --cloud oops --region us-east1 --availability single-zone", fixture: "kafka-cloud-provider-error.golden", wantErrCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region oops --availability single-zone", fixture: "kafka-cloud-region-error.golden", wantErrCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region us-east-1 --availability single-zone --type oops", fixture: "kafka-create-type-error.golden", wantErrCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region us-east-1 --availability single-zone --type dedicated --cku 0", fixture: "kafka-cku-error.golden", wantErrCode: 1},
		{args: "kafka cluster create my-dedicated-cluster --cloud aws --region us-east-1 --type dedicated --cku 1", fixture: "kafka22.golden", wantErrCode: 0},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone -o json", fixture: "kafka23.golden", wantErrCode: 0},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone -o yaml", fixture: "kafka24.golden", wantErrCode: 0},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability oops-zone", fixture: "kafka-availability-zone-error.golden", wantErrCode: 1},

		{args: "kafka cluster update lkc-update ", fixture: "kafka-create-flag-error.golden", wantErrCode: 1},
		{args: "kafka cluster update lkc-update --name lkc-update-name", fixture: "kafka26.golden", wantErrCode: 0},
		{args: "kafka cluster update lkc-update --name lkc-update-name -o json", fixture: "kafka28.golden", wantErrCode: 0},
		{args: "kafka cluster update lkc-update --name lkc-update-name -o yaml", fixture: "kafka29.golden", wantErrCode: 0},
		{args: "kafka cluster update lkc-update-dedicated --name lkc-update-dedicated-name --cku 2", fixture: "kafka27.golden", wantErrCode: 0},
		{args: "kafka cluster update lkc-update-dedicated --cku 2", fixture: "kafka39.golden", wantErrCode: 0},
		{args: "kafka cluster update lkc-update --cku 2", fixture: "kafka-cluster-expansion-error.golden", wantErrCode: 1},

		{args: "kafka cluster delete", fixture: "kafka3.golden", wantErrCode: 1},
		{args: "kafka cluster delete lkc-unknown", fixture: "kafka-delete-unknown-error.golden", wantErrCode: 1},
		{args: "kafka cluster delete lkc-def973", fixture: "kafka5.golden", wantErrCode: 0},

		{args: "kafka region list", fixture: "kafka14.golden", wantErrCode: 0},
		{args: "kafka region list -o json", fixture: "kafka15.golden", wantErrCode: 0},
		{args: "kafka region list -o json", fixture: "kafka16.golden", wantErrCode: 0},
		{args: "kafka region list --cloud gcp", fixture: "kafka9.golden", wantErrCode: 0},
		{args: "kafka region list --cloud aws", fixture: "kafka10.golden", wantErrCode: 0},
		{args: "kafka region list --cloud azure", fixture: "kafka11.golden", wantErrCode: 0},

		{args: "kafka cluster describe lkc-describe", fixture: "kafka17.golden", wantErrCode: 0},
		{args: "kafka cluster describe lkc-describe -o json", fixture: "kafka18.golden", wantErrCode: 0},
		{args: "kafka cluster describe lkc-describe -o yaml", fixture: "kafka19.golden", wantErrCode: 0},

		{args: "kafka cluster describe lkc-describe-dedicated", fixture: "kafka30.golden", wantErrCode: 0},
		{args: "kafka cluster describe lkc-describe-dedicated -o json", fixture: "kafka31.golden", wantErrCode: 0},
		{args: "kafka cluster describe lkc-describe-dedicated -o yaml", fixture: "kafka32.golden", wantErrCode: 0},

		{args: "kafka cluster describe lkc-describe-dedicated-pending", fixture: "kafka33.golden", wantErrCode: 0},
		{args: "kafka cluster describe lkc-describe-dedicated-pending -o json", fixture: "kafka34.golden", wantErrCode: 0},
		{args: "kafka cluster describe lkc-describe-dedicated-pending -o yaml", fixture: "kafka35.golden", wantErrCode: 0},

		{args: "kafka cluster describe lkc-describe-dedicated-with-encryption", fixture: "kafka36.golden", wantErrCode: 0},
		{args: "kafka cluster describe lkc-describe-dedicated-with-encryption -o json", fixture: "kafka37.golden", wantErrCode: 0},
		{args: "kafka cluster describe lkc-describe-dedicated-with-encryption -o yaml", fixture: "kafka38.golden", wantErrCode: 0},

		{args: "kafka acl list --cluster lkc-acls", fixture: "kafka-acls-list.golden", wantErrCode: 0},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBED --topic 'test-topic'", fixture: "kafka-acls-invalid-operation.golden", wantErrCode: 1},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'", fixture: "", wantErrCode: 0},
		{args: "kafka acl delete --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'", fixture: "", wantErrCode: 0},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}
