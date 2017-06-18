package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sync"
)

type Default struct {
	Address string `toml:"address"`
	Port    uint   `toml:"port"`
}

var configOnce sync.Once
var def = &Default{"localhost", 8081}

func readDefault() {
	viper.UnmarshalKey("default", &def)
}

// Concurrency-safe.
func NewDefault() Default {
	configOnce.Do(func() {
		viper.BindPFlag("address", pflag.Lookup("address"))
		viper.BindPFlag("port", pflag.Lookup("port"))
		viper.UnmarshalKey("default", &def)
	})
	return *def
}
