package mock

import (
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
)

func NewDummyAnalyticsMock() *AnalyticsClient {
	return &AnalyticsClient{
		SetCLIConfigFunc:         func(cfg *v3.Config) {},
		SetStartTimeFunc:         func() {},
		TrackCommandFunc:         func(cmd *cobra.Command, args []string) {},
		SendCommandAnalyticsFunc: func(cmd *cobra.Command, args []string, cmdExecutionError error) error { return nil },
		SetCommandTypeFunc:       func(commandType analytics.CommandType) {},
		SessionTimedOutFunc:      func() error { return nil },
		CloseFunc:                func() error { return nil },
		SetSpecialPropertyFunc:   func(propertiesKey string, value interface{}) {},
	}
}

func NewPromptMock(msg string) *Prompt {
	return &Prompt{
		ReadStringFunc: func(delim byte) (string, error) {
			return msg, nil
		},
	}
}
