package config

import (
	"sync"

	"github.com/spf13/viper"
)

type ACL struct {
	Path   string `toml:"path"`
	Filter string `toml:"filter"`
}

var once sync.Once
var acl = &ACL{"/etc/rdtagent/acl/", "url"}

func NewACLConfig() *ACL {
	once.Do(func() {
		viper.UnmarshalKey("acl", acl)
	})
	return acl
}
