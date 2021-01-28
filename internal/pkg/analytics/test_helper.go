//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../../../mock/segment_client.go --pkg mock --selfpkg github.com/confluentinc/cli test_helper.go SegmentClient
package analytics

import (
	//"github.com/jonboulle/clockwork"
	segment "github.com/segmentio/analytics-go"
	//"time"
	//
	//v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	//
	//"github.com/confluentinc/cli/mock"
)

// interface for generating mock of segment.Client
type SegmentClient interface {
	Enqueue(m segment.Message) error
	Close() error
}


//func NewTestAnalyticsClient(config *v3.Config, out *[]segment.Message) analytics.Client {
//	testTime := time.Date(1999, time.December, 31, 23, 59, 59, 0, time.UTC)
//	mockSegmentClient := &mock.SegmentClient{
//		EnqueueFunc: func(m segment.Message) error {
//			*out = append(*out, m)
//			return nil
//		},
//		CloseFunc: func() error { return nil },
//	}
//	return NewAnalyticsClient(config.CLIName, config, "1.1.1.1.1", mockSegmentClient, clockwork.NewFakeClockAt(testTime))
//}

