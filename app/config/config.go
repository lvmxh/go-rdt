package config

import (
	"crypto/tls"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var ClientAuth = map[string]tls.ClientAuthType{
	"no":              tls.NoClientCert,
	"require":         tls.RequestClientCert,
	"require_any":     tls.RequireAnyClientCert,
	"challenge_given": tls.VerifyClientCertIfGiven,
	"challenge":       tls.RequireAndVerifyClientCert,
}

const (
	CAFile       = "ca.pem"
	CertFile     = "rmd-cert.pem"
	KeyFile      = "rmd-key.pem"
	ClientCAFile = "ca.pem"
)

// TODO consider create a new struct for TLSConfig
type Default struct {
	Address      string `toml:"address"`
	Port         uint   `toml:"port"`
	TLSPort      uint   `toml:"tlsport"`
	CertPath     string `toml:"certpath"`
	ClientCAPath string `toml:"certpath"`
	ClientAuth   string `toml:"clientauth"`
	UnixSock     string `toml:"unixsock"`
	PolicyPath   string `toml:"policypath"`
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
var def = &Default{
	"localhost",
	8081,
	0,
	"etc/rdtagent/cert/server",
	"etc/rdtagent/cert/client",
	"challenge",
	"",
	"etc/rdtagent/policy.yaml"}
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
