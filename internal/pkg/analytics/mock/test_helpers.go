package mock

import (
	"github.com/confluentinc/cli/internal/pkg/analytics"
	"github.com/spf13/cobra"
)

func NewDummyAnalyticsMock() *Client {
	return &Client{
		TrackCommandFunc: func(cmd *cobra.Command, args []string, sessionTimedOut bool) error {return nil},
		FlushCommandFailedFunc: func(e error) error {return nil},
		FlushCommandSucceededFunc: func() error {return nil},
		SetCommandTypeFunc: func(commandType analytics.CommandType) {},
	}
}
