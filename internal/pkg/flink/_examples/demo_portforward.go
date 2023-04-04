package main

import (
	application "github.com/confluentinc/flink-sql-client/pkg/controller"
)

func main() {
	application.StartApp(
		"env-g99xx1",
		"570ce633-4dec-4c01-8087-3417050055b0",
		"lkc-y39kdo",
		"lflinkc-12345",
		"authToken",
		&application.ApplicationOptions{
			FLINK_GATEWAY_URL:        "http://localhost:8181",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"kafka.key":              "JIM",
				"kafka.secret":           "SECRET",
				"execution.runtime-mode": "streaming",
			},
		})
}
