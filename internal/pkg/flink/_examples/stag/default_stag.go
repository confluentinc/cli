package main

import (
	"github.com/confluentinc/cli/internal/pkg/flink/app"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

func main() {
	dflt := "TVWUZYA6H5FDOGRP:RVMoRI2H2cGPjM449S70oIIoYeZVyODp+rnFR9ITzUuO4h68oX47izuDvNYqR3hT"
	dfltCluster := "NH6APAPIOJZ6S2GM:bUezUEuqjAcNzClGYbUHDCl3OtUynpfB+GYcecclVtiwya7ypnj4ftqHICWxp+cM"
	app.StartApp(
		"env-v6vj5n",
		"4d86a35c-ede5-4767-b82d-0b426e565d67",
		"cluster",
		"lfcp-8qndv5",
		"idp-id-123",
		func() string { return "authToken" },
		func() error { return nil },
		&types.ApplicationOptions{
			FLINK_GATEWAY_URL:        "https://flink.us-west-2.aws.stag.cpdev.cloud",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"execution.runtime-mode":         "streaming",
				"catalog":                        "integration",
				"confluent.kafka.keys":           "cluster:" + dfltCluster + ";lkc-675zpj:" + dfltCluster + ";",
				"confluent.schema_registry.keys": "default:" + dflt + ";",
			},
		})
}
