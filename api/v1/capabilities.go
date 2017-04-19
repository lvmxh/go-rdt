package v1

import (
	"github.com/emicklei/go-restful"
	"openstackcore-rdtagent/pkg/model/capabilities"
)

type CapabilitiesResource struct {
}

func (c CapabilitiesResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/capabilities").
		Doc("Show the capabilities information of the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(c.CapGet).
		Doc("Get the capacities of the cache, memory").
		Operation("CapGet"))

	container.Add(ws)
}

// GET /v1/capabilities
func (c CapabilitiesResource) CapGet(request *restful.Request, response *restful.Response) {
	caps := capabilities.Get()
	response.WriteEntity(caps)
}
