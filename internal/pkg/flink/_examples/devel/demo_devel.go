package main

import (
	"github.com/confluentinc/flink-sql-client/pkg/app"
	"github.com/confluentinc/flink-sql-client/pkg/types"
)

func main() {
	app.StartApp(
		"env-g99xx1",
		"570ce633-4dec-4c01-8087-3417050055b0",
		"lkc-y39kdo",
		"lfcp-xyz67",
		"authToken",
		func() error { return nil },
		&types.ApplicationOptions{
			FLINK_GATEWAY_URL:        "https://flink.us-west-2.aws.devel.cpdev.cloud",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"execution.runtime-mode": "streaming",
			},
		})
}
