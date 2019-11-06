package analytics

import (
	segment "github.com/segmentio/analytics-go"

	"github.com/confluentinc/cli/internal/pkg/config"
)

type MockSegmentClient struct {
	Out *[]segment.Message
}

func (c *MockSegmentClient) Enqueue(m segment.Message) error {
*c.Out = append(*c.Out, m)
return nil
}

func (c *MockSegmentClient) Close() error {
return nil
}

func NewDummyAnalyticsClient() *Client {
	return NewAnalyticsClient(&config.Config{}, &MockSegmentClient{})
}
