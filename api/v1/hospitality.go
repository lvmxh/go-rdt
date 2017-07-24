package v1

import (
	_ "fmt"
	//log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/emicklei/go-restful"
	m_hospitality "openstackcore-rdtagent/model/hospitality"
)

// Hospitality Info
type HospitalityResource struct{}

func (hospitality HospitalityResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/hospitality").
		Doc("Show the hospitality information of a host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(hospitality.HospitalitysGet).
		Doc("Get the hospitality information, summary.").
		Operation("HospitalitysGet").
		Writes(HospitalityResource{}))

	container.Add(ws)
}

// GET /v1/hospitality
func (hospitality HospitalityResource) HospitalitysGet(request *restful.Request, response *restful.Response) {
	h := &m_hospitality.Hospitality{}
	err := h.Get()
	// FIXME (Shaohe): We should classify the error.
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(h)
}
