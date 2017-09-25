package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"

	"openstackcore-rdtagent/app"
	"openstackcore-rdtagent/lib/proxy"
	"openstackcore-rdtagent/util"
	"openstackcore-rdtagent/util/bootcheck"
	"openstackcore-rdtagent/util/conf"
	"openstackcore-rdtagent/util/flag"
	"openstackcore-rdtagent/util/log"
	"openstackcore-rdtagent/util/options"
	"openstackcore-rdtagent/util/pidfile"
)

var rmduser = "rmd"

func main() {
	// use pipe pair to communicate between root and normal process
	var in, out proxy.PipePair
	s := options.NewServerRunOptions()
	s.AddFlags(pflag.CommandLine)
	flag.InitFlags()
	conf.Init()

	if os.Getuid() == 0 {
		if !util.IsUserExist(rmduser) {
			if err := util.CreateUser(rmduser); err != nil {
				fmt.Println("Failed to create %s user", rmduser)
				os.Exit(1)
			}
		}

		if err := pidfile.CreatePID(); err != nil {
			fmt.Println("Create PID file fail. Reason: " + err.Error())
			os.Exit(1)
		}

		in.Reader, out.Writer, _ = os.Pipe()
		out.Reader, in.Writer, _ = os.Pipe()

		sigc := make(chan os.Signal, 1)
		// We are not allowed to catch SIGKILL
		// https://github.com/golang/go/issues/9463
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

		child, err := util.DropRunAs(rmduser, in.Writer, in.Reader)

		if err != nil {
			fmt.Println("Failed to drop root priviledge")
			os.Exit(1)
		}

		cleanup := func() {
			pidfile.ClosePID()
			in.Reader.Close()
			out.Writer.Close()
			out.Reader.Close()
			in.Writer.Close()
		}

		go func(p *os.Process) {
			sig := <-sigc
			fmt.Printf("Received %s, shutdown RMD\n", sig.String())
			p.Kill()
			cleanup()
			os.Exit(0)
		}(child)

		// wait for child status
		go func(p *os.Process) {
			processState, _ := p.Wait()
			if !processState.Success() {
				fmt.Println("Failed to start rmd API server, check log for details")
				cleanup()
				os.Exit(1)
			}
		}(child)

		fmt.Printf("RMD server started, REST API server serving on process %d\n", child.Pid)
		proxy.RegisterAndServe(out)
	}

	// Below are executed in child process
	log.Init()
	//in.Writer
	in.Writer = os.NewFile(3, "")
	//in.Reader
	in.Reader = os.NewFile(4, "")
	err := proxy.ConnectRPCServer(in)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// should go after connect rpc server
	bootcheck.SanityCheck()
	// should tell root process we are fail or success for the santify check
	app.RunServer(s)
}
