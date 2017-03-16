package v1

import (
	"github.com/emicklei/go-restful"

	"github.com/emicklei/go-restful/log"
)


type COS struct {
	Cos	uint64
}

type CosResource struct {
	size int64
}

func (l2 CosResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/cos").
		Doc("Cos operation of the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(l2.CacheCosGet).
		Doc("Get all cos of the host").
		Operation("CacheCosGet").
		Writes([]COS{}))

	ws.Route(ws.GET("/{socket-id}").To(l2.CacheCosSocketIdGet).
		Doc("Get all cos of the specified socket id").
		Param(ws.QueryParameter("socket-id", "cpu socket id").DataType("uint")).
		Operation("CacheCosSocketIdGet").
		Writes([]COS{}))

	ws.Route(ws.GET("/{socket-id}/{cos-id}").To(l2.CacheCosSocketIdCosIdGet).
		Doc("Get all cos of the specified socket").
		Param(ws.QueryParameter("socket-id", "cpu socket id").DataType("unit")).
		Param(ws.QueryParameter("cos-id", "cos id").DataType("uint")).
		Operation("CacheCosSocketIdGet").
		Writes(COS{}))

	ws.Route(ws.PUT("/{socket-id}/{cos-id}").To(l2.CacheCosSocketIdCosIdPut).
		Doc("Get all cos of the specified socket").
		Param(ws.QueryParameter("socket-id", "cpu socket id").DataType("unit")).
		Param(ws.QueryParameter("cos-id", "cos id").DataType("uint")).
		Operation("CacheCosSocketIdGet").
		Writes(COS{}))

	container.Add(ws)
}

func (c CosResource) CacheCosGet(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request")

	cos := make([]COS, 0)
	response.WriteEntity(cos)
}

func (c CosResource) CacheCosSocketIdGet(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request: %s", request.QueryParameter("process-list"))

	response.WriteEntity(0)
}

func (c CosResource) CacheCosSocketIdCosIdGet(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request: %s", request.QueryParameter("socket-id"))
	log.Printf("Received Request: %s", request.QueryParameter("cos-id"))

	response.WriteEntity(0)
}


func (c CosResource) CacheCosSocketIdCosIdPut(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request: %s", request.QueryParameter("socket-id"))
	log.Printf("Received Request: %s", request.QueryParameter("cos-id"))

	response.WriteEntity(0)
}
