package sso

import (
	"net/url"
	"slices"

	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

var auth0ClientIds = map[string]string{
	"devel":        "sPhOuMMVRSFFR7HfB606KLxf1RAU4SXg",
	"infra-us-gov": "0oa73yrenxpdtNXEe4h7",
	"prod-us-gov":  "0oa41ih4ms3TVVAT04h7",
	"devel-us-gov": "0oaa5al9x9Rf0rF3j1d7",
	"prod":         "oX2nvSKl5jvBKVgwehZfvR4K8RhsZIEs",
	"stag":         "8RxQmZEYtEDah4MTIIzl4hGGeFwdJS6w",
	"test":         "00000000000000000000000000000000",
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

	switch u.Host {
	case "stag.cpdev.cloud":
		return "stag"
	case "devel.cpdev.cloud":
		return "devel"
	case "confluentgov.com":
		return "prod-us-gov"
	case "infra.confluentgov-internal.com":
		return "infra-us-gov"
	case "devel.confluentgov-internal.com":
		return "devel-us-gov"
	case testserver.TestCloudUrl.Host:
		return "test"
	default:
		return "prod"
	}
}

func IsOkta(url string) bool {
	return slices.Contains([]string{"prod-us-gov", "infra-us-gov", "devel-us-gov"}, GetCCloudEnvFromBaseUrl(url))
}
