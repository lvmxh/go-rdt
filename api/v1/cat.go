package v1

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/log"
	cgl_cat "openstackcore-rdtagent/cgolib/cat"
	"strconv"
)

type COS struct {
	Mask uint64
}

// COSs on same socket
type COSs struct {
	CosNum uint32
	Coss   []*COS
}

type HostCOS []COSs

type CosResource struct {
}

func (c CosResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/cos").
		Doc("Cos operation of the host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(c.CacheCosGet).
		Doc("Get all cos of the host").
		Operation("CacheCosGet").
		Writes([]COS{}))

	ws.Route(ws.GET("/{socket-id}").To(c.CacheCosSocketIdGet).
		Doc("Get all cos of the specified socket id").
		Param(ws.PathParameter("socket-id", "cpu socket id").DataType("uint")).
		Operation("CacheCosSocketIdGet").
		Writes([]COS{}))

	ws.Route(ws.GET("/{socket-id}/{cos-id}").To(c.CacheCosSocketIdCosIdGet).
		Doc("Get all cos of the specified socket").
		Param(ws.PathParameter("socket-id", "cpu socket id").DataType("unit")).
		Param(ws.PathParameter("cos-id", "cos id").DataType("uint")).
		Operation("CacheCosSocketIdCosIdGet").
		Writes(COS{}))

	ws.Route(ws.PUT("/{socket-id}/{cos-id}").To(c.CacheCosSocketIdCosIdPut).
		Doc("Get all cos of the specified socket").
		Param(ws.PathParameter("socket-id", "cpu socket id").DataType("unit")).
		Param(ws.PathParameter("cos-id", "cos id").DataType("uint")).
		Operation("CacheCosSocketIdCosIdPut").
		Reads(COS{}))

	container.Add(ws)
}

func (c CosResource) CacheCosGet(request *restful.Request, response *restful.Response) {
	response.WriteEntity(cgl_cat.GetCOS())
}

func (c CosResource) CacheCosSocketIdGet(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request: %s", request.PathParameter("socket-id"))
	ui, _ := strconv.ParseInt(request.PathParameter("socket-id"), 10, 32)
	cos := cgl_cat.GetCOSBySocketId(uint16(ui))
	response.WriteEntity(cos)
}

func (c CosResource) CacheCosSocketIdCosIdGet(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request: %s", request.PathParameter("socket-id"))
	log.Printf("Received Request: %s", request.PathParameter("cos-id"))
	si, _ := strconv.ParseInt(request.PathParameter("socket-id"), 10, 32)
	ci, _ := strconv.ParseInt(request.PathParameter("cos-id"), 10, 32)
	cos := cgl_cat.GetCOSBySocketIdCosId(uint16(si), uint16(ci))
	response.WriteEntity(cos)
}

func (c CosResource) CacheCosSocketIdCosIdPut(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request: %s", request.PathParameter("socket-id"))
	log.Printf("Received Request: %s", request.PathParameter("cos-id"))

	response.WriteEntity(0)
}
