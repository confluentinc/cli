package admin

import (
	"testing"

	"github.com/confluentinc/cli/v3/pkg/config"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76"
)

func TestInitStripe(t *testing.T) {
	tests := []struct {
		platformName string
		isLive       bool
	}{
		{
			platformName: "confluent.cloud",
			isLive:       true,
		},
		{platformName: "stag.cpdev.cloud"},
		{platformName: testserver.TestV2CloudUrl.Host},
	}

	for _, test := range tests {
		cfg := &config.Config{
			CurrentContext: "context",
			Contexts:       map[string]*config.Context{"context": {Platform: &config.Platform{Name: test.platformName}}},
		}
		initStripe(cfg)

		expected := stripeTestKey
		if test.isLive {
			expected = stripeLiveKey
		}
		require.Equal(t, expected, stripe.Key)
	}
}
