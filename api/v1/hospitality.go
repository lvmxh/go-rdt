package v1

import (
	_ "fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"

	. "openstackcore-rdtagent/api/error"
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

	ws.Route(ws.POST("/").To(hospitality.HospitalityGetByRequest).
		Doc("Get the hospitality information per request.").
		Operation("HospitalityGetByRequest").
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

func (hospitality HospitalityResource) HospitalityGetByRequest(request *restful.Request, response *restful.Response) {
	hr := &m_hospitality.HospitalityRequest{}
	err := request.ReadEntity(&hr)

	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	log.Infof("Try to get hospitality score by %v", hr)
	h := &m_hospitality.HospitalityRaw{}
	e := h.GetByRequest(hr)
	if e != nil {
		err := e.(AppError)
		response.WriteErrorString(err.Code, err.Error())
		return
	}
	response.WriteEntity(h)
}
