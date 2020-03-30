package config

import (
	"github.com/blang/semver"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
)

type Config interface {
	Save() error
	Load() error
	Validate() error
	SetParams(params *Params)
	Version() semver.Version
}

type Params struct {
	CLIName    string      `json:"-"`
	MetricSink metric.Sink `json:"-"`
	Logger     *log.Logger `json:"-"`
}
