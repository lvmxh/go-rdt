package cpuinfo

import (
	// "strconv"

	"github.com/emicklei/go-restful"
)

// GET http://localhost:8081/cpuinfo
//

type Cpuinfo struct {
	Id string
}

type CpuinfoResource struct {
	// normally one would use DAO (data access object)
	info map[string]Cpuinfo
}

func (cpuinfo CpuinfoResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
    // FIXME Now, here is hard code for v1, need refactor.
	ws.
		Path("/v1").
		Doc("Show the cupinfo of a host.").
        // FIXME just need to support json.
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/cpuinfo").To(cpuinfo.getCpuinfo).
		// docs
		Doc("get cpuinfo").
		Operation("getCpuinfo").
		Writes(Cpuinfo{})) // on the response

	container.Add(ws)
}

// GET http://localhost:8081/cpuinfo/1
//
func (cpuinfo CpuinfoResource) getCpuinfo(request *restful.Request, response *restful.Response) {
    res := make(map[string]Cpuinfo)
	info := new(Cpuinfo)
	info.Id = "1"
    res["socket"] = *info
	response.WriteEntity(res)
}

