package config

import (
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
)

type Config interface {
	Save() error
	Load() error
	Validate() error
	SetParams(params *Params)
}

type Params struct {
	CLIName    string      `json:"-"`
	MetricSink metric.Sink `json:"-"`
	Logger     *log.Logger `json:"-"`
}
