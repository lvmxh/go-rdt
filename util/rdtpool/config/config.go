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

type CachePool struct {
	Guarantee  uint `toml:"guarantee"`
	Besteffort uint `toml:"besteffort"`
	Shared     uint `toml:"shared"`
}

var infraConfigOnce sync.Once
var osConfigOnce sync.Once
var cachePoolConfigOnce sync.Once

var infragroup = &InfraGroup{}
var osgroup = &OSGroup{1, "0"}
var cachepool = &CachePool{0, 0, 0}

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

func NewOSConfig() *OSGroup {
	osConfigOnce.Do(func() {
		viper.UnmarshalKey("OSGroup", osgroup)
	})
	return osgroup
}

func NewCachePoolConfig() *CachePool {
	cachePoolConfigOnce.Do(func() {
		viper.UnmarshalKey("CachePool", cachepool)
	})
	return cachepool
}
