package flinksqlclient

import (
	"github.com/confluentinc/cli/internal/pkg/flink/pkg/app"
	"github.com/confluentinc/cli/internal/pkg/flink/pkg/types"
)

func StartApp(envId, resourceId, kafkaClusterId, computePoolId string, authToken func() string, authenticated func() error, appOptions *types.ApplicationOptions) {
	app.StartApp(envId, resourceId, kafkaClusterId, computePoolId, authToken, authenticated, appOptions)
}
