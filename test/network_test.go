package test

func (s *CLITestSuite) TestNetworkDescribe() {
	tests := []CLITest{
		{args: "network describe n-abcde1", fixture: "network/describe-aws-ready.golden"},
		{args: "network describe n-abcde2", fixture: "network/describe-gcp-ready.golden"},
		{args: "network describe n-abcde3", fixture: "network/describe-azure-ready.golden"},
		{args: "network describe n-abcde4", fixture: "network/describe-aws-provisioning.golden"},
		{args: "network describe n-abcde5", fixture: "network/describe-gcp-provisioning.golden"},
		{args: "network describe n-abcde6", fixture: "network/describe-azure-provisioning.golden"},
		{args: "network describe n-abcde1 --output yaml", fixture: "network/describe-aws-ready-yaml.golden"},
		{args: "network describe n-abcde4 --output yaml", fixture: "network/describe-aws-provisioning-yaml.golden"},
		{args: "network describe", fixture: "network/describe-missing-id.golden", exitCode: 1},
		{args: "network describe n-invalid", fixture: "network/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDelete() {
	tests := []CLITest{
		{args: "network delete n-abcde1 --force", fixture: "network/delete.golden"},
		{args: "network delete n-abcde1", input: "y\n", fixture: "network/delete-prompt.golden"},
		{args: "network delete n-abcde1 n-invalid", fixture: "network/delete-multiple-fail.golden", exitCode: 1},
		{args: "network delete n-abcde1 n-abcde2", input: "n\n", fixture: "network/delete-multiple-refuse.golden"},
		{args: "network delete n-abcde1 n-abcde2", input: "y\n", fixture: "network/delete-multiple-success.golden"},
		{args: "network delete n-dependency --force", fixture: "network/delete-network-with-dependency.golden", exitCode: 1},
		{args: "network delete n-invalid --force", fixture: "network/delete-network-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkUpdate() {
	tests := []CLITest{
		{args: "network update", fixture: "network/update-missing-args.golden", exitCode: 1},
		{args: "network update n-abcde1", fixture: "network/update-missing-flags.golden", exitCode: 1},
		{args: "network update n-abcde1 --name new-network-name", fixture: "network/update.golden"},
		{args: "network update n-invalid --name new-network-name", fixture: "network/update-network-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkList() {
	tests := []CLITest{
		{args: "network list", fixture: "network/list.golden"},
		{args: "network list --output json", fixture: "network/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkCreate() {
	tests := []CLITest{
		{args: "network create aws-tgw --cloud aws --region us-west-2 --connection-types transitgateway --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16 --environment env-00000", fixture: "network/create-tgw.golden"},
		{args: "network create aws-tgw --cloud aws --region us-west-2 --connection-types transitgateway --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16 --environment env-00000 --output json", fixture: "network/create-tgw-json.golden"},
		{args: "network create aws-tgw --cloud aws --region us-west-2 --connection-types transitgateway --zones usw2-az1,usw2-az2,usw2-az4 --environment env-00000", fixture: "network/create-tgw-missing-cidr.golden", exitCode: 1},
		{args: "network create aws-pl --cloud aws --region us-west-2 --connection-types privatelink --zones usw2-az1,usw2-az2,usw2-az3 --dns-resolution private --environment env-00000", fixture: "network/create-pl.golden"},
		{args: "network create aws-peering --cloud aws --region us-west-2 --connection-types peering --zone-info 10.10.0.0/27,10.10.0.32/27,10.10.0.64/27 --reserved-cidr 172.16.10.0/24 --environment env-00000", fixture: "network/create-peering-zone-info-cidr.golden"},
		{args: "network create aws-peering --cloud aws --region us-west-2 --connection-types peering --zone-info usw2-az1=10.10.0.0/27,usw2-az3=10.10.0.32/27,usw2-az4=10.10.0.64/27 --reserved-cidr 172.16.10.0/24 --environment env-00000", fixture: "network/create-peering-zone-info-pairs.golden"},
		{args: "network create aws-peering --cloud aws --region us-west-2 --connection-types peering --zone-info usw2-az1=10.10.0.0/27,usw2-az3=10.10.0.32/27=usw2-az4=10.10.0.64/27 --reserved-cidr 172.16.10.0/24 --environment env-00000", fixture: "network/create-peering-zone-info-invalid.golden", exitCode: 1},
		{args: "network create aws-tgw-peering --cloud aws --region us-west-2 --connection-types transitgateway,peering --zones usw2-az1,usw2-az3,usw2-az4 --zone-info 192.168.1.0/27,192.168.2.0/27,192.168.3.0/27 --environment env-00000", fixture: "network/create-tgw-peering.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetwork_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network describe ""`, login: "cloud", fixture: "network/describe-autocomplete.golden"},
		{args: `__complete network create new-network --connection-types ""`, login: "cloud", fixture: "network/create-autocomplete-connection-types.golden"},
		{args: `__complete network create new-network --dns-resolution ""`, login: "cloud", fixture: "network/create-autocomplete-dns-resolution.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPeeringList() {
	tests := []CLITest{
		{args: "network peering list", fixture: "network/peering/list.golden"},
		{args: "network peering list --output json", fixture: "network/peering/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPeeringDescribe() {
	tests := []CLITest{
		{args: "network peering describe peer-111111", fixture: "network/peering/describe-aws.golden"},
		{args: "network peering describe peer-111111 --output json", fixture: "network/peering/describe-aws-json.golden"},
		{args: "network peering describe peer-111112", fixture: "network/peering/describe-gcp.golden"},
		{args: "network peering describe peer-111113", fixture: "network/peering/describe-azure.golden"},
		{args: "network peering describe", fixture: "network/peering/describe-missing-id.golden", exitCode: 1},
		{args: "network peering describe peer-invalid", fixture: "network/peering/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPeeringUpdate() {
	tests := []CLITest{
		{args: "network peering update", fixture: "network/peering/update-missing-args.golden", exitCode: 1},
		{args: "network peering update peer-111111", fixture: "network/peering/update-missing-flags.golden", exitCode: 1},
		{args: "network peering update peer-111111 --name new-peering-name", fixture: "network/peering/update.golden"},
		{args: "network peering update peer-invalid --name new-peering-name", fixture: "network/peering/update-peering-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPeeringDelete() {
	tests := []CLITest{
		{args: "network peering delete peer-111111 --force", fixture: "network/peering/delete.golden"},
		{args: "network peering delete peer-111111", input: "y\n", fixture: "network/peering/delete-prompt.golden"},
		{args: "network peering delete peer-111111 peer-invalid", fixture: "network/peering/delete-multiple-fail.golden", exitCode: 1},
		{args: "network peering delete peer-111111 peer-111112", input: "n\n", fixture: "network/peering/delete-multiple-refuse.golden"},
		{args: "network peering delete peer-111111 peer-111112", input: "y\n", fixture: "network/peering/delete-multiple-success.golden"},
		{args: "network peering delete peer-invalid --force", fixture: "network/peering/delete-peering-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPeering_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network peering describe ""`, login: "cloud", fixture: "network/peering/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkTransitGatewayAttachmentList() {
	tests := []CLITest{
		{args: "network tgw-attachment list", fixture: "network/transit-gateway-attachment/list.golden"},
		{args: "network transit-gateway-attachment list", fixture: "network/transit-gateway-attachment/list.golden"},
		{args: "network transit-gateway-attachment list --output json", fixture: "network/transit-gateway-attachment/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkTransitGatewayAttachmentDescribe() {
	tests := []CLITest{
		{args: "network tgw-attachment describe tgwa-111111", fixture: "network/transit-gateway-attachment/describe.golden"},
		{args: "network transit-gateway-attachment describe tgwa-111111", fixture: "network/transit-gateway-attachment/describe.golden"},
		{args: "network transit-gateway-attachment describe tgwa-111111 --output json", fixture: "network/transit-gateway-attachment/describe-json.golden"},
		{args: "network transit-gateway-attachment describe", fixture: "network/transit-gateway-attachment/describe-missing-id.golden", exitCode: 1},
		{args: "network transit-gateway-attachment describe tgwa-invalid", fixture: "network/transit-gateway-attachment/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkTransitGatewayAttachment_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network transit-gateway-attachment describe ""`, login: "cloud", fixture: "network/transit-gateway-attachment/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
