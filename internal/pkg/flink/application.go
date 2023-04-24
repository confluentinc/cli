package flinksqlclient

import (
	application "github.com/confluentinc/flink-sql-client/pkg/controller"
)

func StartApp(envId, resourceId, kafkaClusterId, computePoolId, authToken string, authenticated func() error, appOptions *application.ApplicationOptions) {
	application.StartApp(envId, resourceId, kafkaClusterId, computePoolId, authToken, authenticated, appOptions)
}
