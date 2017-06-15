package main

import (
	"github.com/spf13/pflag"
	"openstackcore-rdtagent/app"
	"openstackcore-rdtagent/util/conf"
	"openstackcore-rdtagent/util/flag"
	"openstackcore-rdtagent/util/options"
)

func main() {

	s := options.NewServerRunOptions()
	s.AddFlags(pflag.CommandLine)
	flag.InitFlags()
	conf.Init()
	app.RunServer(s)
}
