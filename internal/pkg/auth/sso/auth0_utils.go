package sso

import (
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
	if baseUrl == "" {
		baseUrl = "https://confluent.cloud"
	}

	var env string
	if baseUrl == "https://confluent.cloud" {
		env = "prod"
	} else if strings.HasSuffix(baseUrl, "priv.cpdev.cloud") {
		env = "cpd"
	} else if baseUrl == "https://devel.cpdev.cloud" {
		env = "devel"
	} else if baseUrl == "https://stag.cpdev.cloud" {
		env = "stag"
	} else if baseUrl == testserver.TestCloudUrl.String() {
		env = "test"
	} else {
		return ""
	}

	return auth0ClientIds[env]
}
