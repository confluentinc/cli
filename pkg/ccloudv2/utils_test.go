package ccloudv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCCloudURL_True(t *testing.T) {
	for _, url := range []string{
		"confluent.cloud",
		"https://confluent.cloud",
		"https://devel.cpdev.cloud/",
		"devel.cpdev.cloud",
		"stag.cpdev.cloud",
		"premium-oryx.gcp.priv.cpdev.cloud",
	} {
		isCCloud := IsCCloudURL(url, false)
		require.True(t, isCCloud, url+" should return true")
	}
}

func TestIsCCloudURL_False(t *testing.T) {
	for _, url := range []string{
		"example.com",
		"example.com:8090",
		"https://example.com",
	} {
		isCCloud := IsCCloudURL(url, false)
		require.False(t, isCCloud, url+" should return false")
	}
}

func TestGetServerUrl(t *testing.T) {
	m := map[string]string{
		"https://confluent.cloud":                  "https://api.confluent.cloud",
		"https://devel.cpdev.cloud":                "https://api.devel.cpdev.cloud",
		"https://stag.cpdev.cloud":                 "https://api.stag.cpdev.cloud",
		"https://stag.cpdev.cloud/":                "https://api.stag.cpdev.cloud",
		"https://healthy-fox.gcp.priv.cpdev.cloud": "https://healthy-fox.gcp.priv.cpdev.cloud/api",
	}

	for baseUrl, serverUrl := range m {
		assert.Equal(t, serverUrl, getServerUrl(baseUrl))
	}
}

func TestToLower(t *testing.T) {
	require.Equal(t, "sasl-ssl", ToLower("SASL_SSL"))
}

func TestToUpper(t *testing.T) {
	require.Equal(t, "SASL_SSL", ToUpper("sasl-ssl"))
}
