package config

import (
	"encoding/json"

	"github.com/hashicorp/go-version"
)

type BaseConfig struct {
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
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	version, err := version.NewVersion(s)
	if err != nil {
		return err
	}
	v.Version = version

	return nil
}

func NewBaseConfig(ver *version.Version) *BaseConfig {
	return &BaseConfig{
		Filename: "",
		Ver:      Version{ver},
	}
}

func (c *BaseConfig) Version() *version.Version {
	return c.Ver.Version
}
