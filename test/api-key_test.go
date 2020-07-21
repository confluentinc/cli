package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func (s *CLITestSuite) TestAPIKey() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --resource lkc-bob", login: "default", fixture: "apikey/1.golden"}, // MYKEY3
		{args: "api-key list --resource lkc-bob", fixture: "apikey/2.golden"},
		{args: "api-key list --resource lkc-abc", fixture: "apikey/3.golden"},
		{args: "api-key update MYKEY1 --description first-key", fixture: "apikey/40.golden"},
		{args: "api-key list --resource lkc-bob", fixture: "apikey/41.golden"},

		// list json and yaml output
		{args: "api-key list", fixture: "apikey/28.golden"},
		{args: "api-key list -o json", fixture: "apikey/29.golden"},
		{args: "api-key list -o yaml", fixture: "apikey/30.golden"},

		// create api key for kafka cluster
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/4.golden"},
		{args: "api-key create --description my-cool-app --resource lkc-cool1", fixture: "apikey/5.golden"}, // MYKEY4
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/6.golden"},

		// create api key for other kafka cluster
		{args: "api-key create --description my-other-app --resource lkc-other1", fixture: "apikey/7.golden"}, // MYKEY5
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/6.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "apikey/8.golden"},

		// create api key for ksql cluster
		{args: "api-key create --description my-ksql-app --resource lksqlc-ksql1", fixture: "apikey/9.golden"}, // MYKEY6
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/6.golden"},
		{args: "api-key list --resource lksqlc-ksql1", fixture: "apikey/10.golden"},

		// create api key for schema registry cluster
		{args: "api-key create --resource lsrc-1", fixture: "apikey/20.golden"}, // MYKEY7
		{args: "api-key list --resource lsrc-1", fixture: "apikey/21.golden"},

		// create cloud api key
		{args: "api-key create --resource cloud", fixture: "apikey/34.golden"}, // MYKEY8
		{args: "api-key list --resource cloud", fixture: "apikey/35.golden"},

		// use an api key for kafka cluster
		{args: "api-key use MYKEY4 --resource lkc-cool1", fixture: "apikey/45.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/11.golden"},

		// use an api key for other kafka cluster
		{args: "api-key use MYKEY5 --resource lkc-other1", fixture: "apikey/46.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/11.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "apikey/12.golden"},

		// delete api key that is in use
		{args: "api-key delete MYKEY5", fixture: "apikey/42.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "apikey/43.golden"},

		// store an api-key for kafka cluster
		{args: "api-key store UIAPIKEY100 @test/fixtures/input/UIAPISECRET100.txt --resource lkc-cool1", fixture: "apikey/47.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/11.golden"},

		// store an api-key for other kafka cluster
		{args: "api-key store UIAPIKEY101 @test/fixtures/input/UIAPISECRET101.txt --resource lkc-other1", fixture: "apikey/48.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/11.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "apikey/44.golden"},

		// store exists already error
		{args: "api-key store UIAPIKEY101 @test/fixtures/input/UIAPISECRET101.txt --resource lkc-other1", fixture: "apikey/override-error.golden", wantErrCode: 1},

		// store an api-key for ksql cluster (not yet supported)
		//{args: "api-key store UIAPIKEY103 UIAPISECRET103 --resource lksqlc-ksql1", fixture: "empty.golden"},
		//{args: "api-key list --resource lksqlc-ksql1", fixture: "apikey/10.golden"},
		// TODO: change test back once api-key store and use command allows for non kafka clusters
		{args: "api-key store UIAPIKEY103 UIAPISECRET103 --resource lksqlc-ksql1", fixture: "apikey/36.golden", wantErrCode: 1},
		{args: "api-key use UIAPIKEY103 --resource lksqlc-ksql1", fixture: "apikey/36.golden", wantErrCode: 1},

		// list all api-keys
		{args: "api-key list", fixture: "apikey/22.golden"},

		// list api-keys belonging to currently logged in user
		{args: "api-key list --current-user", fixture: "apikey/23.golden"},

		// create api-key for a service account
		{args: "api-key create --resource lkc-cool1 --service-account 99", fixture: "apikey/24.golden"},
		{args: "api-key list --current-user", fixture: "apikey/23.golden"},
		{args: "api-key list", fixture: "apikey/25.golden"},
		{args: "api-key list --service-account 99", fixture: "apikey/26.golden"},
		{args: "api-key list --resource lkc-cool1", fixture: "apikey/27.golden"},
		{args: "api-key list --resource lkc-cool1 --service-account 99", fixture: "apikey/26.golden"},

		// create json yaml output
		{args: "api-key create --description human-output --resource lkc-other1", fixture: "apikey/31.golden"},
		{args: "api-key create --description json-output --resource lkc-other1 -o json", fixture: "apikey/32.golden"},
		{args: "api-key create --description yaml-output --resource lkc-other1 -o yaml", fixture: "apikey/33.golden"},

		// store: error handling
		{name: "error if storing unknown api key", args: "api-key store UNKNOWN @test/fixtures/input/UIAPISECRET100.txt --resource lkc-cool1", fixture: "apikey/15.golden"},
		{name: "error if storing api key with existing secret", args: "api-key store UIAPIKEY100 NEWSECRET --resource lkc-cool1", fixture: "apikey/16.golden"},
		{name: "succeed if forced to overwrite existing secret", args: "api-key store -f UIAPIKEY100 NEWSECRET --resource lkc-cool1", fixture: "apikey/49.golden",
			wantFunc: func(t *testing.T) {
				logger := log.New()
				cfg := v3.New(&config.Params{
					CLIName:    "ccloud",
					MetricSink: nil,
					Logger:     logger,
				})
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
		{name: "error if using non-existent api-key", args: "api-key use UNKNOWN --resource lkc-cool1", fixture: "apikey/17.golden"},
		{name: "error if using api-key for wrong cluster", args: "api-key use MYKEY2 --resource lkc-cool1", fixture: "apikey/18.golden"},
		{name: "error if using api-key without existing secret", args: "api-key use UIAPIKEY103 --resource lkc-cool1", fixture: "apikey/19.golden"},

		// more errors
		{args: "api-key use UIAPIKEY103", fixture: "apikey/37.golden", wantErrCode: 1},
		{args: "api-key create", fixture: "apikey/38.golden", wantErrCode: 1},
		{args: "api-key use UIAPIKEY103 --resource lkc-unknown", fixture: "apikey/resource-unknown-error.golden", wantErrCode: 1},
		{args: "api-key create --resource lkc-unknown", fixture: "apikey/resource-unknown-error.golden", wantErrCode: 1},
	}

	resetConfiguration(s.T(), "ccloud")
	kafkaURL := serveKafkaAPI(s.T()).URL
	loginURL := serve(s.T(), kafkaURL).URL

	for _, tt := range tests {
		tt.workflow = true
		s.runCcloudTest(tt, loginURL)
	}
}
