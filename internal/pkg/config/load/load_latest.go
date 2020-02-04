package load

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/migrations"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

var (
	cfgVersions = []config.Config{new(v0.Config), new(v1.Config)}
)

// LoadAndMigrate loads the config file into memory using the latest config
// version, and migrates the config file to the latest version if it's not using it already.
func LoadAndMigrate(latestCfg *v1.Config) (*v1.Config, error) {
	cfg, err := loadLatestNoErr(latestCfg, len(cfgVersions)-1)
	if err != nil {
		return nil, err
	}
	// MigrateV0ToV1
	switch cfg.(type) {
	case *v0.Config:
		cfgV0 := cfg.(*v0.Config)
		return migrations.MigrateV0ToV1(cfgV0)
	case *v1.Config:
		cfgV1 := cfg.(*v1.Config)
		return cfgV1, nil
	default:
		panic("unknown config type")
	}
}

// loadLatestNoErr loads the config file into memory using the latest config version that doesn't result in an error. 
// If the earliest config version is reached and there's still an error, an error will be returned.
func loadLatestNoErr(latestCfg *v1.Config, cfgIndex int) (config.Config, error) {
	cfg := cfgVersions[cfgIndex]
	cfg.SetParams(latestCfg.Params)
	err := cfg.Load()
	if err == nil {
		return cfg, nil
	}
	if cfgIndex == 0 {
		return nil, err
	}
	return loadLatestNoErr(latestCfg, cfgIndex-1)
}
