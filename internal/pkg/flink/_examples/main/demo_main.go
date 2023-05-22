package main

import (
	"github.com/confluentinc/cli/internal/pkg/flink/pkg/app"
	"github.com/confluentinc/cli/internal/pkg/flink/pkg/types"
)

// This is a mock function that would be used to refresh the CCloud SSO token when starting from the CLI
func authenticated() error {
	return nil
}

func main() {
	app.StartApp("envId", "orgResourceId", "kafkaClusterId", "computePoolId", func() string { return "authToken" }, authenticated, &types.ApplicationOptions{MOCK_STATEMENTS_OUTPUT_DEMO: true})
}
