package api_key

import (
	"github.com/confluentinc/cli/test"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type APIKeyTestSuite struct {
	test.CLITestSuite
}

func TestAPIKey(t *testing.T) {
	suite.Run(t, new(APIKeyTestSuite))
}

func (s *APIKeyTestSuite) SetupSuite() {
	s.CLITestSuite.SetupSuite()
}

func (s *APIKeyTestSuite) TestAPIKey() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []test.CLITest{
		{Args: "api-key create --resource lkc-bob", Login: "default", Fixture: "apikey/1.golden"}, // MYKEY3
		{Args: "api-key list --resource lkc-bob", Fixture: "apikey/2.golden"},
		{Args: "api-key list --resource lkc-abc", Fixture: "apikey/3.golden"},
		{Args: "api-key update MYKEY1 --description first-key", Fixture: "apikey/40.golden"},
		{Args: "api-key list --resource lkc-bob", Fixture: "apikey/41.golden"},

		// list json and yaml output
		{Args: "api-key list", Fixture: "apikey/28.golden"},
		{Args: "api-key list -o json", Fixture: "apikey/29.golden"},
		{Args: "api-key list -o yaml", Fixture: "apikey/30.golden"},

		// create api key for kafka cluster
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/4.golden"},
		{Args: "api-key create --description my-cool-app --resource lkc-cool1", Fixture: "apikey/5.golden"}, // MYKEY4
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/6.golden"},

		// create api key for other kafka cluster
		{Args: "api-key create --description my-other-app --resource lkc-other1", Fixture: "apikey/7.golden"}, // MYKEY5
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/6.golden"},
		{Args: "api-key list --resource lkc-other1", Fixture: "apikey/8.golden"},

		// create api key for ksql cluster
		{Args: "api-key create --description my-ksql-app --resource lksqlc-ksql1", Fixture: "apikey/9.golden"}, // MYKEY6
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/6.golden"},
		{Args: "api-key list --resource lksqlc-ksql1", Fixture: "apikey/10.golden"},

		// create api key for schema registry cluster
		{Args: "api-key create --resource lsrc-1", Fixture: "apikey/20.golden"}, // MYKEY7
		{Args: "api-key list --resource lsrc-1", Fixture: "apikey/21.golden"},

		// create cloud api key
		{Args: "api-key create --resource cloud", Fixture: "apikey/34.golden"}, // MYKEY8
		{Args: "api-key list --resource cloud", Fixture: "apikey/35.golden"},

		// use an api key for kafka cluster
		{Args: "api-key use MYKEY4 --resource lkc-cool1", Fixture: "apikey/45.golden"},
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/11.golden"},

		// use an api key for other kafka cluster
		{Args: "api-key use MYKEY5 --resource lkc-other1", Fixture: "apikey/46.golden"},
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/11.golden"},
		{Args: "api-key list --resource lkc-other1", Fixture: "apikey/12.golden"},

		// delete api key that is in use
		{Args: "api-key delete MYKEY5", Fixture: "apikey/42.golden"},
		{Args: "api-key list --resource lkc-other1", Fixture: "apikey/43.golden"},

		// store an api-key for kafka cluster
		{Args: "api-key store UIAPIKEY100 @test/Fixtures/input/UIAPISECRET100.txt --resource lkc-cool1", Fixture: "apikey/47.golden"},
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/11.golden"},

		// store an api-key for other kafka cluster
		{Args: "api-key store UIAPIKEY101 @test/Fixtures/input/UIAPISECRET101.txt --resource lkc-other1", Fixture: "apikey/48.golden"},
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/11.golden"},
		{Args: "api-key list --resource lkc-other1", Fixture: "apikey/44.golden"},

		// store exists already error
		{Args: "api-key store UIAPIKEY101 @test/Fixtures/input/UIAPISECRET101.txt --resource lkc-other1", Fixture: "apikey/override-error.golden", WantErrCode: 1},

		// store an api-key for ksql cluster (not yet supported)
		//{Args: "api-key store UIAPIKEY103 UIAPISECRET103 --resource lksqlc-ksql1", Fixture: "empty.golden"},
		//{Args: "api-key list --resource lksqlc-ksql1", Fixture: "apikey/10.golden"},
		// TODO: change test back once api-key store and use command allows for non kafka clusters
		{Args: "api-key store UIAPIKEY103 UIAPISECRET103 --resource lksqlc-ksql1", Fixture: "apikey/36.golden", WantErrCode: 1},
		{Args: "api-key use UIAPIKEY103 --resource lksqlc-ksql1", Fixture: "apikey/36.golden", WantErrCode: 1},

		// list all api-keys
		{Args: "api-key list", Fixture: "apikey/22.golden"},

		// list api-keys belonging to currently logged in user
		{Args: "api-key list --current-user", Fixture: "apikey/23.golden"},

		// create api-key for a service account
		{Args: "api-key create --resource lkc-cool1 --service-account 12345", Fixture: "apikey/24.golden"},
		{Args: "api-key list --current-user", Fixture: "apikey/23.golden"},
		{Args: "api-key list", Fixture: "apikey/25.golden"},
		{Args: "api-key list --service-account 12345", Fixture: "apikey/26.golden"},
		{Args: "api-key list --resource lkc-cool1", Fixture: "apikey/27.golden"},
		{Args: "api-key list --resource lkc-cool1 --service-account 12345", Fixture: "apikey/50.golden"},
		{Args: "api-key create --resource lkc-cool1 --service-account sa-12345", Fixture: "apikey/51.golden"},
		{Args: "api-key list --service-account sa-12345", Fixture: "apikey/52.golden"},

		// create json yaml output
		{Args: "api-key create --description human-output --resource lkc-other1", Fixture: "apikey/31.golden"},
		{Args: "api-key create --description json-output --resource lkc-other1 -o json", Fixture: "apikey/32.golden"},
		{Args: "api-key create --description yaml-output --resource lkc-other1 -o yaml", Fixture: "apikey/33.golden"},

		// store: error handling
		{Name: "error if storing unknown api key", Args: "api-key store UNKNOWN @test/Fixtures/input/UIAPISECRET100.txt --resource lkc-cool1", Fixture: "apikey/15.golden"},
		{Name: "error if storing api key with existing secret", Args: "api-key store UIAPIKEY100 NEWSECRET --resource lkc-cool1", Fixture: "apikey/16.golden"},
		{Name: "succeed if forced to overwrite existing secret", Args: "api-key store -f UIAPIKEY100 NEWSECRET --resource lkc-cool1", Fixture: "apikey/49.golden",
			WantFunc: func(t *testing.T) {
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
		{Name: "error if using non-existent api-key", Args: "api-key use UNKNOWN --resource lkc-cool1", Fixture: "apikey/17.golden"},
		{Name: "error if using api-key for wrong cluster", Args: "api-key use MYKEY2 --resource lkc-cool1", Fixture: "apikey/18.golden"},
		{Name: "error if using api-key without existing secret", Args: "api-key use UIAPIKEY103 --resource lkc-cool1", Fixture: "apikey/19.golden"},

		// more errors
		{Args: "api-key use UIAPIKEY103", Fixture: "apikey/37.golden", WantErrCode: 1},
		{Args: "api-key create", Fixture: "apikey/38.golden", WantErrCode: 1},
		{Args: "api-key use UIAPIKEY103 --resource lkc-unknown", Fixture: "apikey/resource-unknown-error.golden", WantErrCode: 1},
		{Args: "api-key create --resource lkc-unknown", Fixture: "apikey/resource-unknown-error.golden", WantErrCode: 1},
	}

	test.ResetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.Workflow = true
		s.RunCcloudTest(tt)
	}
}
