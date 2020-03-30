package config

import "github.com/blang/semver"

type BaseConfig struct {
	*Params  `json:"-"`
	Filename string          `json:"-"`
	Ver      semver.Version `json:"version"`
}

func NewBaseConfig(params *Params, ver semver.Version) *BaseConfig {
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

func (c *BaseConfig) Version() semver.Version {
	return c.Ver
}
