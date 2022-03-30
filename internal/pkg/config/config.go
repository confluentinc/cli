package config

import (
	"github.com/hashicorp/go-version"
)

type Config interface {
	Save() error
	Load() error
	Validate() error
	Version() *version.Version
}
