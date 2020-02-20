package load

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/migrations"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
)

var (
	cfgVersions = []config.Config{v0.New(nil), v1.New(nil), v2.New(nil)}
)

// LoadAndMigrate loads the config file into memory using the latest config
// version, and migrates the config file to the latest version if it's not using it already.
func LoadAndMigrate(latestCfg *v2.Config) (*v2.Config, error) {
	cfg, err := loadLatestNoErr(latestCfg, len(cfgVersions)-1)
	if err != nil {
		return nil, err
	}
	// Migrate to latest config format.
	return migrateToLatest(cfg)
}

// loadLatestNoErr loads the config file into memory using the latest config version that doesn't result in an error.
// If the earliest config version is reached and there's still an error, an error will be returned.
func loadLatestNoErr(latestCfg *v2.Config, cfgIndex int) (config.Config, error) {
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

func migrateToLatest(cfg config.Config) (*v2.Config, error) {
	switch cfg.(type) {
	case *v0.Config:
		cfgV0 := cfg.(*v0.Config)
		cfgV1, err := migrations.MigrateV0ToV1(cfgV0)
		if err != nil {
			return nil, err
		}
		return migrateToLatest(cfgV1)
	case *v1.Config:
		cfgV1 := cfg.(*v1.Config)
		cfgV2, err := migrations.MigrateV1ToV2(cfgV1)
		if err != nil {
			return nil, err
		}
		err = cfgV2.Save()
		if err != nil {
			return nil, err
		}
		return cfgV2, nil
	case *v2.Config:
		cfgV2 := cfg.(*v2.Config)
		return cfgV2, nil
	default:
		panic("unknown config type")
	}
}
