package sso

import "strings"

var (
	auth0ClientIds = map[string]string{
		"prod":  "hPbGZM8G55HSaUsaaieiiAprnJaEc9rH",
		"stag":  "Lk2u2MHszzpmmiJ1LetzZw3ur41nqLrw",
		"devel": "XKlqgOEo39iyonTl3Yv3IHWIXGKDP3fA",
		"cpd":   "Ru1HRWIyKdu2xNOOwuEuL6n0cjtbSeQb",
	}
)

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

	return auth0ClientIds[env]
}
