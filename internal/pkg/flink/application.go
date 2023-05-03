package flinksqlclient

import (
	"github.com/confluentinc/flink-sql-client/pkg/app"
	"github.com/confluentinc/flink-sql-client/pkg/types"
)

func StartApp(envId, resourceId, kafkaClusterId, computePoolId, authToken string, authenticated func() error, appOptions *types.ApplicationOptions) {
	app.StartApp(envId, resourceId, kafkaClusterId, computePoolId, authToken, authenticated, appOptions)
}
