package v1

import (
	"net/http"
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/log"
	cgl_cat "openstackcore-rdtagent/cgolib/cat"
)

type COS struct {
	Socket_id uint32
	Cos_id    uint32
	Mask      uint64
}

type COA struct {
	Cos_id uint32
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
		Path("/v1/cache/cos").
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
		Doc("Get cos of the specified socket and cos id").
		Param(ws.PathParameter("socket-id", "cpu socket id").DataType("unit")).
		Param(ws.PathParameter("cos-id", "cos id").DataType("uint")).
		Operation("CacheCosSocketIdCosIdGet").
		Writes(COS{}))

	ws.Route(ws.PUT("/{socket-id}/{cos-id}").To(c.CacheCosSocketIdCosIdPut).
		Doc("Update mask of cos of the specified socket and cos id").
		Param(ws.PathParameter("socket-id", "cpu socket id").DataType("unit")).
		Param(ws.PathParameter("cos-id", "cos id").DataType("uint")).
		Operation("CacheCosSocketIdCosIdPut").
		Reads(COS{}))

	ws.Route(ws.DELETE("/{socket-id}/{cos-id}").To(c.CacheCosSocketIdCosIdPut).
		Doc("Reset mask to cos of the specified socket and id").
		Param(ws.PathParameter("socket-id", "cpu socket id").DataType("unit")).
		Param(ws.PathParameter("cos-id", "cos id").DataType("uint")).
		Operation("CacheCosSocketIdCosIdDelete"))

	ws.Route(ws.GET("/cpu").To(c.CacheCosCpuGet).
		Doc("Get all cpu cos association of the host").
		Operation("CacheCosCpuCpuIdGet").
		Writes(COS{}))

	ws.Route(ws.GET("/cpu/{cpu-id}").To(c.CacheCosCpuCpuIdGet).
		Doc("Get cpu cos association of the specified cpu id").
		Param(ws.PathParameter("cpu-id", "cpu socket id").DataType("unit")).
		Operation("CacheCosCpuCpuIdGet").
		Writes(COS{}))

	ws.Route(ws.PUT("/cpu/{cpu-id}").To(c.CacheCosCpuCpuIdPut).
		Doc("Update cos association for cpu id").
		Param(ws.PathParameter("cpu-id", "cpu socket id").DataType("unit")).
		Operation("CacheCosCpuCpuIdPut").
		Reads(COA{}))

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
	cos := new(COS)
	err := request.ReadEntity(&cos)

	if err == nil {
		si, err := strconv.ParseInt(request.PathParameter("socket-id"), 10, 32)
		if err != nil {
			response.WriteError(http.StatusBadRequest, err)
		}
		ci, err := strconv.ParseInt(request.PathParameter("cos-id"), 10, 32)
		if err != nil {
			response.WriteError(http.StatusBadRequest, err)
		}
		cos.Socket_id = uint32(si)
		cos.Cos_id = uint32(ci)
		r := cgl_cat.SetCOSBySocketIdCosId(uint16(si), uint16(ci), cos.Mask)
		response.WriteEntity(r)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func (c CosResource) CacheCosSocketIdCosIdDelete(request *restful.Request, response *restful.Response) {
	// TODO
	// reset cos to 0xfffff
}

func (c CosResource) CacheCosCpuGet(request *restful.Request, response *restful.Response) {
	response.WriteEntity(cgl_cat.GetCOSAssociations())
}

func (c CosResource) CacheCosCpuCpuIdGet(request *restful.Request, response *restful.Response) {
	log.Printf("Received Request: %s", request.PathParameter("cpu-id"))
	ci, _ := strconv.ParseInt(request.PathParameter("cpu-id"), 10, 32)
	response.WriteEntity(cgl_cat.GetCOSAssociation(uint32(ci)))
}

func (c CosResource) CacheCosCpuCpuIdPut(request *restful.Request, response *restful.Response) {
	log.Printf("Received Request: %s", request.PathParameter("cpu-id"))
	ci, _ := strconv.ParseInt(request.PathParameter("cpu-id"), 10, 32)
	coa := new(COA)
	err := request.ReadEntity(&coa)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
	}
	response.WriteEntity(cgl_cat.SetCOSAssociation(uint32(coa.Cos_id), uint32(ci)))
}
