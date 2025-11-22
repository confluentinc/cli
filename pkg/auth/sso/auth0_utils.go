package sso

import (
	"net/url"
	"slices"
	"strings"

	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

var auth0ClientIds = map[string]string{
	"devel":        "sPhOuMMVRSFFR7HfB606KLxf1RAU4SXg",
	"infra-us-gov": "0oa73yrenxpdtNXEe4h7",
	"prod-us-gov":  "0oa41ih4ms3TVVAT04h7",
	"devel-us-gov": "0oaa5al9x9Rf0rF3j1d7",
	"prod":         "oX2nvSKl5jvBKVgwehZfvR4K8RhsZIEs",
	"stag":         "8RxQmZEYtEDah4MTIIzl4hGGeFwdJS6w",
	"test":         "00000000000000000000000000000000",
	"dr-test":      "5nyOPfaw4CDyMZFCu2AgtPVNKoO8kpKB",
}

func GetAuth0CCloudClientIdFromBaseUrl(baseUrl string) string {
	env := GetCCloudEnvFromBaseUrl(baseUrl)
	return auth0ClientIds[env]
}

func GetCCloudEnvFromBaseUrl(baseUrl string) string {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "prod"
	}

	if strings.HasSuffix(u.Host, "cpdev.cloud") {
		if strings.Contains(u.Host, "devel") {
			return "devel"
		} else if strings.Contains(u.Host, "stag") {
			return "stag"
		} else if strings.Contains(u.Host, "dr-test") {
			return "dr-test"
		}
	} else if strings.HasSuffix(u.Host, "confluentgov.com") {
		return "prod-us-gov"
	} else if strings.HasSuffix(u.Host, "confluentgov-internal.com") {
		if strings.Contains(u.Host, "devel") {
			return "devel-us-gov"
		} else if strings.Contains(u.Host, "infra") {
			return "infra-us-gov"
		}
	} else if u.Host == testserver.TestCloudUrl.Host {
		return "test"
	}

	return "prod"
}

func IsOkta(url string) bool {
	return slices.Contains([]string{"prod-us-gov", "infra-us-gov", "devel-us-gov"}, GetCCloudEnvFromBaseUrl(url))
}
