package config

import (
	"github.com/hashicorp/go-version"
)

type BaseConfig struct {
	*Params  `json:"-"`
	Filename string           `json:"-"`
	Ver      *version.Version `json:"version"`
}

func NewBaseConfig(params *Params, ver *version.Version) *BaseConfig {
	if params == nil {
		params = &Params{}
	}
	return &BaseConfig{
		Params:   params,
		Filename: "",
		Ver:      ver,
	}
}

func (c *BaseConfig) SetParams(params *Params) {
	c.Params = params
}
