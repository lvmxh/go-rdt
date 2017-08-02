package config

import (
	"github.com/spf13/viper"
	"sync"
)

type OSGroup struct {
	CacheWays uint   `toml:"cacheways"`
	CpuSet    string `toml:"cpuset"`
}

var configOnce sync.Once

var osgroup = &OSGroup{1, "0"}

// Concurrency-safe.
func NewConfig() OSGroup {
	configOnce.Do(func() {
		viper.UnmarshalKey("OSGroup", osgroup)
	})

	return *osgroup
}
