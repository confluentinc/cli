package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func (s *CLITestSuite) TestAPIKeyCommands() {
	loginURL := serve(s.T()).URL

	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --cluster bob", login: "default", fixture: "apikey1.golden"}, // MYKEY3
		{args: "api-key list", useKafka: "bob", fixture: "apikey2.golden"},
		{args: "api-key list", useKafka: "abc", fixture: "apikey3.golden"},

		// create api key for active kafka cluster
		{args: "kafka cluster use lkc-cool1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey4.golden"},
		{args: "api-key create --description my-cool-app", fixture: "apikey5.golden"}, // MYKEY4
		{args: "api-key list", fixture: "apikey6.golden"},

		// create api key for other kafka cluster
		{args: "api-key create --description my-other-app --cluster lkc-other1", fixture: "apikey7.golden"}, // MYKEY5
		{args: "api-key list", fixture: "apikey6.golden"},
		{args: "api-key list --cluster lkc-other1", fixture: "apikey8.golden"},

		// create api key for non-kafka cluster
		{args: "api-key create --description my-ksql-app --cluster lksqlc-ksql1", fixture: "apikey9.golden"}, // MYKEY6
		{args: "api-key list", fixture: "apikey6.golden"},
		{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey10.golden"},

		// use an api key for active kafka cluster
		// TODO: use MYKEY2 should error since its not in lkc-cool1. right now i think it fails silently. no * by anything in list
		// TODO: switch to "abc" then use MYKEY2 since it was created outside CLI and we'll need to prompt for the secret
		{args: "api-key use MYKEY4", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey12.golden"}, // TODO: this should show * MYKEY4

		// use an api key for other kafka cluster
		{args: "api-key use MYKEY5 --cluster lkc-other1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey12.golden"},
		{args: "api-key list --cluster lkc-other1", fixture: "apikey13.golden"},

		// use an api key for non-kafka cluster
		{args: "api-key use MYKEY6 --cluster lksqlc-ksql1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey12.golden"},
		{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey14.golden"},

		// store an api-key for active kafka cluster
		{args: "api-key store UIAPIKEY100 UIAPISECRET100", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey12.golden"},

		// store an api-key for other kafka cluster
		{args: "api-key store UIAPIKEY101 UIAPISECRET101 --cluster lkc-other1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey12.golden"},
		{args: "api-key list --cluster lkc-other1", fixture: "apikey13.golden"},

		// store an api-key for non-kafka cluster
		{args: "api-key store UIAPIKEY102 UIAPISECRET102 --cluster lksqlc-ksql1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey12.golden"},
		{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey15.golden"},

		// store: error handling
		{name: "error if storing unknown api key", args: "api-key store UNKNOWN SECRET", fixture: "apikey16.golden"},
		{name: "error if storing api key with existing secret", args: "api-key store EXISTING NEWSECRET", fixture: "apikey17.golden"},
		{name: "succeed if forced to overwrite existing secret", args: "api-key store -f UIAPIKEY101 NEWSECRET", fixture: "empty.golden",
			wantFunc: func(t *testing.T) {
				logger := log.New()
				cfg := config.New(&config.Config{
					CLIName: binaryName,
					Logger:  logger,
				})
				require.NoError(t, cfg.Load())
				ctx, err := cfg.Context()
				require.NoError(t, err)
				kcc := ctx.KafkaClusters["lkc-cool1"]
				pair := kcc.APIKeys["UIAPIKEY101"]
				require.Equal(t, "NEWSECRET", pair.Secret)
			}},
	}
	resetConfiguration(s.T())
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.workflow = true
		s.runTest(tt, loginURL, serveKafkaAPI(s.T()).URL)
	}
}
