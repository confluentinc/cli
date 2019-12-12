//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../../../mock/segment_client.go --pkg mock --selfpkg github.com/confluentinc/cli test_helper.go SegmentClient
package analytics

import (
	segment "github.com/segmentio/analytics-go"
)

// interface for generating mock of segment.Client
type SegmentClient interface {
	Enqueue(m segment.Message) error
	Close() error
}
