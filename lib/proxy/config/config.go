package config

import (
	"github.com/spf13/viper"
	"sync"
)

type PAM struct {
	Service string `toml:"service"`
}

var once sync.Once
var pam = &PAM{"rmd"}

func GetPAMConfig() *PAM {
	once.Do(func() {
		viper.UnmarshalKey("pam", pam)
	})
	return pam
}
