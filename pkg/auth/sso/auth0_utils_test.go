package sso

import (
	"testing"

	"github.com/stretchr/testify/assert"

	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

func TestGetCCloudEnvFromBaseUrl(t *testing.T) {
	for baseUrl, want := range map[string]string{
		"":                                         "prod",
		":no-scheme-error.com":                     "prod",
		"confluent.cloud":                          "prod",
		"default-to-prod.com":                      "prod",
		"https://confluent.cloud":                  "prod",
		"https://confluent.cloud/":                 "prod",
		"https://confluentgov.com":                 "prod-us-gov",
		"https://confluentgov.com/":                "prod-us-gov",
		"https://infra.confluentgov-internal.com":  "infra-us-gov",
		"https://infra.confluentgov-internal.com/": "infra-us-gov",
		"https://devel.confluentgov-internal.com":  "devel-us-gov",
		"https://devel.confluentgov-internal.com/": "devel-us-gov",
		"https://devel.cpdev.cloud":                "devel",
		"https://devel.cpdev.cloud/":               "devel",
		"https://stag.cpdev.cloud":                 "stag",
		"https://stag.cpdev.cloud/":                "stag",
		testserver.TestCloudUrl.String():           "test",
	} {
		env := GetCCloudEnvFromBaseUrl(baseUrl)
		assert.Equal(t, want, env)
	}
}
