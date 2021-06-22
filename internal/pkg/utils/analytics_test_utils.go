package utils

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
)

func NewTestAnalyticsClient(config *v3.Config, out *[]segment.Message) analytics.Client {
	testTime := time.Date(1999, time.December, 31, 23, 59, 59, 0, time.UTC)
	mockSegmentClient := &cliMock.SegmentClient{
		EnqueueFunc: func(m segment.Message) error {
			*out = append(*out, m)
			return nil
		},
		CloseFunc: func() error { return nil },
	}
	return analytics.NewAnalyticsClient(config.CLIName, config, "1.1.1.1.1", mockSegmentClient, clockwork.NewFakeClockAt(testTime))
}

func ExecuteCommandWithAnalytics(cmd *cobra.Command, args []string, analyticsClient analytics.Client) error {
	cmd.SetArgs(args)
	analyticsClient.SetStartTime()
	err := cmd.Execute()
	if err != nil {
		return err
	}
	return analyticsClient.SendCommandAnalytics(cmd, args, err)
}
