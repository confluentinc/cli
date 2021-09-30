package load

import (
	"fmt"
	"path"

	"github.com/blang/semver"
	"github.com/mitchellh/go-homedir"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/migrations"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	cfgVersions = []config.Config{v0.New(nil), v1.New(nil), v2.New(nil), v3.New(nil)}
)

// LoadAndMigrate loads the config file into memory using the latest config
// version, and migrates the config file to the latest version if it's not using it already.
func LoadAndMigrate(latestCfg *v3.Config) (*v3.Config, error) {
	cfg, err := loadLatestNoErr(latestCfg, len(cfgVersions)-1)
	if err != nil {
		return nil, err
	}
	// Migrate to latest config format.
	return migrateToLatest(cfg)
}

// loadLatestNoErr loads the config file into memory using the latest config version that doesn't result in an error.
// If the earliest config version is reached and there's still an error, an error will be returned.
func loadLatestNoErr(latestCfg *v3.Config, cfgIndex int) (config.Config, error) {
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

func migrateToLatest(cfg config.Config) (*v3.Config, error) {
	switch cfg := cfg.(type) {
	case *v0.Config:
		cfgV1, err := migrations.MigrateV0ToV1(cfg)
		if err != nil {
			return nil, err
		}
		return migrateToLatest(cfgV1)
	case *v1.Config:
		cfgV2, err := migrations.MigrateV1ToV2(cfg)
		if err != nil {
			return nil, err
		}
		err = cfgV2.Save()
		if err != nil {
			return nil, err
		}
		return migrateToLatest(cfgV2)
	case *v2.Config:
		if err := catchV1Config(cfg); err != nil {
			return nil, err
		}
		cfgV3, err := migrations.MigrateV2ToV3(cfg)
		if err != nil {
			return nil, err
		}
		err = cfgV3.Save()
		if err != nil {
			return nil, err
		}
		return cfgV3, nil
	case *v3.Config:
		if err := catchV1Config(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	default:
		panic("unknown config type")
	}
}

// catchV1Config will return an error if a user tries to use confluent v1 with a confluent v2 config file.
func catchV1Config(cfg config.Config) error {
	// After updating the CLI to v2 (and therefore updating the config file version to v1), attempting to use confluent
	// v1 will result in the config file being mistaken for v2 or v3. This is easy to catch, since the config file
	// version will be "1.0.0" which doesn't match "2.0.0" or "3.0.0".
	if cfg.Version().Equals(semver.MustParse("1.0.0")) {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}
		configPath := path.Join(home, ".confluent", "config.json")
		return errors.NewErrorWithSuggestions(
			"config file version is incorrectly set to v1",
			fmt.Sprintf("You've updated the CLI to v2, which reset the config file version to v1. To use the old CLI, revert the config file: `mv %s.old %s`.", configPath, configPath),
		)
	}

	return nil
}
