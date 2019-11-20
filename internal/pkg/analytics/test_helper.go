package analytics

import (
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"

	mockAnalytics "github.com/confluentinc/cli/internal/pkg/analytics/mock"
)

type MockSegmentClient struct {
	Out []segment.Message
}

func (c *MockSegmentClient) Enqueue(m segment.Message) error {
	c.Out = append(c.Out, m)
	return nil
}

func (c *MockSegmentClient) Close() error {
	return nil
}

func NewMockSegmentClient() *MockSegmentClient {
	 out := make([]segment.Message, 0)
	 return &MockSegmentClient{Out: out}
}

func NewDummyAnalyticsClient() Client {
	return &mockAnalytics.Client{
		TrackCommandFunc:          func(cmd *cobra.Command, args []string) {},
		FlushCommandSucceededFunc: func() error {return nil},
		FlushCommandFailedFunc:    func(e error) error {return nil},
	}
}
