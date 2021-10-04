package config

import (
	"encoding/json"

	"github.com/hashicorp/go-version"
)

type BaseConfig struct {
	*Params  `json:"-"`
	Filename string  `json:"-"`
	Ver      Version `json:"version"`
}

// Version is a wrapper type that can be marshalled into a simple version string.
// Temporary fix until https://github.com/hashicorp/go-version/pull/75 is merged.
type Version struct {
	*version.Version
}

func (v *Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v *Version) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	v.Version, err = version.NewVersion(s)
	return err
}

func NewBaseConfig(params *Params, ver *version.Version) *BaseConfig {
	if params == nil {
		params = &Params{}
	}
	return &BaseConfig{
		Params:   params,
		Filename: "",
		Ver:      Version{ver},
	}
}

func (c *BaseConfig) SetParams(params *Params) {
	c.Params = params
}

func (c *BaseConfig) Version() *version.Version {
	return c.Ver.Version
}
