package flinksqlclient

import (
	application "github.com/confluentinc/flink-sql-client/pkg/controller"
)

func StartApp(envId, computePoolId, authToken string, appOptions *application.ApplicationOptions) {
	application.StartApp(envId, computePoolId, authToken, appOptions)
}
