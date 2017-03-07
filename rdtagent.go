package main

import (
	"openstackcore-rdtagent/api/v1"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-swagger12"
)

func main() {

	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})

	cpuinfo := v1.CpuinfoResource{}
	// Register controller to container
	cpuinfo.Register(wsContainer)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8081/apidocs and enter http://localhost:8081/apidocs.json in the api input field.
	// FIXME we should use config file for WebServicesUrl
	config := swagger.Config{
		WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: "http://localhost:8081",
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath: "/apidocs/",
		// FIXME (eliqiao): this depends on https://github.com/swagger-api/swagger-ui.git need to copy dist from it
		SwaggerFilePath: "/usr/local/share/go/src/github.com/wordnik/swagger-ui/dist"}
	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("start listening on localhost:8081")
	server := &http.Server{Addr: ":8081", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
