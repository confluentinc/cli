package sso

import (
	"net/url"
	"strings"

	testserver "github.com/confluentinc/cli/test/test-server"
)

var auth0ClientIds = map[string]string{
	"prod":  "oX2nvSKl5jvBKVgwehZfvR4K8RhsZIEs",
	"stag":  "8RxQmZEYtEDah4MTIIzl4hGGeFwdJS6w",
	"devel": "sPhOuMMVRSFFR7HfB606KLxf1RAU4SXg",
	"cpd":   "7rG4pmRbnMn5mIsEBLAP941IE1x2rNqC",
	"test":  "00000000000000000000000000000000",
}

func GetAuth0CCloudClientIdFromBaseUrl(baseUrl string) string {
	env := getCCloudEnvFromBaseUrl(baseUrl)
	return auth0ClientIds[env]
}

func getCCloudEnvFromBaseUrl(baseUrl string) string {
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
	case testserver.TestCloudUrl.Host:
		return "test"
	default:
		return "prod"
	}
}
