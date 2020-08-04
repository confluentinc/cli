package mock

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
)

func NewDummyAnalyticsMock() *AnalyticsClient {
	return &AnalyticsClient{
		SetStartTimeFunc:         func() {},
		TrackCommandFunc:         func(cmd *cobra.Command, args []string) {},
		SendCommandAnalyticsFunc: func(cmd *cobra.Command, args []string, cmdExecutionError error) error { return nil },
		SetCommandTypeFunc:       func(commandType analytics.CommandType) {},
		SessionTimedOutFunc:      func() error { return nil },
		CloseFunc:                func() error { return nil },
		SetSpecialPropertyFunc:   func(propertiesKey string, value interface{}) {},
	}
}

func NewPromptMock(lines ...string) *Prompt {
	i := 0

	return &Prompt{
		ReadLineFunc: func() (string, error) {
			line := lines[i]
			i++
			return line, nil
		},
		ReadLineMaskedFunc: func() (string, error) {
			line := lines[i]
			i++
			return line, nil
		},
	}
}
