package config

import (
	"sync"

	"github.com/spf13/viper"
)

type OSGroup struct {
	CacheWays uint   `toml:"cacheways"`
	CpuSet    string `toml:"cpuset"`
}

type InfraGroup struct {
	//OSGroup
	CacheWays uint     `toml:"cacheways"`
	CpuSet    string   `toml:"cpuset"`
	Tasks     []string `toml:"tasks"`
}

var infraConfigOnce sync.Once
var osConfigOnce sync.Once

var infragroup = &InfraGroup{}
var osgroup = &OSGroup{1, "0"}

func NewInfraConfig() *InfraGroup {
	infraConfigOnce.Do(func() {
		key := "InfraGroup"
		if !viper.IsSet(key) {
			infragroup = nil
			return
		}
		viper.UnmarshalKey(key, infragroup)
	})
	return infragroup
}

func NewOSConfig() OSGroup {
	osConfigOnce.Do(func() {
		viper.UnmarshalKey("OSGroup", osgroup)
	})

	return *osgroup
}
