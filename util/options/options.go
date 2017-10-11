package options

import (
	"github.com/spf13/pflag"
)

// ServerRunOptions are options for a running server
type ServerRunOptions struct {
	Port     string
	Addr     string
	TLSPort  string
	UnixSock string
}

// NewServerRunOptions creates a new ServerRunOptions object with default parameters
func NewServerRunOptions() *ServerRunOptions {
	return new(ServerRunOptions)
}

// AddFlags adds options from cmdline
func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {

	// TODO Cobra and viper are good friends, viper.BindPFlag can benefit from them.
	fs.StringVar(&s.Addr, "address", s.Addr,
		"listen address")

	fs.StringVar(&s.Port, "port", s.Port,
		"listen port")

	// Maybe we need to distinguish FlagSet and Flag.  So a good place to set Flag.
	pflag.String("conf-dir", "", "Directy of config file")
}
