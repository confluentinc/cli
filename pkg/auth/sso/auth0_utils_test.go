package sso

import (
	"testing"

	"github.com/stretchr/testify/assert"

	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

func TestGetCCloudEnvFromBaseUrl(t *testing.T) {
	for url, expected := range map[string]string{
		"":                                               "prod",
		":no-scheme-error.com":                           "prod",
		"confluent.cloud":                                "prod",
		"default-to-prod.com":                            "prod",
		"https://confluent.cloud":                        "prod",
		"https://confluent.cloud/":                       "prod",
		"https://confluentgov.com":                       "prod-us-gov",
		"https://us-east-1.confluentgov.com":             "prod-us-gov",
		"https://devel-1.confluentgov-internal.com":      "devel-us-gov",
		"https://devel.confluentgov-internal.com":        "devel-us-gov",
		"https://devel.cpdev.cloud":                      "devel",
		"https://east-1.devel.confluentgov-internal.com": "devel-us-gov",
		"https://infra.confluentgov-internal.com":        "infra-us-gov",
		"https://stag.cpdev.cloud":                       "stag",
		"https://dr-test.cpdev.cloud":                    "dr-test",
		testserver.TestCloudUrl.String():                 "test",
	} {
		actual := GetCCloudEnvFromBaseUrl(url)
		assert.Equal(t, expected, actual, url)
	}
}
