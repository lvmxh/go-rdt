package main

import (
	"github.com/spf13/pflag"
	"openstackcore-rdtagent/app"
	"openstackcore-rdtagent/util/bootcheck"
	"openstackcore-rdtagent/util/conf"
	"openstackcore-rdtagent/util/flag"
	"openstackcore-rdtagent/util/log"
	"openstackcore-rdtagent/util/options"
)

func main() {

	s := options.NewServerRunOptions()
	s.AddFlags(pflag.CommandLine)
	flag.InitFlags()
	conf.Init()
	log.Init()
	bootcheck.SanityCheck()
	app.RunServer(s)
}
