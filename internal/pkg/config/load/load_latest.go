package load

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

var cfgVersions = []config.Config{v1.New()}

// LoadAndMigrate loads the config file into memory using the latest config
// version, and migrates the config file to the latest version if it's not using it already.
func LoadAndMigrate(latestCfg *v1.Config) (*v1.Config, error) {
	cfg, err := loadLatestNoErr(latestCfg, len(cfgVersions)-1)
	if err != nil {
		return nil, err
	}
	// Migrate to latest config format.
	return migrateToLatest(cfg), nil
}

// loadLatestNoErr loads the config file into memory using the latest config version that doesn't result in an error.
// If the earliest config version is reached and there's still an error, an error will be returned.
func loadLatestNoErr(latestCfg *v1.Config, cfgIndex int) (config.Config, error) {
	cfg := cfgVersions[cfgIndex]
	err := cfg.Load()
	if err == nil {
		return cfg, nil
	}
	if cfgIndex == 0 {
		return nil, err
	}
	return loadLatestNoErr(latestCfg, cfgIndex-1)
}

func migrateToLatest(cfg config.Config) *v1.Config {
	switch cfg := cfg.(type) {
	case *v1.Config:
		return cfg
	default:
		return nil
	}
}
