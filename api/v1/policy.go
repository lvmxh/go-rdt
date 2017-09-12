package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"openstackcore-rdtagent/model/policy"
)

type PolicyResource struct {
}

func (c PolicyResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/policy").
		Doc("Show the policy defined on the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(c.PolicyGet).
		Doc("Get the policy on the host").
		Operation("PolicyGet"))

	container.Add(ws)
}

// GET /v1/policy
func (c PolicyResource) PolicyGet(request *restful.Request, response *restful.Response) {
	p, err := policy.GetDefaultPlatformPolicy()
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(p)
}
