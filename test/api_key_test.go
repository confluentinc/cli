package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/config"
)

func (s *CLITestSuite) TestApiKey() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --resource lkc-bob", login: "cloud", fixture: "api-key/1.golden"}, // MYKEY3
		{args: "api-key list --resource lkc-bob", fixture: "api-key/2.golden"},
		{args: "api-key list --resource lkc-abc", fixture: "api-key/3.golden"},
		{args: "api-key update MYKEY1 --description first-key", fixture: "api-key/4.golden"},
		{args: "api-key list --resource lkc-bob", fixture: "api-key/5.golden"},

		// list json and yaml output
		{args: "api-key list", fixture: "api-key/6.golden"},
		{args: "api-key list -o json", fixture: "api-key/7.golden"},
		{args: "api-key list -o yaml", fixture: "api-key/8.golden"},

		// create API key for kafka cluster
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/9.golden"},
		{args: "api-key create --description my-cool-app --resource lkc-cool1", fixture: "api-key/10.golden"}, // MYKEY4
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/11.golden"},

		// create API key for other kafka cluster
		{args: "api-key create --description my-other-app --resource lkc-other1", fixture: "api-key/12.golden"}, // MYKEY5
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/11.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "api-key/13.golden"},

		// create API key for KSQL cluster
		{args: "api-key create --description my-ksql-app --resource lksqlc-ksql1", fixture: "api-key/14.golden"}, // MYKEY6
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/11.golden"},
		{args: "api-key list --resource lksqlc-ksql1", fixture: "api-key/15.golden"},

		// create API key for schema registry cluster
		{args: "api-key create --resource lsrc-1234", fixture: "api-key/16.golden"}, // MYKEY7
		{args: "api-key list --resource lsrc-1234", fixture: "api-key/17.golden"},

		// create cloud API key
		{args: "api-key create --resource cloud", fixture: "api-key/18.golden"}, // MYKEY8
		{args: "api-key list --resource cloud", fixture: "api-key/19.golden"},

		// use an API key for kafka cluster
		{args: "api-key use MYKEY4", fixture: "api-key/20.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/21.golden"},

		// use an API key for other kafka cluster
		{args: "api-key use MYKEY5", fixture: "api-key/22.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/21.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "api-key/23.golden"},

		// delete API key that is in use
		{args: "api-key delete MYKEY5 --force", fixture: "api-key/delete/success.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "api-key/25.golden"},

		// store an API key for kafka cluster
		{args: "api-key store UIAPIKEY100 UIAPISECRET100 --resource lkc-cool1", fixture: "api-key/26.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/21.golden"},

		// store an API key for other kafka cluster
		{args: "api-key store UIAPIKEY101 UIAPISECRET101 --resource lkc-other1", fixture: "api-key/27.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/21.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "api-key/28.golden"},

		// store exists already error
		{args: "api-key store UIAPIKEY101 UIAPISECRET101 --resource lkc-other1", fixture: "api-key/override-error.golden", exitCode: 1},

		{args: "api-key store UIAPIKEY103 UIAPISECRET103 --resource lksqlc-ksql1", fixture: "api-key/61.golden", exitCode: 1},
		{args: "api-key use UIAPIKEY103", fixture: "api-key/29.golden", exitCode: 1},

		// list all api-keys
		{args: "api-key list", fixture: "api-key/30.golden"},

		// list api-keys belonging to currently logged in user
		{args: "api-key list --current-user", fixture: "api-key/31.golden"},

		// create api-key for a service account
		{args: "api-key create --resource lkc-cool1 --service-account sa-12345", fixture: "api-key/32.golden"}, // MYKEY9
		{args: "api-key list --current-user", fixture: "api-key/31.golden"},
		{args: "api-key list", fixture: "api-key/33.golden"},
		{args: "api-key list --service-account sa-12345", fixture: "api-key/34.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/35.golden"},
		{args: "api-key list --resource lkc-cool1 --service-account sa-12345", fixture: "api-key/36.golden"},
		{args: "api-key create --resource lkc-cool1 --service-account sa-12345", fixture: "api-key/37.golden"}, // MYKEY10
		{args: "api-key list --service-account sa-12345", fixture: "api-key/38.golden"},
		{name: "error listing API keys for non-existent service account", args: "api-key list --service-account sa-123456", fixture: "api-key/39.golden", exitCode: 1},

		// create api-key for audit log
		{args: "api-key create --resource lkc-cool1 --service-account sa-1337 --description auditlog-key", fixture: "api-key/40.golden"}, // MYKEY11
		{args: "api-key list", fixture: "api-key/41.golden"},
		{args: "api-key create --resource lkc-cool1 --service-account sa-1337 --description auditlog-key", fixture: "api-key/42.golden", disableAuditLog: true}, // MYKEY11
		{args: "api-key list", fixture: "api-key/43.golden", disableAuditLog: true},

		// create json yaml output
		{args: "api-key create --description human-output --resource lkc-other1", fixture: "api-key/44.golden"},
		{args: "api-key create --description json-output --resource lkc-other1 -o json", fixture: "api-key/45.golden"},
		{args: "api-key create --description yaml-output --resource lkc-other1 -o yaml", fixture: "api-key/46.golden"},

		// create api-key and use for the resource
		{args: "api-key create --description my-cool-app --resource lkc-cool1 --use", fixture: "api-key/60.golden"}, // MYKEY16

		// create tableflow API key
		{args: "api-key create --resource tableflow", fixture: "api-key/62.golden"}, // MYKEY17
		{args: "api-key list --resource tableflow", fixture: "api-key/63.golden"},

		// store: error handling
		{name: "error if storing unknown API key", args: "api-key store UNKNOWN UIAPISECRET100 --resource lkc-cool1", fixture: "api-key/47.golden", exitCode: 1},
		{name: "error if storing API key with existing secret", args: "api-key store UIAPIKEY100 NEWSECRET --resource lkc-cool1", fixture: "api-key/48.golden", exitCode: 1},
		{
			name: "succeed if forced to overwrite existing secret", args: "api-key store UIAPIKEY100 NEWSECRET --resource lkc-cool1 --force", fixture: "api-key/49.golden",
			wantFunc: func(t *testing.T) {
				cfg := config.New()
				err := cfg.Load()
				require.NoError(t, err)
				ctx := cfg.Context()
				require.NotNil(t, ctx)
				kcc := ctx.KafkaClusterContext.GetKafkaClusterConfig("lkc-cool1")
				err = kcc.DecryptAPIKeys()
				require.NoError(t, err)
				pair := kcc.APIKeys["UIAPIKEY100"]
				require.NotNil(t, pair)
				require.Equal(t, "NEWSECRET", pair.Secret)
			},
		},

		// use: error handling
		{name: "error if using non-existent api-key", args: "api-key use UNKNOWN", fixture: "api-key/50.golden", exitCode: 1},
		{name: "error if using api-key for wrong cluster", args: "api-key use MYKEY2", fixture: "api-key/51.golden", exitCode: 1},
		{name: "error if using api-key without existing secret", args: "api-key use UIAPIKEY103", fixture: "api-key/52.golden", exitCode: 1},

		// more errors
		{args: "api-key use UIAPIKEY103", fixture: "api-key/53.golden", exitCode: 1},
		{args: "api-key create", fixture: "api-key/54.golden", exitCode: 1},
		{args: "api-key use UIAPIKEY103 --resource lkc-unknown", fixture: "api-key/resource-unknown-error.golden", exitCode: 1},
		{args: "api-key create --resource lkc-unknown", fixture: "api-key/resource-unknown-error.golden", exitCode: 1},

		// test multicluster keys
		{name: "listing multicluster API keys", args: "api-key list", login: "cloud", env: []string{fmt.Sprintf("%s=multicluster-key-org", pauth.ConfluentCloudOrganizationId)}, fixture: "api-key/56.golden"},
		{name: "listing multicluster API keys with --resource field", args: "api-key list --resource lsrc-1234", login: "cloud", env: []string{fmt.Sprintf("%s=multicluster-key-org", pauth.ConfluentCloudOrganizationId)}, fixture: "api-key/57.golden"},
		{name: "listing multicluster API keys with --current-user field", args: "api-key list --current-user", login: "cloud", env: []string{fmt.Sprintf("%s=multicluster-key-org", pauth.ConfluentCloudOrganizationId)}, fixture: "api-key/58.golden"},
		{name: "listing multicluster API keys with --service-account field", args: "api-key list --service-account sa-12345", login: "cloud", env: []string{fmt.Sprintf("%s=multicluster-key-org", pauth.ConfluentCloudOrganizationId)}, fixture: "api-key/59.golden"},
	}

	resetConfiguration(s.T(), false)

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestApiKeyCreate() {
	tests := []CLITest{
		{args: "api-key create --resource flink --cloud aws --region us-east-1", fixture: "api-key/create-flink.golden"},
		{args: "api-key create --resource lkc-ab123 --service-account sa-123456", fixture: "api-key/55.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestApiKeyDescribe() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: "api-key describe MYKEY1", fixture: "api-key/describe.golden"},
		{args: "api-key describe MYKEY1 -o json", fixture: "api-key/describe-json.golden"},
		{args: "api-key describe MULTICLUSTERKEY1", fixture: "api-key/describe-multicluster.golden", env: []string{fmt.Sprintf("%s=multicluster-key-org", pauth.ConfluentCloudOrganizationId)}},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestApiKeyDelete() {
	tests := []CLITest{
		// delete multiple API keys
		{args: "api-key delete MYKEY7 MYKEY8 MYKEY19", fixture: "api-key/delete/multiple-fail.golden", exitCode: 1},
		{args: "api-key delete MYKEY6 MYKEY18 MYKEY19", fixture: "api-key/delete/multiple-fail-plural.golden", exitCode: 1},
		{args: "api-key delete MYKEY7 MYKEY8", input: "n\n", fixture: "api-key/delete/multiple-refuse.golden"},
		{args: "api-key delete MYKEY7 MYKEY8", input: "y\n", fixture: "api-key/delete/multiple-success.golden"},
	}

	resetConfiguration(s.T(), false)

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestApiKeyList() {
	tests := []CLITest{
		{args: "api-key list --resource lkc-dne", env: []string{fmt.Sprintf("%s=no-environment-user@example.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}, fixture: "api-key/no-env.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestApiKey_Autocomplete() {
	tests := []CLITest{
		{args: `__complete api-key create --resource ""`, fixture: "api-key/create-resource-autocomplete.golden"},
		{args: `__complete api-key describe ""`, fixture: "api-key/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
