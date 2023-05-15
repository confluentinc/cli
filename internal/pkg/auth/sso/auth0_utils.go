package sso

import (
	"net/url"
	"strings"

	testserver "github.com/confluentinc/cli/test/test-server"
)

var auth0ClientIds = map[string]string{
	"cpd":              "7rG4pmRbnMn5mIsEBLAP941IE1x2rNqC",
	"devel":            "sPhOuMMVRSFFR7HfB606KLxf1RAU4SXg",
	"fedramp-internal": "0oa7c9gkc6bHBD2OW1d7",
	"prod":             "oX2nvSKl5jvBKVgwehZfvR4K8RhsZIEs",
	"stag":             "8RxQmZEYtEDah4MTIIzl4hGGeFwdJS6w",
	"test":             "00000000000000000000000000000000",
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

	if strings.HasSuffix(u.Host, "priv.cpdev.cloud") {
		return "cpd"
	}

	switch u.Host {
	case "stag.cpdev.cloud":
		return "stag"
	case "devel.cpdev.cloud":
		return "devel"
	case "infra.confluentgov-internal.com":
		return "fedramp-internal"
	case testserver.TestCloudUrl.Host:
		return "test"
	default:
		return "prod"
	}
}

func IsOkta(url string) bool {
	return GetCCloudEnvFromBaseUrl(url) == "fedramp-internal"
}
