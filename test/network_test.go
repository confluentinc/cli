package test

func (s *CLITestSuite) TestNetworkDescribe() {
	tests := []CLITest{
		{args: "network describe n-abcde1", fixture: "network/describe-aws-ready.golden"},
		{args: "network describe n-abcde2", fixture: "network/describe-gcp-ready.golden"},
		{args: "network describe n-abcde3", fixture: "network/describe-azure-ready.golden"},
		{args: "network describe n-abcde4", fixture: "network/describe-aws-provisioning.golden"},
		{args: "network describe n-abcde5", fixture: "network/describe-gcp-provisioning.golden"},
		{args: "network describe n-abcde6", fixture: "network/describe-azure-provisioning.golden"},
		{args: "network describe n-abcde7", fixture: "network/describe-aws-privatelink-ready.golden"},
		{args: "network describe n-abcde8", fixture: "network/describe-gcp-privatelink-ready.golden"},
		{args: "network describe n-abcde9", fixture: "network/describe-azure-privatelink-ready.golden"},
		{args: "network describe n-abcde10", fixture: "network/describe-aws-privatelink-provisioning.golden"},
		{args: "network describe n-abcde11", fixture: "network/describe-gcp-privatelink-provisioning.golden"},
		{args: "network describe n-abcde12", fixture: "network/describe-azure-privatelink-provisioning.golden"},
		{args: "network describe n-abcde1 --output yaml", fixture: "network/describe-aws-ready-yaml.golden"},
		{args: "network describe n-abcde4 --output yaml", fixture: "network/describe-aws-provisioning-yaml.golden"},
		{args: "network describe n-abcde7 --output json", fixture: "network/describe-aws-privatelink-ready-json.golden"},
		{args: "network describe n-abcde8 --output json", fixture: "network/describe-gcp-privatelink-ready-json.golden"},
		{args: "network describe n-abcde9 --output json", fixture: "network/describe-azure-privatelink-ready-json.golden"},
		{args: "network describe n-abcde10 --output json", fixture: "network/describe-aws-privatelink-provisioning-json.golden"},
		{args: "network describe n-abcde11 --output json", fixture: "network/describe-gcp-privatelink-provisioning-json.golden"},
		{args: "network describe n-abcde12 --output json", fixture: "network/describe-azure-privatelink-provisioning-json.golden"},
		{args: "network describe n-abcde13", fixture: "network/describe-gateway.golden"},
		{args: "network describe n-abcde13 --output json", fixture: "network/describe-gateway-json.golden"},
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
		{args: "network list --name prod-gcp-us-central1,prod-aws-us-east1 --cloud aws", fixture: "network/list-name-cloud.golden"},
		{args: "network list --region eastus2 --cidr 10.0.0.0/16", fixture: "network/list-region-cidr.golden"},
		{args: "network list --phase ready --connection-types transitgateway,peering", fixture: "network/list-phase-connection.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkCreate() {
	tests := []CLITest{
		{args: "network create --cloud aws --region us-west-2 --connection-types transitgateway --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16 --environment env-00000", fixture: "network/create-no-name.golden"},
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
		{args: `__complete network update ""`, login: "cloud", fixture: "network/update-autocomplete.golden"},
		{args: `__complete network delete ""`, login: "cloud", fixture: "network/delete-autocomplete.golden"},
		{args: `__complete network create new-network --connection-types ""`, login: "cloud", fixture: "network/create-autocomplete-connection-types.golden"},
		{args: `__complete network create new-network --dns-resolution ""`, login: "cloud", fixture: "network/create-autocomplete-dns-resolution.golden"},
		{args: `__complete network create new-network --region ""`, login: "cloud", fixture: "network/create-autocomplete-region.golden"},
		{args: `__complete network create new-network --cloud aws --region ""`, login: "cloud", fixture: "network/create-autocomplete-region-with-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkGatewayDescribe() {
	tests := []CLITest{
		{args: "network gateway describe gw-12345", fixture: "network/gateway/describe-aws.golden"},
		{args: "network gateway describe gw-67890", fixture: "network/gateway/describe-azure.golden"},
		{args: "network gateway describe gw-12345 --output json", fixture: "network/gateway/describe-aws-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkGatewayList() {
	tests := []CLITest{
		{args: "network gateway list", fixture: "network/gateway/list.golden"},
		{args: "network gateway list --output json", fixture: "network/gateway/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkGateway_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network gateway describe ""`, login: "cloud", fixture: "network/gateway/describe-autocomplete.golden"},
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
		{args: "network peering list --network n-abcde1 --name aws-peering", fixture: "network/peering/list-network-name.golden"},
		{args: "network peering list --phase ready --name gcp-peering,azure-peering", fixture: "network/peering/list-phase-name.golden"},
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

func (s *CLITestSuite) TestNetworkPeeringCreate() {
	tests := []CLITest{
		{args: "network peering create --network n-abcde1 --cloud aws --cloud-account 012345678901 --virtual-network vpc-abcdef0123456789a --aws-routes 172.31.0.0/16,10.108.16.0/21", fixture: "network/peering/create-no-name.golden"},
		{args: "network peering create aws-peering --network n-abcde1 --cloud aws --cloud-account 012345678901 --virtual-network vpc-abcdef0123456789a --aws-routes 172.31.0.0/16,10.108.16.0/21", fixture: "network/peering/create-aws.golden"},
		{args: "network peering create aws-peering --network n-abcde1 --cloud aws --cloud-account 012345678901 --virtual-network vpc-abcdef0123456789a --aws-routes 172.31.0.0/16,10.108.16.0/21 --customer-region us-west-2", fixture: "network/peering/create-aws-customer-region.golden"},
		{args: "network peering create aws-peering --network n-abcde1 --cloud aws --cloud-account 012345678901 --aws-routes 172.31.0.0/16,10.108.16.0/21", fixture: "network/peering/create-aws-missing-flags.golden", exitCode: 1},
		{args: "network peering create gcp-peering --network n-abcde1 --cloud gcp --cloud-account temp-123456 --virtual-network customer-test-vpc-network", fixture: "network/peering/create-gcp.golden"},
		{args: "network peering create gcp-peering --network n-abcde1 --cloud gcp --cloud-account temp-123456 --virtual-network customer-test-vpc-network --gcp-routes", fixture: "network/peering/create-gcp-import-custom-routes.golden"},
		{args: "network peering create gcp-peering --network n-abcde1 --cloud gcp --virtual-network customer-test-vpc-network --gcp-routes", fixture: "network/peering/create-gcp-missing-flags.golden", exitCode: 1},
		{args: "network peering create azure-peering --network n-abcde1 --cloud azure --cloud-account 1111tttt-1111-1111-1111-111111tttttt --virtual-network /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet", fixture: "network/peering/create-azure.golden"},
		{args: "network peering create azure-peering --network n-abcde1 --cloud azure --cloud-account 1111tttt-1111-1111-1111-111111tttttt --virtual-network /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet --customer-region centralus", fixture: "network/peering/create-azure-customer-region.golden"},
		{args: "network peering create azure-peering --network n-abcde1 --cloud azure --cloud-account 1111tttt-1111-1111-1111-111111tttttt", fixture: "network/peering/create-azure-missing-flags.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPeering_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network peering describe ""`, login: "cloud", fixture: "network/peering/describe-autocomplete.golden"},
		{args: `__complete network peering update ""`, login: "cloud", fixture: "network/peering/update-autocomplete.golden"},
		{args: `__complete network peering delete ""`, login: "cloud", fixture: "network/peering/delete-autocomplete.golden"},
		{args: `__complete network peering create aws-peering --network ""`, login: "cloud", fixture: "network/peering/create-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkTransitGatewayAttachmentList() {
	tests := []CLITest{
		{args: "network tgwa list", fixture: "network/transit-gateway-attachment/list.golden"},
		{args: "network transit-gateway-attachment list", fixture: "network/transit-gateway-attachment/list.golden"},
		{args: "network transit-gateway-attachment list --output json", fixture: "network/transit-gateway-attachment/list-json.golden"},
		{args: "network transit-gateway-attachment list --network n-abcde1 --name aws-tgwa1,aws-tgwa2", fixture: "network/transit-gateway-attachment/list-network-name.golden"},
		{args: "network transit-gateway-attachment list --phase ready --name aws-tgwa3", fixture: "network/transit-gateway-attachment/list-phase-name.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkTransitGatewayAttachmentDescribe() {
	tests := []CLITest{
		{args: "network tgwa describe tgwa-111111", fixture: "network/transit-gateway-attachment/describe.golden"},
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

func (s *CLITestSuite) TestNetworkTransitGatewayAttachmentUpdate() {
	tests := []CLITest{
		{args: "network transit-gateway-attachment update", fixture: "network/transit-gateway-attachment/update-missing-args.golden", exitCode: 1},
		{args: "network transit-gateway-attachment update tgwa-111111", fixture: "network/transit-gateway-attachment/update-missing-flags.golden", exitCode: 1},
		{args: "network tgwa update tgwa-111111 --name new-name", fixture: "network/transit-gateway-attachment/update.golden"},
		{args: "network transit-gateway-attachment update tgwa-111111 --name new-name", fixture: "network/transit-gateway-attachment/update.golden"},
		{args: "network transit-gateway-attachment update tgwa-invalid --name new-name", fixture: "network/transit-gateway-attachment/update-tgwa-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkTransitGatewayAttachmentDelete() {
	tests := []CLITest{
		{args: "network transit-gateway-attachment delete tgwa-111111 --force", fixture: "network/transit-gateway-attachment/delete.golden"},
		{args: "network transit-gateway-attachment delete tgwa-111111", input: "y\n", fixture: "network/transit-gateway-attachment/delete-prompt.golden"},
		{args: "network transit-gateway-attachment delete tgwa-111111 tgwa-222222", input: "n\n", fixture: "network/transit-gateway-attachment/delete-multiple-refuse.golden"},
		{args: "network transit-gateway-attachment delete tgwa-111111 tgwa-222222", input: "y\n", fixture: "network/transit-gateway-attachment/delete-multiple-success.golden"},
		{args: "network transit-gateway-attachment delete tgwa-111111 tgwa-invalid", fixture: "network/transit-gateway-attachment/delete-multiple-fail.golden", exitCode: 1},
		{args: "network transit-gateway-attachment delete tgwa-invalid --force", fixture: "network/transit-gateway-attachment/delete-tgwa-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkTransitGatewayAttachmentCreate() {
	tests := []CLITest{
		{args: "network tgwa create aws-tgwa --network n-abcde1 --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10", fixture: "network/transit-gateway-attachment/create.golden"},
		{args: "network transit-gateway-attachment create --network n-abcde1 --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10", fixture: "network/transit-gateway-attachment/create-no-name.golden"},
		{args: "network transit-gateway-attachment create aws-tgwa --network n-abcde1 --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10", fixture: "network/transit-gateway-attachment/create.golden"},
		{args: "network transit-gateway-attachment create aws-tgwa --network n-azure --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10", fixture: "network/transit-gateway-attachment/create-duplicate.golden", exitCode: 1},
		{args: "network transit-gateway-attachment create aws-tgwa --network n-duplicate --aws-ram-share-arn arn:aws:ram:us-west-2:123456789012:resource-share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx --aws-transit-gateway tgw-xxxxxxxxxxxxxxxxx --routes 10.0.0.0/16,100.64.0.0/10", fixture: "network/transit-gateway-attachment/create-wrong-network-cloud.golden", exitCode: 1},
		{args: "network transit-gateway-attachment create aws-tgwa --network n-abcde1 --routes 10.0.0.0/16,100.64.0.0/10", fixture: "network/transit-gateway-attachment/create-missing-flags.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkTransitGatewayAttachment_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network transit-gateway-attachment describe ""`, login: "cloud", fixture: "network/transit-gateway-attachment/describe-autocomplete.golden"},
		{args: `__complete network transit-gateway-attachment update ""`, login: "cloud", fixture: "network/transit-gateway-attachment/update-autocomplete.golden"},
		{args: `__complete network transit-gateway-attachment delete ""`, login: "cloud", fixture: "network/transit-gateway-attachment/delete-autocomplete.golden"},
		{args: `__complete network transit-gateway-attachment create tgwa --network ""`, login: "cloud", fixture: "network/transit-gateway-attachment/create-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAccessList() {
	tests := []CLITest{
		{args: "network pl access list", fixture: "network/private-link/access/list.golden"},
		{args: "network private-link access list", fixture: "network/private-link/access/list.golden"},
		{args: "network private-link access list --output json", fixture: "network/private-link/access/list-json.golden"},
		{args: "network private-link access list --network n-abcde1 --name aws-pla", fixture: "network/private-link/access/list-network-name.golden"},
		{args: "network private-link access list --phase ready --name gcp-pla,azure-pla", fixture: "network/private-link/access/list-phase-name.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAccessDescribe() {
	tests := []CLITest{
		{args: "network pl access describe pla-111111", fixture: "network/private-link/access/describe-aws.golden"},
		{args: "network private-link access describe pla-111111", fixture: "network/private-link/access/describe-aws.golden"},
		{args: "network private-link access describe pla-111111 --output json", fixture: "network/private-link/access/describe-aws-json.golden"},
		{args: "network private-link access describe pla-111112", fixture: "network/private-link/access/describe-gcp.golden"},
		{args: "network private-link access describe pla-111113", fixture: "network/private-link/access/describe-azure.golden"},
		{args: "network private-link access describe", fixture: "network/private-link/access/describe-missing-id.golden", exitCode: 1},
		{args: "network private-link access describe pla-invalid", fixture: "network/private-link/access/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAccessUpdate() {
	tests := []CLITest{
		{args: "network private-link access update", fixture: "network/private-link/access/update-missing-args.golden", exitCode: 1},
		{args: "network private-link access update pla-111111", fixture: "network/private-link/access/update-missing-flags.golden", exitCode: 1},
		{args: "network pl access update pla-111111 --name new-name", fixture: "network/private-link/access/update.golden"},
		{args: "network private-link access update pla-111111 --name new-name", fixture: "network/private-link/access/update.golden"},
		{args: "network private-link access update pla-invalid --name new-name", fixture: "network/private-link/access/update-pla-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAccessDelete() {
	tests := []CLITest{
		{args: "network private-link access delete pla-111111 --force", fixture: "network/private-link/access/delete.golden"},
		{args: "network private-link access delete pla-111111", input: "y\n", fixture: "network/private-link/access/delete-prompt.golden"},
		{args: "network private-link access delete pla-111111 pla-222222", input: "n\n", fixture: "network/private-link/access/delete-multiple-refuse.golden"},
		{args: "network private-link access delete pla-111111 pla-222222", input: "y\n", fixture: "network/private-link/access/delete-multiple-success.golden"},
		{args: "network private-link access delete pla-111111 pla-invalid", fixture: "network/private-link/access/delete-multiple-fail.golden", exitCode: 1},
		{args: "network private-link access delete pla-invalid --force", fixture: "network/private-link/access/delete-pla-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAccessCreate() {
	tests := []CLITest{
		{args: "network private-link access create pla --network n-abcde1", fixture: "network/private-link/access/create-missing-flags.golden", exitCode: 1},
		{args: "network private-link access create --network n-abcde1 --cloud aws --cloud-account 012345678901", fixture: "network/private-link/access/create-no-name.golden"},
		{args: "network private-link access create aws-pla --network n-abcde1 --cloud aws --cloud-account 012345678901", fixture: "network/private-link/access/create-aws.golden"},
		{args: "network private-link access create gcp-pla --network n-abcde1 --cloud gcp --cloud-account temp-123456", fixture: "network/private-link/access/create-gcp.golden"},
		{args: "network private-link access create azure-pla --network n-abcde1 --cloud azure --cloud-account 1234abcd-12ab-34cd-1234-123456abcdef", fixture: "network/private-link/access/create-azure.golden"},
		{args: "network private-link access create aws-pla --network n-azure --cloud aws --cloud-account 012345678901", fixture: "network/private-link/access/create-duplicate.golden", exitCode: 1},
		{args: "network private-link access create aws-pla --network n-duplicate --cloud aws --cloud-account 012345678901", fixture: "network/private-link/access/create-wrong-network-cloud.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAccess_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network private-link access describe ""`, login: "cloud", fixture: "network/private-link/access/describe-autocomplete.golden"},
		{args: `__complete network private-link access update ""`, login: "cloud", fixture: "network/private-link/access/update-autocomplete.golden"},
		{args: `__complete network private-link access delete ""`, login: "cloud", fixture: "network/private-link/access/delete-autocomplete.golden"},
		{args: `__complete network private-link access create pla --network ""`, login: "cloud", fixture: "network/private-link/access/create-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentList() {
	tests := []CLITest{
		{args: "network pl attachment list", fixture: "network/private-link/attachment/list.golden"},
		{args: "network private-link attachment list", fixture: "network/private-link/attachment/list.golden"},
		{args: "network private-link attachment list --output json", fixture: "network/private-link/attachment/list-json.golden"},
		{args: "network private-link attachment list --name aws-platt-1,aws-platt-2 --cloud aws", fixture: "network/private-link/attachment/list-name-cloud.golden"},
		{args: "network private-link attachment list --region us-west-2 --phase provisioning ", fixture: "network/private-link/attachment/list-region-phase.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentDescribe() {
	tests := []CLITest{
		{args: "network pl attachment describe platt-111111", fixture: "network/private-link/attachment/describe-aws.golden"},
		{args: "network private-link attachment describe platt-111111", fixture: "network/private-link/attachment/describe-aws.golden"},
		{args: "network private-link attachment describe platt-111112", fixture: "network/private-link/attachment/describe-aws-provisioning.golden"},
		{args: "network private-link attachment describe platt-111111 --output json", fixture: "network/private-link/attachment/describe-aws-json.golden"},
		{args: "network private-link attachment describe", fixture: "network/private-link/attachment/describe-missing-id.golden", exitCode: 1},
		{args: "network private-link attachment describe platt-invalid", fixture: "network/private-link/attachment/describe-invalid.golden", exitCode: 1},
		{args: "network private-link attachment describe platt-azure", fixture: "network/private-link/attachment/describe-azure.golden"},
		{args: "network private-link attachment describe platt-azure-2", fixture: "network/private-link/attachment/describe-azure-provisioning.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentUpdate() {
	tests := []CLITest{
		{args: "network private-link attachment update", fixture: "network/private-link/attachment/update-missing-args.golden", exitCode: 1},
		{args: "network private-link attachment update platt-111111", fixture: "network/private-link/attachment/update-missing-flags.golden", exitCode: 1},
		{args: "network pl attachment update platt-111111 --name new-name", fixture: "network/private-link/attachment/update.golden"},
		{args: "network private-link attachment update platt-111111 --name new-name", fixture: "network/private-link/attachment/update.golden"},
		{args: "network private-link attachment update platt-invalid --name new-name", fixture: "network/private-link/attachment/update-platt-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentDelete() {
	tests := []CLITest{
		{args: "network private-link attachment delete platt-111111 --force", fixture: "network/private-link/attachment/delete.golden"},
		{args: "network private-link attachment delete platt-111111", input: "y\n", fixture: "network/private-link/attachment/delete-prompt.golden"},
		{args: "network private-link attachment delete platt-111111 platt-222222", input: "n\n", fixture: "network/private-link/attachment/delete-multiple-refuse.golden"},
		{args: "network private-link attachment delete platt-111111 platt-222222", input: "y\n", fixture: "network/private-link/attachment/delete-multiple-success.golden"},
		{args: "network private-link attachment delete platt-111111 platt-invalid", fixture: "network/private-link/attachment/delete-multiple-fail.golden", exitCode: 1},
		{args: "network private-link attachment delete platt-invalid --force", fixture: "network/private-link/attachment/delete-platt-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentCreate() {
	tests := []CLITest{
		{args: "network private-link attachment create platt", fixture: "network/private-link/attachment/create-missing-flags.golden", exitCode: 1},
		{args: "network private-link attachment create --cloud aws --region us-west-2", fixture: "network/private-link/attachment/create-no-name.golden", exitCode: 1},
		{args: "network private-link attachment create aws-platt --cloud aws --region us-west-2", fixture: "network/private-link/attachment/create-aws.golden"},
		{args: "network private-link attachment create gcp-platt --cloud gcp --region us-central1", fixture: "network/private-link/attachment/create-gcp-fail.golden", exitCode: 1},
		{args: "network private-link attachment create azure-platt --cloud azure --region eastus2", fixture: "network/private-link/attachment/create-azure.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachment_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network private-link attachment describe ""`, login: "cloud", fixture: "network/private-link/attachment/describe-autocomplete.golden"},
		{args: `__complete network private-link attachment update ""`, login: "cloud", fixture: "network/private-link/attachment/update-autocomplete.golden"},
		{args: `__complete network private-link attachment delete ""`, login: "cloud", fixture: "network/private-link/attachment/delete-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentConnectionList() {
	tests := []CLITest{
		{args: "network pl attachment connection list --attachment platt-111111", fixture: "network/private-link/attachment/connection/list.golden"},
		{args: "network private-link attachment connection list", fixture: "network/private-link/attachment/connection/list-missing-flags.golden", exitCode: 1},
		{args: "network private-link attachment connection list --attachment platt-111111", fixture: "network/private-link/attachment/connection/list.golden"},
		{args: "network private-link attachment connection list --attachment platt-invalid", fixture: "network/private-link/attachment/connection/list-invalid.golden"},
		{args: "network private-link attachment connection list --attachment platt-111111 --output json", fixture: "network/private-link/attachment/connection/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentConnectionDescribe() {
	tests := []CLITest{
		{args: "network pl attachment connection describe plattc-111111", fixture: "network/private-link/attachment/connection/describe-aws.golden"},
		{args: "network private-link attachment connection describe plattc-111111", fixture: "network/private-link/attachment/connection/describe-aws.golden"},
		{args: "network private-link attachment connection describe plattc-111112", fixture: "network/private-link/attachment/connection/describe-aws-provisioning.golden"},
		{args: "network private-link attachment connection describe plattc-111111 --output json", fixture: "network/private-link/attachment/connection/describe-aws-json.golden"},
		{args: "network private-link attachment connection describe", fixture: "network/private-link/attachment/connection/describe-missing-id.golden", exitCode: 1},
		{args: "network private-link attachment connection describe plattc-invalid", fixture: "network/private-link/attachment/connection/describe-invalid.golden", exitCode: 1},
		{args: "network private-link attachment connection describe plattc-azure", fixture: "network/private-link/attachment/connection/describe-azure.golden"},
		{args: "network private-link attachment connection describe plattc-azure-2", fixture: "network/private-link/attachment/connection/describe-azure-provisioning.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentConnectionUpdate() {
	tests := []CLITest{
		{args: "network private-link attachment connection update", fixture: "network/private-link/attachment/connection/update-missing-args.golden", exitCode: 1},
		{args: "network private-link attachment connection update plattc-111111", fixture: "network/private-link/attachment/connection/update-missing-flags.golden", exitCode: 1},
		{args: "network pl attachment connection update plattc-111111 --name new-name", fixture: "network/private-link/attachment/connection/update.golden"},
		{args: "network private-link attachment connection update plattc-111111 --name new-name", fixture: "network/private-link/attachment/connection/update.golden"},
		{args: "network private-link attachment connection update plattc-invalid --name new-name", fixture: "network/private-link/attachment/connection/update-plattc-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentConnectionDelete() {
	tests := []CLITest{
		{args: "network private-link attachment connection delete plattc-111111 --force", fixture: "network/private-link/attachment/connection/delete.golden"},
		{args: "network private-link attachment connection delete plattc-111111", input: "y\n", fixture: "network/private-link/attachment/connection/delete-prompt.golden"},
		{args: "network private-link attachment connection delete plattc-111111 plattc-222222", input: "n\n", fixture: "network/private-link/attachment/connection/delete-multiple-refuse.golden"},
		{args: "network private-link attachment connection delete plattc-111111 plattc-222222", input: "y\n", fixture: "network/private-link/attachment/connection/delete-multiple-success.golden"},
		{args: "network private-link attachment connection delete plattc-111111 plattc-invalid", fixture: "network/private-link/attachment/connection/delete-multiple-fail.golden", exitCode: 1},
		{args: "network private-link attachment connection delete plattc-invalid --force", fixture: "network/private-link/attachment/connection/delete-plattc-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentConnectionCreate() {
	tests := []CLITest{
		{args: "network private-link attachment connection create plattc", fixture: "network/private-link/attachment/connection/create-missing-flags.golden", exitCode: 1},
		{args: "network private-link attachment connection create plattc-wrong-cloud --cloud invalid --endpoint vpce-1234567890abcdef0 --attachment platt-123456", fixture: "network/private-link/attachment/connection/create-invalid-cloud.golden", exitCode: 1},
		{args: "network private-link attachment connection create --cloud aws --endpoint vpce-1234567890abcdef0 --attachment platt-123456", fixture: "network/private-link/attachment/connection/create-no-name.golden", exitCode: 1},
		{args: "network private-link attachment connection create aws-plattc --cloud aws --endpoint vpce-1234567890abcdef0 --attachment platt-123456", fixture: "network/private-link/attachment/connection/create-aws.golden"},
		{args: "network private-link attachment connection create aws-plattc-wrong-endpoint --cloud aws --endpoint vpce-invalid --attachment platt-123456", fixture: "network/private-link/attachment/connection/create-aws-invalid-endpoint.golden", exitCode: 1},
		{args: "network private-link attachment connection create aws-plattc-invalid-platt --cloud aws --endpoint vpce-1234567890abcdef0 --attachment platt-invalid", fixture: "network/private-link/attachment/connection/create-aws-platt-not-found.golden", exitCode: 1},
		{args: "network private-link attachment connection create gcp-plattc-wrong-platt-cloud --cloud gcp --endpoint vpce-1234567890abcdef0 --attachment platt-aws123", fixture: "network/private-link/attachment/connection/create-wrong-platt-cloud.golden", exitCode: 1},
		{args: "network private-link attachment connection create azure-plattc --cloud azure --endpoint azure-pl-endpoint --attachment platt-azure", fixture: "network/private-link/attachment/connection/create-azure.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkPrivateLinkAttachmentConnection_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network private-link attachment connection list --attachment ""`, login: "cloud", fixture: "network/private-link/attachment/connection/list-autocomplete.golden"},
		{args: `__complete network private-link attachment connection create platt-connection --attachment ""`, login: "cloud", fixture: "network/private-link/attachment/connection/create-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkServiceDescribe() {
	tests := []CLITest{
		{args: "network link service describe nls-123456", fixture: "network/link/service/describe.golden"},
		{args: "network link service describe nls-123456 --output json", fixture: "network/link/service/describe-json.golden"},
		{args: "network link service describe", fixture: "network/link/service/describe-missing-id.golden", exitCode: 1},
		{args: "network link service describe nls-invalid", fixture: "network/link/service/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkServiceList() {
	tests := []CLITest{
		{args: "network link service list", fixture: "network/link/service/list.golden"},
		{args: "network link service list --output json", fixture: "network/link/service/list-json.golden"},
		{args: "network link service list --network n-abcde1 --name my-network-link-service-1", fixture: "network/link/service/list-network-name.golden"},
		{args: "network link service list --phase ready --name my-network-link-service-2,my-network-link-service-3", fixture: "network/link/service/list-phase-name.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkServiceDelete() {
	tests := []CLITest{
		{args: "network link service delete nls-111111 --force", fixture: "network/link/service/delete.golden"},
		{args: "network link service delete nls-111111", input: "y\n", fixture: "network/link/service/delete-prompt.golden"},
		{args: "network link service delete nls-111111 nls-222222", input: "n\n", fixture: "network/link/service/delete-multiple-refuse.golden"},
		{args: "network link service delete nls-111111 nls-222222", input: "y\n", fixture: "network/link/service//delete-multiple-success.golden"},
		{args: "network link service delete nls-111111 nls-invalid", fixture: "network/link/service/delete-multiple-fail.golden", exitCode: 1},
		{args: "network link service delete nls-invalid --force", fixture: "network/link/service/delete-nls-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkServiceCreate() {
	tests := []CLITest{
		{args: "network link service create", fixture: "network/link/service/create-missing-network.golden", exitCode: 1},
		{args: "network link service create --network n-123456", fixture: "network/link/service/create-missing-flag.golden", exitCode: 1},
		{args: "network link service create --network n-123456 --description 'example network link service' --accepted-environments env-11111,env-22222", fixture: "network/link/service/create-no-name.golden"},
		{args: "network link service create my-network-link-service --network n-123456 --description 'example network link service' --accepted-environments env-11111,env-22222", fixture: "network/link/service/create-accepted-environments.golden"},
		{args: "network link service create my-network-link-service --network n-123456 --description 'example network link service' --accepted-networks n-111111,n-222222", fixture: "network/link/service/create-accepted-networks.golden"},
		{args: "network link service create my-network-link-service --network n-123456 --description 'example network link service' --accepted-networks n-111111,n-222222 --accepted-environments env-11111,env-22222", fixture: "network/link/service/create.golden"},
		{args: "network link service create nls-duplicate --network n-123455 --accepted-networks n-111111", fixture: "network/link/service/create-duplicate.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkServiceUpdate() {
	tests := []CLITest{
		{args: "network link service update", fixture: "network/link/service/update-missing-args.golden", exitCode: 1},
		{args: "network link service update nls-111111", fixture: "network/link/service/update-missing-flags.golden", exitCode: 1},
		{args: "network link service update nls-111111 --name my-new-network-link-service --description 'example new network link service'", fixture: "network/link/service/update.golden"},
		{args: "network link service update nls-111111 --accepted-environments env-22222 --accepted-networks n-111111", fixture: "network/link/service/update-accept-policy.golden"},
		{args: "network link service update nls-111111 --accepted-environments env-11111,env-22222 --accepted-networks n-111111,n-222222", fixture: "network/link/service/update-accept-policy-multiple.golden"},
		{args: "network link service update nls-invalid --name 'my-network-link-service'", fixture: "network/link/service/update-nls-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkService_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network link service describe ""`, login: "cloud", fixture: "network/link/service/describe-autocomplete.golden"},
		{args: `__complete network link service delete ""`, login: "cloud", fixture: "network/link/service/delete-autocomplete.golden"},
		{args: `__complete network link service create my-network-link-service --network ""`, login: "cloud", fixture: "network/link/service/create-autocomplete.golden"},
		{args: `__complete network link service update ""`, login: "cloud", fixture: "network/link/service/update-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkEndpointDescribe() {
	tests := []CLITest{
		{args: "network link endpoint describe nle-123456", fixture: "network/link/endpoint/describe.golden"},
		{args: "network link endpoint describe nle-123456 --output json", fixture: "network/link/endpoint/describe-json.golden"},
		{args: "network link endpoint describe", fixture: "network/link/endpoint/describe-missing-id.golden", exitCode: 1},
		{args: "network link endpoint describe nle-invalid", fixture: "network/link/endpoint/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkEndpointList() {
	tests := []CLITest{
		{args: "network link endpoint list", fixture: "network/link/endpoint/list.golden"},
		{args: "network link endpoint list --output json", fixture: "network/link/endpoint/list-json.golden"},
		{args: "network link endpoint list --network n-abcde1 --name my-network-link-endpoint-1", fixture: "network/link/endpoint/list-network-name.golden"},
		{args: "network link endpoint list --phase ready --name my-network-link-endpoint-2,my-network-link-endpoint-3", fixture: "network/link/endpoint/list-phase-name.golden"},
		{args: "network link endpoint list --network-link-service nls-123456", fixture: "network/link/endpoint/list-service.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkEndpointDelete() {
	tests := []CLITest{
		{args: "network link endpoint delete nle-111111 --force", fixture: "network/link/endpoint/delete.golden"},
		{args: "network link endpoint delete nle-111111", input: "y\n", fixture: "network/link/endpoint/delete-prompt.golden"},
		{args: "network link endpoint delete nle-111111 nle-222222", input: "n\n", fixture: "network/link/endpoint/delete-multiple-refuse.golden"},
		{args: "network link endpoint delete nle-111111 nle-222222", input: "y\n", fixture: "network/link/endpoint/delete-multiple-success.golden"},
		{args: "network link endpoint delete nle-111111 nle-invalid", fixture: "network/link/endpoint/delete-multiple-fail.golden", exitCode: 1},
		{args: "network link endpoint delete nle-invalid --force", fixture: "network/link/endpoint/delete-nle-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkEndpointCreate() {
	tests := []CLITest{
		{args: "network link endpoint create", fixture: "network/link/endpoint/create-missing-flags.golden", exitCode: 1},
		{args: "network link endpoint create --network n-123456 --description 'example network link endpoint' --network-link-service nls-abcde1", fixture: "network/link/endpoint/create-no-name.golden"},
		{args: "network link endpoint create my-network-link-endpoint --network n-123456 --description 'example network link endpoint' --network-link-service nls-abcde1", fixture: "network/link/endpoint/create.golden"},
		{args: "network link endpoint create nle-duplicate --network n-123455 --network-link-service nls-abcde1", fixture: "network/link/endpoint/create-duplicate.golden", exitCode: 1},
		{args: "network link endpoint create nle-same-id --network n-123455 --network-link-service nls-abcde1", fixture: "network/link/endpoint/create-same-id.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkEndpointUpdate() {
	tests := []CLITest{
		{args: "network link endpoint update", fixture: "network/link/endpoint/update-missing-args.golden", exitCode: 1},
		{args: "network link endpoint update nle-111111", fixture: "network/link/endpoint/update-missing-flags.golden", exitCode: 1},
		{args: "network link endpoint update nle-111111 --name my-new-network-link-endpoint", fixture: "network/link/endpoint/update.golden"},
		{args: "network link endpoint update nle-111111 --name my-new-network-link-endpoint --description 'example new network link endpoint'", fixture: "network/link/endpoint/update-name-description.golden"},
		{args: "network link endpoint update nle-invalid --name 'my-network-link-endpoint'", fixture: "network/link/endpoint/update-nle-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkEndpoint_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network link endpoint describe ""`, login: "cloud", fixture: "network/link/endpoint/describe-autocomplete.golden"},
		{args: `__complete network link endpoint delete ""`, login: "cloud", fixture: "network/link/endpoint/delete-autocomplete.golden"},
		{args: `__complete network link endpoint create my-network-link-endpoint --network ""`, login: "cloud", fixture: "network/link/endpoint/create-autocomplete.golden"},
		{args: `__complete network link endpoint update ""`, login: "cloud", fixture: "network/link/endpoint/update-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkRegionList() {
	tests := []CLITest{
		{args: "network region list", fixture: "network/region/list.golden"},
		{args: "network region list --output json", fixture: "network/region/list-json.golden"},
		{args: "network region list --cloud aws", fixture: "network/region/list-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkIpAddressList() {
	tests := []CLITest{
		{args: "network ip-address list", fixture: "network/ip-address/list.golden"},
		{args: "network ip-address list --output json", fixture: "network/ip-address/list-json.golden"},
		{args: "network ip-address list --cloud aws --region us-east-1", fixture: "network/ip-address/list-cloud-region.golden"},
		{args: "network ip-address list --services kafka --address-type egress", fixture: "network/ip-address/list-services-address.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkServiceAssociationDescribe() {
	tests := []CLITest{
		{args: "network link service association describe nle-123456 --network-link-service nls-123456", fixture: "network/link/service/association/describe.golden"},
		{args: "network link service association describe nle-123456 --network-link-service nls-123456 --output json", fixture: "network/link/service/association/describe-json.golden"},
		{args: "network link service association describe nle-123456", fixture: "network/link/service/association/describe-missing-flag.golden", exitCode: 1},
		{args: "network link service association describe", fixture: "network/link/service/association/describe-missing-id.golden", exitCode: 1},
		{args: "network link service association describe nle-invalid --network-link-service nls-123456", fixture: "network/link/service/association/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkServiceAssociationList() {
	tests := []CLITest{
		{args: "network link service association list --network-link-service nls-123456", fixture: "network/link/service/association/list.golden"},
		{args: "network link service association list --network-link-service nls-123456 --output json", fixture: "network/link/service/association/list-json.golden"},
		{args: "network link service association list --network-link-service nls-invalid", fixture: "network/link/service/association/list-nls-invalid.golden", exitCode: 1},
		{args: "network link service association list --network-link-service nls-no-endpoints", fixture: "network/link/service/association/list-no-endpoints.golden", exitCode: 1},
		{args: "network link service association list ", fixture: "network/link/service/association/list-missing-flag.golden", exitCode: 1},
		{args: "network link service association list --network-link-service nls-123456 --phase pending-accept", fixture: "network/link/service/association/list-phase.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkNetworkLinkServiceAssociation_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network link service association describe ""`, login: "cloud", fixture: "network/link/service/association/describe-autocomplete.golden"},
		{args: `__complete network link service association list --network-link-service ""`, login: "cloud", fixture: "network/link/service/association/list-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsForwarderDescribe() {
	tests := []CLITest{
		{args: "network dns forwarder describe dnsf-abcde1", fixture: "network/dns/forwarder/describe.golden"},
		{args: "network dns forwarder describe dnsf-abcde1 --output json", fixture: "network/dns/forwarder/describe-json.golden"},
		{args: "network dns forwarder describe", fixture: "network/dns/forwarder/describe-missing-id.golden", exitCode: 1},
		{args: "network dns forwarder describe dnsf-invalid", fixture: "network/dns/forwarder/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsForwarderList() {
	tests := []CLITest{
		{args: "network dns forwarder list", fixture: "network/dns/forwarder/list.golden"},
		{args: "network dns forwarder list --output json", fixture: "network/dns/forwarder/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsForwarderDelete() {
	tests := []CLITest{
		{args: "network dns forwarder delete dnsf-111111", input: "y\n", fixture: "network/dns/forwarder/delete-prompt.golden"},
		{args: "network dns forwarder delete dnsf-111111 dnsf-222222", input: "n\n", fixture: "network/dns/forwarder/delete-multiple-refuse.golden"},
		{args: "network dns forwarder delete dnsf-111111 dnsf-222222", input: "y\n", fixture: "network/dns/forwarder/delete-multiple-success.golden"},
		{args: "network dns forwarder delete dnsf-111111 dnsf-invalid", fixture: "network/dns/forwarder/delete-multiple-fail.golden", exitCode: 1},
		{args: "network dns forwarder delete dnsf-invalid --force", fixture: "network/dns/forwarder/delete-dnsf-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsForwarderUpdate() {
	tests := []CLITest{
		{args: "network dns forwarder update", fixture: "network/dns/forwarder/update-missing-args.golden", exitCode: 1},
		{args: "network dns forwarder update dnsf-111111", fixture: "network/dns/forwarder/update-missing-flags.golden", exitCode: 1},
		{args: "network dns forwarder update dnsf-111111 --name my-new-dns-forwarder --domains ghi.com,jkl.com,xyz.com", fixture: "network/dns/forwarder/update.golden"},
		{args: "network dns forwarder update dnsf-111111 --dns-server-ips 10.208.0.0,10.209.0.0", fixture: "network/dns/forwarder/update-ips.golden"},
		{args: "network dns forwarder update dnsf-invalid --name my-new-dns-forwarder", fixture: "network/dns/forwarder/update-dnsf-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsForwarderCreate() {
	tests := []CLITest{
		{args: "network dns forwarder create my-dns-forwarder", fixture: "network/dns/forwarder/create-missing-flags.golden", exitCode: 1},
		{args: "network dns forwarder create dnsf-invalid-gateway --dns-server-ips 10.200.0.0 --gateway gw-123456 --domains abc.com", fixture: "network/dns/forwarder/create-invalid-gateway.golden", exitCode: 1},
		{args: "network dns forwarder create dnsf-duplicate --dns-server-ips 10.200.0.0 --gateway gw-123456 --domains abc.com", fixture: "network/dns/forwarder/create-duplicate.golden", exitCode: 1},
		{args: "network dns forwarder create dnsf-exceed-quota --dns-server-ips 10.200.0.0 --gateway gw-123456 --domains abc.com", fixture: "network/dns/forwarder/create-exceed-quota.golden", exitCode: 1},
		{args: "network dns forwarder create my-dns-forwarder --dns-server-ips 10.200.0.0 --gateway gw-123456 --domains abc.com,def.com,xyz.com", fixture: "network/dns/forwarder/create.golden"},
		{args: "network dns forwarder create --dns-server-ips 10.200.0.0 --gateway gw-123456 --domains abc.com,def.com,xyz.com", fixture: "network/dns/forwarder/create-no-name.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsForwarder_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network dns forwarder describe ""`, login: "cloud", fixture: "network/dns/forwarder/describe-autocomplete.golden"},
		{args: `__complete network dns forwarder delete ""`, login: "cloud", fixture: "network/dns/forwarder/delete-autocomplete.golden"},
		{args: `__complete network dns forwarder update ""`, login: "cloud", fixture: "network/dns/forwarder/update-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsRecordDelete() {
	tests := []CLITest{
		{args: "network dns record delete dnsrec-12345", input: "y\n", fixture: "network/dns/record/delete.golden"},
		{args: "network dns record delete dnsrec-12345 dnsrec-67890", input: "y\n", fixture: "network/dns/record/delete-multiple.golden"},
		{args: "network dns record delete dnsrec-invalid", fixture: "network/dns/record/delete-fail.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsRecordCreate() {
	tests := []CLITest{
		{args: "network dns record create --gateway gw-123456 --private-link-access-point ap-123456 --domain www.example.com", fixture: "network/dns/record/create.golden"},
		{args: "network dns record create my-dns-record --gateway gw-123456 --private-link-access-point ap-123456 --domain www.example.com", fixture: "network/dns/record/create-name.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsRecordDescribe() {
	tests := []CLITest{
		{args: "network dns record describe dnsrec-12345", fixture: "network/dns/record/describe.golden"},
		{args: "network dns record describe dnsrec-12345 --output json", fixture: "network/dns/record/describe-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsRecordList() {
	tests := []CLITest{
		{args: "network dns record list", fixture: "network/dns/record/list.golden"},
		{args: "network dns record list --output json", fixture: "network/dns/record/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsRecordUpdate() {
	tests := []CLITest{
		{args: "network dns record update dnsrec-12345", fixture: "network/dns/record/update-missing-flags.golden", exitCode: 1},
		{args: "network dns record update dnsrec-12345 --name my-new-dns-record", fixture: "network/dns/record/update.golden"},
		{args: "network dns record update dnsrec-12345 --private-link-access-point ap-67890", fixture: "network/dns/record/update-access-point.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkDnsRecord_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network dns record describe ""`, login: "cloud", fixture: "network/dns/record/describe-autocomplete.golden"},
		{args: `__complete network dns record delete ""`, login: "cloud", fixture: "network/dns/record/delete-autocomplete.golden"},
		{args: `__complete network dns record update ""`, login: "cloud", fixture: "network/dns/record/update-autocomplete.golden"},
		{args: `__complete network dns record create --private-link-access-point ""`, login: "cloud", fixture: "network/dns/record/create-autocomplete-private-link-access-point-flag.golden"},
		{args: `__complete network dns record create --gateway ""`, login: "cloud", fixture: "network/dns/record/create-autocomplete-gateway-flag.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkAccessPointPrivateLinkEgressEndpointDelete() {
	tests := []CLITest{
		{args: "network access-point private-link egress-endpoint delete ap-12345", input: "y\n", fixture: "network/access-point/private-link/egress-endpoint/delete.golden"},
		{args: "network access-point private-link egress-endpoint delete ap-12345 ap-67890", input: "y\n", fixture: "network/access-point/private-link/egress-endpoint/delete-multiple.golden"},
		{args: "network access-point private-link egress-endpoint delete ap-invalid", fixture: "network/access-point/private-link/egress-endpoint/delete-fail.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkAccessPointPrivateLinkEgressEndpointCreate() {
	tests := []CLITest{
		{args: "network access-point private-link egress-endpoint create --cloud aws --gateway gw-123456 --service com.amazonaws.vpce.us-west-2.vpce-svc-00000000000000000 --high-availability", fixture: "network/access-point/private-link/egress-endpoint/create-aws.golden"},
		{args: "network access-point private-link egress-endpoint create my-egress-endpoint --cloud azure --gateway gw-123456 --service /subscriptions/0000000/resourceGroups/plsRgName/providers/Microsoft.Network/privateLinkServices/privateLinkServiceName --subresource subresource1", fixture: "network/access-point/private-link/egress-endpoint/create-azure.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkAccessPointPrivateLinkEgressEndpointDescribe() {
	tests := []CLITest{
		{args: "network access-point private-link egress-endpoint describe ap-12345", fixture: "network/access-point/private-link/egress-endpoint/describe-aws.golden"},
		{args: "network access-point private-link egress-endpoint describe ap-67890", fixture: "network/access-point/private-link/egress-endpoint/describe-azure.golden"},
		{args: "network access-point private-link egress-endpoint describe ap-12345 --output json", fixture: "network/access-point/private-link/egress-endpoint/describe-aws-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkAccessPointPrivateLinkEgressEndpointList() {
	tests := []CLITest{
		{args: "network access-point private-link egress-endpoint list", fixture: "network/access-point/private-link/egress-endpoint/list.golden"},
		{args: "network access-point private-link egress-endpoint list --output json", fixture: "network/access-point/private-link/egress-endpoint/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkAccessPointPrivateLinkEgressEndpointUpdate() {
	tests := []CLITest{
		{args: "network access-point private-link egress-endpoint update ap-12345 --name my-new-aws-egress-access-point", input: "y\n", fixture: "network/access-point/private-link/egress-endpoint/update-aws.golden"},
		{args: "network access-point private-link egress-endpoint update ap-67890 --name my-new-azure-egress-access-point", input: "y\n", fixture: "network/access-point/private-link/egress-endpoint/update-azure.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestNetworkAccessPointPrivateLinkEgressEndpoint_Autocomplete() {
	tests := []CLITest{
		{args: `__complete network access-point private-link egress-endpoint describe ""`, login: "cloud", fixture: "network/access-point/private-link/egress-endpoint/describe-autocomplete.golden"},
		{args: `__complete network access-point private-link egress-endpoint delete ""`, login: "cloud", fixture: "network/access-point/private-link/egress-endpoint/delete-autocomplete.golden"},
		{args: `__complete network access-point private-link egress-endpoint update ""`, login: "cloud", fixture: "network/access-point/private-link/egress-endpoint/update-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
