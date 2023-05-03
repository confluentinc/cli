package controller

import (
	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
	"testing"
)

func TestGatewayClientHasOrgIdInDefaultHeader(s *testing.T) {
	rapid.Check(s, func(t *rapid.T) {
		// Given:
		envId := rapid.StringMatching("env-[a-zA-Z0-9]{3,3}").Draw(t, "Environment Id")
		orgResourceId := rapid.StringMatching("org-[a-zA-Z0-9]{3,3}").Draw(t, "Org Resource Id")
		kafkaClusterId := rapid.StringMatching("cmk-[a-zA-Z0-9]{3,3}").Draw(t, "Kafka Cluster Id")
		computePoolId := rapid.StringMatching("flcp-[a-zA-Z0-9]{3,3}").Draw(t, "Compute Pool Id")
		authToken := rapid.StringN(1, -1, -1).Draw(t, "Auth Token")
		options := &ApplicationOptions{
			FLINK_GATEWAY_URL:        "https://flink.us-west-2.aws.devel.cpdev.cloud",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"execution.runtime-mode": "streaming",
			},
		}

		// When:
		client := NewGatewayClient(envId, orgResourceId, kafkaClusterId, computePoolId, authToken, options)

		// Then:
		header := client.(*GatewayClient).client.GetConfig().DefaultHeader
		assert.Equal(t, orgResourceId, header["Org-Id"])
	})
}
