package main

import (
	application "github.com/confluentinc/flink-sql-client/pkg/controller"
)

func main() {
	application.StartApp("envId", "computePoolId", "authToken", &application.ApplicationOptions{MOCK_STATEMENTS_OUTPUT_DEMO: true})
}
