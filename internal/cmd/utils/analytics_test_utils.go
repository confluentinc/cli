package utils

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"time"

	"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
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

func GetPagePropertyValue(segmentMsg segment.Message, key string) (interface{}, error) {
	page, ok := segmentMsg.(segment.Page)
	if !ok {
		return "", errors.New("failed to convert segment Message to Page")
	}
	val, ok := page.Properties[key]
	if !ok {
		return "", errors.New(fmt.Sprintf("key %s does not exist in properties map", key))
	}
	return val, nil
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
