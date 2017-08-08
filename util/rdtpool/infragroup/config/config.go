package config

import (
	"sync"

	"github.com/spf13/viper"
)

// NOTE InfraGroup can derive from OSGroup?
type InfraGroup struct {
	CacheWays uint     `toml:"cacheways"`
	CpuSet    string   `toml:"cpuset"`
	Tasks     []string `toml:"tasks"`
}

var configOnce sync.Once

var infragroup = &InfraGroup{}

// Concurrency-safe.
func NewConfig() *InfraGroup {
	configOnce.Do(func() {
		key := "InfraGroup"
		if !viper.IsSet(key) {
			infragroup = nil
			return
		}
		viper.UnmarshalKey(key, infragroup)
	})
	return infragroup
}
