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

	switch env {
	case "prod":
		return "hPbGZM8G55HSaUsaaieiiAprnJaEc9rH"
	case "stag":
		return "Lk2u2MHszzpmmiJ1LetzZw3ur41nqLrw"
	case "devel":
		return "XKlqgOEo39iyonTl3Yv3IHWIXGKDP3fA"
	case "cpd":
		return "Ru1HRWIyKdu2xNOOwuEuL6n0cjtbSeQb"
	default:
		return ""
	}
}
