package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudposse/atmos/pkg/schema"
	u "github.com/cloudposse/atmos/pkg/utils"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type CacheConfig struct {
	LastChecked int64 `mapstructure:"last_checked"`
}

func GetCacheFilePath() (string, error) {
	xdgCacheHome := os.Getenv("XDG_CACHE_HOME")
	var cacheDir string
	if xdgCacheHome == "" {
		cacheDir = filepath.Join(".", ".atmos")
	} else {
		cacheDir = filepath.Join(xdgCacheHome, "atmos")
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", errors.Wrap(err, "error creating cache directory")
	}

	return filepath.Join(cacheDir, "cache.yaml"), nil
}

func withCacheFileLock(cacheFile string, fn func() error) error {
	lock := flock.New(cacheFile)
	err := lock.Lock()
	if err != nil {
		return errors.Wrap(err, "error acquiring file lock")
	}
	defer lock.Unlock()
	return fn()
}

func LoadCache() (CacheConfig, error) {
	cacheFile, err := GetCacheFilePath()
	if err != nil {
		return CacheConfig{}, err
	}

	var cfg CacheConfig
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		// No file yet, return default
		return cfg, nil
	}

	v := viper.New()
	v.SetConfigFile(cacheFile)
	if err := v.ReadInConfig(); err != nil {
		return cfg, errors.Wrap(err, "failed to read cache file")
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, errors.Wrap(err, "failed to unmarshal cache file")
	}
	return cfg, nil
}

func SaveCache2(cfg CacheConfig) error {
	cacheFile, err := GetCacheFilePath()
	if err != nil {
		return err
	}

	return withCacheFileLock(cacheFile, func() error {
		v := viper.New()
		v.Set("last_checked", cfg.LastChecked)
		if err := v.WriteConfigAs(cacheFile); err != nil {
			return errors.Wrap(err, "failed to write cache file")
		}
		return nil
	})
}

func SaveCache(cfg CacheConfig) error {
	cacheFile, err := GetCacheFilePath()
	if err != nil {
		return err
	}

	v := viper.New()
	v.Set("last_checked", cfg.LastChecked)
	if err := v.WriteConfigAs(cacheFile); err != nil {
		return errors.Wrap(err, "failed to write cache file")
	}
	return nil
}

func ShouldCheckForUpdates(lastChecked int64, frequency string) bool {
	now := time.Now().Unix()
	var interval int64
	switch frequency {
	case "minute":
		interval = 60 // 1 minute
	case "hourly":
		interval = 3600 // 1 hour
	case "daily":
		interval = 86400 // 1 day
	case "weekly":
		interval = 604800 // 1 week
	case "monthly":
		interval = 2592000 // 30 days (approximate)
	case "yearly":
		interval = 31536000 // 365 days
	default:
		u.LogWarning(schema.CliConfiguration{}, fmt.Sprintf("Unexpected frequency '%s' encountered. Defaulting to daily.", frequency))
		interval = 86400 // default to daily
	}
	return now-lastChecked >= interval
}
