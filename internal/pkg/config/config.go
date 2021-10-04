package config

import (
	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
)

type Config interface {
	Save() error
	Load() error
	Validate() error
	Version() *version.Version
	SetParams(params *Params)
}

type Params struct {
	CLIName    string      `json:"-"`
	MetricSink metric.Sink `json:"-"`
	Logger     *log.Logger `json:"-"`
}
