package mock

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
)

func NewDummyAnalyticsMock() *AnalyticsClient {
	return &AnalyticsClient{
		TrackCommandFunc: func(cmd *cobra.Command, args []string) {},
		FlushCommandFailedFunc: func(e error) error {return nil},
		FlushCommandSucceededFunc: func() error {return nil},
		SetCommandTypeFunc: func(commandType analytics.CommandType) {},
	}
}
