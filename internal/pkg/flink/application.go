package flinksqlclient

import (
	application "github.com/confluentinc/flink-sql-client/pkg/controller"
)

func StartApp(envId, resourceId, kafkaClusterId, computePoolId, authToken string, appOptions *application.ApplicationOptions) {
	application.StartApp(envId, resourceId, kafkaClusterId, computePoolId, authToken, appOptions)
}
