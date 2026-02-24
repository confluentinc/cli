package test

func (s *CLITestSuite) TestEndpointList() {
	tests := []CLITest{
		// Basic list with required flags
		{
			args:    "endpoint list --service KAFKA --environment env-00000",
			fixture: "endpoint/list.golden",
		},

		// List with cloud filter
		{
			args:    "endpoint list --service KAFKA --environment env-00000 --cloud AWS",
			fixture: "endpoint/list-cloud-filter.golden",
		},

		// List with region filter
		{
			args:    "endpoint list --service KAFKA --environment env-00000 --region us-west-2",
			fixture: "endpoint/list-region-filter.golden",
		},

		// List private endpoints only
		{
			args:    "endpoint list --service KAFKA --environment env-00000 --is-private=true",
			fixture: "endpoint/list-private.golden",
		},

		// List public endpoints only
		{
			args:    "endpoint list --service KAFKA --environment env-00000 --is-private=false",
			fixture: "endpoint/list-public.golden",
		},

		// List with resource filter
		{
			args:    "endpoint list --service KAFKA --environment env-00000 --resource lkc-abc123",
			fixture: "endpoint/list-resource-filter.golden",
		},

		// List with multiple filters
		{
			args:    "endpoint list --service KAFKA --environment env-00000 --cloud AWS --region us-west-2 --is-private=true",
			fixture: "endpoint/list-multiple-filters.golden",
		},

		// List Schema Registry endpoints
		{
			args:    "endpoint list --service SCHEMA_REGISTRY --environment env-00000",
			fixture: "endpoint/list-schema-registry.golden",
		},

		// JSON output
		{
			args:    "endpoint list --service KAFKA --environment env-00000 --output json",
			fixture: "endpoint/list-json.golden",
		},

		// YAML output
		{
			args:    "endpoint list --service KAFKA --environment env-00000 --output yaml",
			fixture: "endpoint/list-yaml.golden",
		},

		// Missing required service flag
		{
			args:     "endpoint list --environment env-00000",
			fixture:  "endpoint/list-missing-service.golden",
			exitCode: 1,
		},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
