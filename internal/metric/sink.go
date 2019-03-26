package metric

import (
	"time"

	s "github.com/armon/go-metrics"
)

// Sink is used to transmit metrics information to an external system.
type Sink struct {
	s.MetricSink
}

// NewSink returns a new in-memory metrics sink.
func NewSink() *Sink {
	return &Sink{
		MetricSink: s.NewInmemSink(time.Second, 15*time.Second),
	}
}
