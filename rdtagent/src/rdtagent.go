package main

import (
	"log"
	"net/http"
	//"strconv"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-swagger12"
    "api/v1"
)

// This example show a complete (GET,PUT,POST,DELETE) conventional example of
// a REST Resource including documentation to be served by e.g. a Swagger UI
// It is recommended to create a Resource struct (CpuinfoResource) that can encapsulate
// an object that provide domain access (a DAO)
// It has a Register method including the complete Route mapping to methods together
// with all the appropriate documentation
//

func main() {
	// to see what happens in the package, uncomment the following
	//restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

	wsContainer := restful.NewContainer()
	cpuinfo := cpuinfo.CpuinfoResource{}
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
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "/Cpuinfos/emicklei/xProjects/swagger-ui/dist"}
	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("start listening on localhost:8081")
	server := &http.Server{Addr: ":8081", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
