package main

import (
	application "github.com/confluentinc/flink-sql-client/pkg/controller"
)

// This is a mock function that would be used to refresh the CCloud SSO token when starting from the CLI
func authenticated() error {
	return nil
}

func main() {
	application.StartApp("envId", "orgResourceId", "kafkaClusterId", "computePoolId", "authToken", authenticated, &application.ApplicationOptions{MOCK_STATEMENTS_OUTPUT_DEMO: true})
}
