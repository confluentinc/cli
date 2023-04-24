package main

import (
	application "github.com/confluentinc/flink-sql-client/pkg/controller"
)

func main() {
	application.StartApp(
		"env-g99xx1",
		"570ce633-4dec-4c01-8087-3417050055b0",
		"lkc-y39kdo",
		"lfcp-xyz67",
		"authToken",
		authenticated,
		&application.ApplicationOptions{
			FLINK_GATEWAY_URL:        "https://flink.us-west-2.aws.devel.cpdev.cloud",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"execution.runtime-mode": "streaming",
			},
		})
}
