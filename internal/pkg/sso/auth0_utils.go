package sso

import "strings"

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
	} else {
		return ""
	}

	return ssoConfigs[env].ssoProviderClientID
}
