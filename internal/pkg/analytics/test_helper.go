//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst mock/segment_client.go --pkg mock --selfpkg github.com/confluentinc/cli test_helper.go SegmentClient
package analytics

 import (
	 segment "github.com/segmentio/analytics-go"
	 "github.com/spf13/cobra"

	 mockAnalytics "github.com/confluentinc/cli/internal/pkg/analytics/mock"
 )

// interface for generating mock of segment.Client
type SegmentClient interface {
	Enqueue(m segment.Message) error
	Close() error
}

 func NewDummyAnalyticsClient() Client {
	 return &mockAnalytics.Client{
		 TrackCommandFunc:          func(cmd *cobra.Command, args []string) {},
		 FlushCommandSucceededFunc: func() error {return nil},
		 FlushCommandFailedFunc:    func(e error) error {return nil},
	 }
 }
