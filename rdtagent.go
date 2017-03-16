package main

import (
	"openstackcore-rdtagent/api/v1"
	"openstackcore-rdtagent/util/flag"
	"openstackcore-rdtagent/util/options"

	"fmt"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-swagger12"
	"github.com/spf13/pflag"
)

func main() {

	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})

	cpuinfo := v1.CpuinfoResource{}
	l2cacheusage := v1.L2CacheUsage{}
	cos := v1.CosResource{}
	// Register controller to container
	cpuinfo.Register(wsContainer)
	l2cacheusage.Register(wsContainer)
	cos.Register(wsContainer)

	s := options.NewServerRunOptions()
	s.AddFlags(pflag.CommandLine)

	flag.InitFlags()

	weburl := fmt.Sprintf("http://%s:%s", s.Addr, s.Port)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8081/apidocs and enter http://localhost:8081/apidocs.json in the api input field.
	config := swagger.Config{
		WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: weburl,
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath: "/apidocs/",
		// FIXME (eliqiao): this depends on https://github.com/swagger-api/swagger-ui.git need to copy dist from it
		SwaggerFilePath: "/usr/local/share/go/src/github.com/wordnik/swagger-ui/dist",
		ApiVersion:      "1.0"}
	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("start listening on %s:%s", s.Addr, s.Port)
	server := &http.Server{Addr: s.Addr + ":" + s.Port, Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
