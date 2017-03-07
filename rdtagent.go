package main

import (
	"openstackcore-rdtagent/api/v1"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
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
	config := restfulspec.Config{
		WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
		WebServicesURL: "http://localhost:8081",
		APIPath:        "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}

	restfulspec.RegisterOpenAPIService(config, wsContainer)

	log.Printf("start listening on localhost:8081")
	server := &http.Server{Addr: ":8081", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}


func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Rdtagent",
			Description: "Resource for managing RDT",
			Contact: &spec.ContactInfo{
				Name:  "todo",
				Email: "todo@intel.com",
				URL:   "http://todo.org",
			},
			License: &spec.License{
				Name: "MIT",
				URL:  "http://mit.org",
			},
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{spec.Tag{TagProps: spec.TagProps{
		Name:        "cpuinfo",
		Description: "Expose cpuinfo"}}}
}
