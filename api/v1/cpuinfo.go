package v1

import (
	"github.com/emicklei/go-restful"

	"github.com/emicklei/go-restful/log"
)

// GET http://localhost:8081/cpuinfo

type Cpuinfo struct {
	Id   string `json:"socket_id,omitempty"`
	Cpus string `json:"cpus,omitempty"`
}

type CpuinfoResource struct {
	// normally one would use DAO (data access object)
	info map[string]Cpuinfo
}

func (cpuinfo CpuinfoResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/cpuinfo").
		Doc("Show the cup information of a host.").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(cpuinfo.getCpuinfo).
		Doc("get cpuinfo").
		Operation("getCpuinfo").
		Writes(Cpuinfo{}))

	ws.Route(ws.GET("/{socket-id}").To(cpuinfo.getSocketId).
		Doc("get cpuinfo per socket id").
		Param(ws.PathParameter("socket-id", "indentifier for a CPU socket").DataType("string")).
		Operation("getSocketId").
		Writes(Cpuinfo{}))

	container.Add(ws)
}

// GET http://localhost:8081/cpuinfo/
func (cpuinfo CpuinfoResource) getCpuinfo(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request: %s", request.PathParameter("socket_id"))

	res := make(map[string]Cpuinfo)

	info := new(Cpuinfo)
	info.Id = "1"

	res["socket"] = *info

	response.WriteEntity(res)
}

// GET http://localhost:8081/cpuinfo/{socket_id}

func (cpuinfo CpuinfoResource) getSocketId(request *restful.Request, response *restful.Response) {

	log.Printf("In get socket id, received Request: %s", request.PathParameter("socket-id"))

	info := new(Cpuinfo)
	info.Id = "1"

	response.WriteEntity(info)
}


func (cpuinfo CpuinfoResource) SwaggerDoc() map[string]string {
    return map[string]string{
        "":		"Cpuinfo doc",
        "socket_id":	"ID of physical CPU socket",
        "cpus":		"Cpu list which sits on this socket",
    }
}