package options

import (
	"github.com/spf13/pflag"
)

// Server options
type ServerRunOptions struct {
	Port string
	Addr string
}

// NewServerRunOptions creates a new ServerRunOptions object with default parameters
func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{
		Port:	"8081",
		Addr:	"localhost",
	}
	return &s
}

// Add options from cmdline
func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {

	fs.StringVar(&s.Addr, "address", s.Addr,
	"listen address")

	fs.StringVar(&s.Port, "port", s.Port,
	"listen port")
}