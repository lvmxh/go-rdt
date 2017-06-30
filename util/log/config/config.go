package config

import (
	// "github.com/heirko/go-contrib/logrusHelper"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"strconv"
	"sync"
)

type Log struct {
	Path   string `toml:"path"`
	Env    string `toml:"env"`
	Level  string `toml:"level"`
	Stdout bool   `toml:"stdout"`
}

var configOnce sync.Once
var log = &Log{"var/log/rdtagent.log", "dev", "debug", true}

// Concurrency-safe.
func NewConfig() Log {
	configOnce.Do(func() {
		// FIXME (Shaohe), we are planing to use logrusHelper. Seems we still
		// need missing some initialization for logrus. But it repors error as
		// follow:
		// # github.com/heirko/go-contrib/logrusHelper
		// undefined: logrus_mate.LoggerConfig
		// var c = logrusHelper.UnmarshalConfiguration(viper) // Unmarshal configuration from Viper
		// logrusHelper.SetConfig(logrus.StandardLogger(), c) // for e.g. apply it to logrus default instance

		viper.UnmarshalKey("log", log)

		log_dir := pflag.Lookup("log-dir").Value.String()
		if log_dir != "" {
			log.Path = log_dir
		}

		// FIXME (Shaohe), we should get the value of logtostderr by reflect
		// or flag directly, instead of strconv.ParseBool
		tostd, _ := strconv.ParseBool(pflag.Lookup("logtostderr").Value.String())
		if tostd == true {
			log.Stdout = true
		}
	})

	return *log
}
