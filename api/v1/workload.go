package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"openstackcore-rdtagent/db"
	"openstackcore-rdtagent/model/workload"
)

type WorkLoadResource struct {
	Db db.DB
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

	ws.Route(ws.POST("/").To(w.WorkLoadNew).
		Doc("Create new work load").
		Operation("WorkLoadNew"))

	ws.Route(ws.GET("/{id:[0-9]*}").To(w.WorkLoadGetById).
		Doc("Get workload by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadGetById"))

	ws.Route(ws.DELETE("/{id:[0-9]*}").To(w.WorkLoadDeleteById).
		Doc("Delete workload by id").
		Param(ws.PathParameter("id", "id").DataType("string")).
		Operation("WorkLoadDeleteById"))

	container.Add(ws)
}

// GET /v1/workloads
func (w WorkLoadResource) WorkLoadGet(request *restful.Request, response *restful.Response) {
	ws, err := w.Db.GetAllWorkload()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteEntity(ws)
}

// GET /v1/workloads/{id}
func (w WorkLoadResource) WorkLoadGetById(request *restful.Request, response *restful.Response) {

	id := request.PathParameter("id")
	wl, err := w.Db.GetWorkloadById(id)
	if len(wl.ID) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Could not found workload")
		return
	}
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteEntity(wl)
}

// POST /v1/workloads
// body : '{"core_ids":["1","2"], "task_ids":["123","456"], "policys": ["foo"], "algorithms": ["bar"], "group": ["infra"]}'
func (w *WorkLoadResource) WorkLoadNew(request *restful.Request, response *restful.Response) {
	wl := new(workload.RDTWorkLoad)
	err := request.ReadEntity(&wl)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	err = w.Db.CreateWorkload(wl)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	err = wl.Enforce()

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, wl)
}

// DELETE /v1/workloads/{id}
func (w WorkLoadResource) WorkLoadDeleteById(request *restful.Request, response *restful.Response) {

	id := request.PathParameter("id")
	wl := workload.RDTWorkLoad{ID: id}
	// TODO (eliqiao): We need to first remove Cos from the resctrl
	err := w.Db.DeleteWorkload(&wl)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
}
