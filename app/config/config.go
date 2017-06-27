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

type Database struct {
	Backend   string `toml:"backend"`
	Transport string `toml:"transport"`
	DBName    string `toml:"dbname"`
}

type Config struct {
	Def *Default
	Db  *Database
}

var configOnce sync.Once
var def = &Default{"localhost", 8081}
var db = &Database{}
var config = &Config{def, db}

// Concurrency-safe.
func NewConfig() Config {
	configOnce.Do(func() {
		viper.BindPFlag("address", pflag.Lookup("address"))
		viper.BindPFlag("port", pflag.Lookup("port"))
		viper.BindPFlag("backend", pflag.Lookup("backend"))
		viper.BindPFlag("transport", pflag.Lookup("transport"))
		viper.BindPFlag("dbname", pflag.Lookup("dbname"))
		viper.UnmarshalKey("default", config.Def)
		viper.UnmarshalKey("database", config.Db)
	})
	return *config
}
