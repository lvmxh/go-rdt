package v1

import (
	"net/http"
	"strconv"

	"github.com/emicklei/go-restful"
	"openstackcore-rdtagent/pkg/model/workload"
)

type WorkLoadResource struct {
	WorkLoads map[string]workload.RDTWorkLoad
}

func (w WorkLoadResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/workloads").
		Doc("Show work loads").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(w.WorkLoadGet).
		Doc("Get all work loads").
		Operation("WorkLoadGet"))

	ws.Route(ws.GET("/{id:[0-9]*}").To(w.WorkLoadGetById).
		Doc("Get work load by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadGetById"))

	ws.Route(ws.POST("/").To(w.WorkLoadNew).
		Doc("Create new work load").
		Operation("WorkLoadNew"))
	container.Add(ws)
}

// GET /v1/workloads
func (w WorkLoadResource) WorkLoadGet(request *restful.Request, response *restful.Response) {
	response.WriteEntity(w)
}

// GET /v1/workloads/{id}
func (w WorkLoadResource) WorkLoadGetById(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	wl := w.WorkLoads[id]
	if len(wl.ID) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Could not found workload")
		return
	}
	response.WriteEntity(wl)
}

// POST /v1/workloads
// body : '{"core_ids":["1","2"], "task_ids":["123","456"], "policys": ["foo"], "algorithms": ["bar"]}'
func (w *WorkLoadResource) WorkLoadNew(request *restful.Request, response *restful.Response) {
	wl := new(workload.RDTWorkLoad)
	err := request.ReadEntity(&wl)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	wl.ID = strconv.Itoa(len(w.WorkLoads) + 1)

	w.WorkLoads[wl.ID] = *wl
	response.WriteHeaderAndEntity(http.StatusCreated, wl)
}
