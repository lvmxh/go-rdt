package config

import (
	"sync"

	"github.com/spf13/viper"
)

// ACL what?
type ACL struct {
	Path   string `toml:"path"`
	Filter string `toml:"filter"`
}

var once sync.Once
var acl = &ACL{"/etc/rdtagent/acl/", "url"}

// NewACLConfig create new ACL config
func NewACLConfig() *ACL {
	once.Do(func() {
		viper.UnmarshalKey("acl", acl)
	})
	return acl
}
