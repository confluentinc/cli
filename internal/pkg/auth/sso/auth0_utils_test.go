package sso

import (
	"testing"

	testserver "github.com/confluentinc/cli/test/test-server"
	"github.com/stretchr/testify/assert"
)

func TestGetCCloudEnvFromBaseUrl(t *testing.T) {
	for baseUrl, want := range map[string]string{
		"":                                "prod",
		":no-scheme-error.com":            "prod",
		"confluent.cloud":                 "prod",
		"default-to-prod.com":             "prod",
		"https://confluent.cloud":         "prod",
		"https://confluent.cloud/":        "prod",
		"https://devel.cpdev.cloud":       "devel",
		"https://devel.cpdev.cloud/":      "devel",
		"https://prefix.priv.cpdev.cloud": "cpd",
		"https://stag.cpdev.cloud":        "stag",
		"https://stag.cpdev.cloud/":       "stag",
		testserver.TestCloudUrl.String():  "test",
	} {
		env := getCCloudEnvFromBaseUrl(baseUrl)
		assert.Equal(t, want, env)
	}
}
