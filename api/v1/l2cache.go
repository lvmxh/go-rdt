package v1

import (
	"github.com/emicklei/go-restful"

	"github.com/emicklei/go-restful/log"
)


type L2CacheUsage struct {
	size int64
}

func (l2 L2CacheUsage) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/cache/l2/usage").
		Doc("Show the level 2 cache usage of specific processes.").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(l2.getCacheUsage).
		Doc("get level 2 cache usage for process list").
		Param(ws.QueryParameter("process-list", "list of process on host").DataType("string")).
		Operation("getSocketId").
		Writes(L2CacheUsage{}))

	container.Add(ws)
}

// GET http://localhost:8081/cpuinfo/
func (l2 L2CacheUsage) getCacheUsage(request *restful.Request, response *restful.Response) {

	log.Printf("Received Request: %s", request.QueryParameter("process-list"))

	res := L2CacheUsage{}
	res.size = 100

	response.WriteEntity(res)
}
