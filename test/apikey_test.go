package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func (s *CLITestSuite) TestAPIKeyCommands() {
	kafkaAPIURL := serveKafkaAPI(s.T()).URL
	loginURL := serve(s.T(), kafkaAPIURL).URL

	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --resource bob", login: "default", fixture: "apikey1.golden"}, // MYKEY3
		{args: "api-key list", useKafka: "bob", fixture: "apikey2.golden"},
		{args: "api-key list", useKafka: "abc", fixture: "apikey3.golden"},

		// create api key for active kafka cluster
		{args: "kafka cluster use lkc-cool1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey4.golden"},
		{args: "api-key create --description my-cool-app", fixture: "apikey5.golden"}, // MYKEY4
		{args: "api-key list", fixture: "apikey6.golden"},

		// create api key for other kafka cluster
		{args: "api-key create --description my-other-app --resource lkc-other1", fixture: "apikey7.golden"}, // MYKEY5
		{args: "api-key list", fixture: "apikey6.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "apikey8.golden"},

		// create api key for non-kafka cluster
		{args: "api-key create --description my-ksql-app --resource lksqlc-ksql1", fixture: "apikey9.golden"}, // MYKEY6
		{args: "api-key list", fixture: "apikey6.golden"},
		{args: "api-key list --resource lksqlc-ksql1", fixture: "apikey10.golden"},

		// create api key for schema registry cluster
		{args: "api-key create --resource lsrc-1", fixture: "apikey20.golden"}, // MYKEY7
		{args: "api-key list --resource lsrc-1", fixture: "apikey21.golden"},

		// use an api key for active kafka cluster
		{args: "api-key use MYKEY4", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey11.golden"},

		// use an api key for other kafka cluster
		{args: "api-key use MYKEY5 --resource lkc-other1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey11.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "apikey12.golden"},

		// use an api key for non-kafka cluster
		{args: "api-key use MYKEY6 --resource lksqlc-ksql1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey11.golden"},
		{args: "api-key list --resource lksqlc-ksql1", fixture: "apikey13.golden"},

		// store an api-key for active kafka cluster
		{args: "api-key store UIAPIKEY100 UIAPISECRET100", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey11.golden"},

		// store an api-key for other kafka cluster
		{args: "api-key store UIAPIKEY101 UIAPISECRET101 --resource lkc-other1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey11.golden"},
		{args: "api-key list --resource lkc-other1", fixture: "apikey12.golden"},

		// store an api-key for non-kafka cluster
		{args: "api-key store UIAPIKEY102 UIAPISECRET102 --resource lksqlc-ksql1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey11.golden"},
		{args: "api-key list --resource lksqlc-ksql1", fixture: "apikey14.golden"},

		// store: error handling
		{name: "error if storing unknown api key", args: "api-key store UNKNOWN SECRET", fixture: "apikey15.golden"},
		{name: "error if storing api key with existing secret", args: "api-key store UIAPIKEY100 NEWSECRET", fixture: "apikey16.golden"},
		{name: "succeed if forced to overwrite existing secret", args: "api-key store -f UIAPIKEY100 NEWSECRET", fixture: "empty.golden",
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
				pair := kcc.APIKeys["UIAPIKEY100"]
				require.NotNil(t, pair)
				require.Equal(t, "NEWSECRET", pair.Secret)

			}},

		// use: error handling
		{name: "error if using non-existent api-key", args: "api-key use UNKNOWN", fixture: "apikey17.golden"},
		{name: "error if using api-key for wrong cluster", args: "api-key use MYKEY2", fixture: "apikey18.golden"},
		{name: "error if using api-key without existing secret", args: "api-key use UIAPIKEY103", fixture: "apikey19.golden"},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.workflow = true
		s.runCcloudTest(tt, loginURL, serveKafkaAPI(s.T()).URL)
	}
}
