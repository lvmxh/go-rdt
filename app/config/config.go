package config

import (
	"crypto/tls"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ClientAuth is a string to tls clientAuthType map
var ClientAuth = map[string]tls.ClientAuthType{
	"no":              tls.NoClientCert,
	"require":         tls.RequestClientCert,
	"require_any":     tls.RequireAnyClientCert,
	"challenge_given": tls.VerifyClientCertIfGiven,
	"challenge":       tls.RequireAndVerifyClientCert,
}

const (
	// CAFile is the certificate authority file
	CAFile = "ca.pem"
	// CertFile is the certificate file
	CertFile = "rmd-cert.pem"
	// KeyFile is the rmd private key file
	KeyFile = "rmd-key.pem"
	// ClientCAFile certificate authority file of client side
	ClientCAFile = "ca.pem"
)

// Default is the configuration in default section of config file
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

// Database represents data base configuration
type Database struct {
	Backend   string `toml:"backend"`
	Transport string `toml:"transport"`
	DBName    string `toml:"dbname"`
}

// Config represent the configuration struct
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
	"etc/rdtagent/policy.yaml",
}

var db = &Database{}
var config = &Config{def, db}

// NewConfig loads configurations from config file
func NewConfig() Config {
	configOnce.Do(func() {
		viper.BindPFlag("address", pflag.Lookup("address"))
		viper.BindPFlag("port", pflag.Lookup("port"))
		viper.UnmarshalKey("default", config.Def)
		viper.UnmarshalKey("database", config.Db)
	})
	return *config
}
