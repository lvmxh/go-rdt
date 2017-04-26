package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"openstackcore-rdtagent/pkg/model/policy"
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

	ws.Route(ws.PATCH("/").To(c.PolicyUpdate).
		Doc("Update the policy").
		Operation("PolicyUpdate"))
	container.Add(ws)
}

// GET /v1/policy
func (c PolicyResource) PolicyGet(request *restful.Request, response *restful.Response) {
	// TODO get host's platform
	p, err := policy.GetPolicy("broadwell")
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(p)
}

func (c PolicyResource) PolicyUpdate(request *restful.Request, response *restful.Response) {
	// TODO
	p := PolicyResource{}
	response.WriteEntity(p)
}
