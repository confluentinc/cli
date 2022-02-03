package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (s *CLITestSuite) TestAPIKey() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --resource lkc-bob", login: "default", fixture: "api-key/1.golden"}, // MYKEY3
		{args: "api-key list --resource lkc-bob", fixture: "api-key/2.golden"},
		{args: "api-key list --resource lkc-abc", fixture: "api-key/3.golden"},
		{args: "api-key update MYKEY1 --description first-key", fixture: "api-key/4.golden"},
		{args: "api-key list --resource lkc-bob", fixture: "api-key/5.golden"},

		// list json and yaml output
		{args: "api-key list", fixture: "api-key/6.golden"},
		{args: "api-key list -o json", fixture: "api-key/7.golden"},
		{args: "api-key list -o yaml", fixture: "api-key/8.golden"},

		// create api key for kafka cluster
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/9.golden"},
		{args: "api-key create --description my-cool-app --resource lkc-cool1", fixture: "api-key/10.golden"}, // MYKEY4
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/11.golden"},

		// create api key for other kafka cluster
		{args: "api-key create --description my-other-app --resource lkc-other1", fixture: "api-key/12.golden"}, // MYKEY5
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/11.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "api-key/13.golden"},

		// create api key for ksql cluster
		{args: "api-key create --description my-ksql-app --resource lksqlc-ksql1", fixture: "api-key/14.golden"}, // MYKEY6
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/11.golden"},
		{args: "api-key list --resource lksqlc-ksql1", fixture: "api-key/15.golden"},

		// create api key for schema registry cluster
		{args: "api-key create --resource lsrc-1", fixture: "api-key/16.golden"}, // MYKEY7
		{args: "api-key list --resource lsrc-1", fixture: "api-key/17.golden"},

		// create cloud api key
		{args: "api-key create --resource cloud", fixture: "api-key/18.golden"}, // MYKEY8
		{args: "api-key list --resource cloud", fixture: "api-key/19.golden"},

		// use an api key for kafka cluster
		{args: "api-key use MYKEY4 --resource lkc-cool1", fixture: "api-key/20.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/21.golden"},

		// use an api key for other kafka cluster
		{args: "api-key use MYKEY5 --resource lkc-other1", fixture: "api-key/22.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/21.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "api-key/23.golden"},

		// delete api key that is in use
		{args: "api-key delete MYKEY5", fixture: "api-key/24.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "api-key/25.golden"},

		// store an api-key for kafka cluster
		{args: "api-key store UIAPIKEY100 @test/fixtures/input/api-key/UIAPISECRET100.txt --resource lkc-cool1", fixture: "api-key/26.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/21.golden"},

		// store an api-key for other kafka cluster
		{args: "api-key store UIAPIKEY101 @test/fixtures/input/api-key/UIAPISECRET101.txt --resource lkc-other1", fixture: "api-key/27.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "api-key/21.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "api-key/28.golden"},

		// store exists already error
		{args: "api-key store UIAPIKEY101 @test/fixtures/input/api-key/UIAPISECRET101.txt --resource lkc-other1", fixture: "api-key/override-error.golden", wantErrCode: 1},

		// store an api-key for ksql cluster (not yet supported)
		//{args: "api-key store UIAPIKEY103 UIAPISECRET103 --resource lksqlc-ksql1", fixture: "empty.golden"},
		//{args: "api-key list --resource lksqlc-ksql1", fixture: "api-key/10.golden"},
		// TODO: change test back once api-key store and use command allows for non kafka clusters
		{args: "api-key store UIAPIKEY103 UIAPISECRET103 --resource lksqlc-ksql1", fixture: "api-key/29.golden", wantErrCode: 1},
		{args: "api-key use UIAPIKEY103 --resource lksqlc-ksql1", fixture: "api-key/29.golden", wantErrCode: 1},

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
		{name: "error listing api keys for non-existent service account", args: "api-key list --service-account sa-123456", fixture: "api-key/39.golden"},

		// create api-key for audit log
		{args: "api-key create --resource lkc-cool1 --service-account sa-1337 --description auditlog-key", fixture: "api-key/40.golden"}, // MYKEY11
		{args: "api-key list", fixture: "api-key/41.golden"},
		{args: "api-key create --resource lkc-cool1 --service-account sa-1337 --description auditlog-key", fixture: "api-key/42.golden", disableAuditLog: true}, // MYKEY11
		{args: "api-key list", fixture: "api-key/43.golden", disableAuditLog: true},

		// create json yaml output
		{args: "api-key create --description human-output --resource lkc-other1", fixture: "api-key/44.golden"},
		{args: "api-key create --description json-output --resource lkc-other1 -o json", fixture: "api-key/45.golden"},
		{args: "api-key create --description yaml-output --resource lkc-other1 -o yaml", fixture: "api-key/46.golden"},

		// store: error handling
		{name: "error if storing unknown api key", args: "api-key store UNKNOWN @test/fixtures/input/api-key/UIAPISECRET100.txt --resource lkc-cool1", fixture: "api-key/47.golden"},
		{name: "error if storing api key with existing secret", args: "api-key store UIAPIKEY100 NEWSECRET --resource lkc-cool1", fixture: "api-key/48.golden"},
		{name: "succeed if forced to overwrite existing secret", args: "api-key store -f UIAPIKEY100 NEWSECRET --resource lkc-cool1", fixture: "api-key/49.golden",
			wantFunc: func(t *testing.T) {
				cfg := v1.New(&config.Params{})
				cfg, err := load.LoadAndMigrate(cfg)
				require.NoError(t, err)
				ctx := cfg.Context()
				require.NotNil(t, ctx)
				kcc := ctx.KafkaClusterContext.GetKafkaClusterConfig("lkc-cool1")
				pair := kcc.APIKeys["UIAPIKEY100"]
				require.NotNil(t, pair)
				require.Equal(t, "NEWSECRET", pair.Secret)
			}},

		// use: error handling
		{name: "error if using non-existent api-key", args: "api-key use UNKNOWN --resource lkc-cool1", fixture: "api-key/50.golden"},
		{name: "error if using api-key for wrong cluster", args: "api-key use MYKEY2 --resource lkc-cool1", fixture: "api-key/51.golden"},
		{name: "error if using api-key without existing secret", args: "api-key use UIAPIKEY103 --resource lkc-cool1", fixture: "api-key/52.golden"},

		// more errors
		{args: "api-key use UIAPIKEY103", fixture: "api-key/53.golden", wantErrCode: 1},
		{args: "api-key create", fixture: "api-key/54.golden", wantErrCode: 1},
		{args: "api-key use UIAPIKEY103 --resource lkc-unknown", fixture: "api-key/resource-unknown-error.golden", wantErrCode: 1},
		{args: "api-key create --resource lkc-unknown", fixture: "api-key/resource-unknown-error.golden", wantErrCode: 1},
	}

	resetConfiguration(s.T())

	for _, tt := range tests {
		tt.workflow = true
		s.runCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestAPIKeyCreate_ServiceAccountNotValid() {
	tt := CLITest{args: "api-key create --resource lkc-ab123 --service-account sa-123456", login: "default", fixture: "api-key/55.golden", wantErrCode: 1}
	s.runCcloudTest(tt)
}
