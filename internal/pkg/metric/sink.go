package metric

import (
	"time"

	s "github.com/armon/go-metrics"
)

// sink is used to transmit metrics information to an external system.
type sink struct {
	Sink
}

// NewSink returns a new in-memory metrics sink.
func NewSink() *sink {
	return &sink{
		Sink: s.NewInmemSink(time.Second, 15*time.Second),
	}
}
