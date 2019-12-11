package mock

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
)

func NewDummyAnalyticsMock() *AnalyticsClient {
	return &AnalyticsClient{
		SetStartTimeFunc: func() {},
		TrackCommandFunc: func(cmd *cobra.Command, args []string) {},
		CatchHelpCallFunc: func(cmd *cobra.Command, args []string) {},
		SendCommandFailedFunc: func(e error) error {return nil},
		SendCommandSucceededFunc: func() error {return nil},
		SetCommandTypeFunc: func(commandType analytics.CommandType) {},
		SessionTimedOutFunc: func() error {return nil},
		CloseFunc: func() error {return nil},
	}
}
