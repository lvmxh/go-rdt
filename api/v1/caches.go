package v1

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	m_cache "openstackcore-rdtagent/pkg/model/cache"
)

// Cache Info
type CachesResource struct{}

// Cache Level Info
// This should merge into CachesResource
type CachesLevelResource struct {
}

func (cache CachesResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/cache").
		Doc("Show the cache information of a host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(cache.CachesGet).
		Doc("Get the cache information, summary.").
		Operation("CachesGet").
		Writes(CachesResource{}))

	ws.Route(ws.GET("/{cache-level:l[2-3]}").To(cache.CachesLevelGet).
		Doc("Get the info of a specified level cache.").
		Param(ws.PathParameter("cache-level", "cache level").DataType("string")).
		Operation("CachesLevelGet").
		Writes(CachesLevelResource{}))

	// FIXME (Shaohe): should use pattern, \d\{1,3\}
	ws.Route(ws.GET("/{cache-level:l[2-3]}/{id}").To(cache.CacheGet).
		Doc("Get the info of a specified cache.").
		Param(ws.PathParameter("cache-level", "cache level").DataType("string")).
		Param(ws.PathParameter("id", "cache id").DataType("uint")).
		Operation("CacheGet").
		Writes(CachesLevelResource{}))
	// NOTE (Shaohe): seems DataType("uint") just for check?

	container.Add(ws)
}

// GET http://localhost:8081/cache
func (cache CachesResource) CachesGet(request *restful.Request, response *restful.Response) {
	c := &m_cache.CachesSummary{}
	err := c.Get()
	// FIXME (Shaohe): We should classify the error.
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(c)
}

// GET http://localhost:8081/cache/l2
func (cache CachesResource) CachesLevelGet(request *restful.Request, response *restful.Response) {
	c := &CachesResource{}

	level := request.PathParameter("cache-level")
	log.Println("Request Level" + strings.TrimLeft(level, "l"))

	response.WriteEntity(c)
}

// GET http://localhost:8081/cache/l2
func (cache CachesResource) CacheGet(request *restful.Request, response *restful.Response) {
	c := &CachesResource{}

	level := request.PathParameter("cache-level")

	// FIXME (Shaohe): should use pattern, \d\{1,3\}
	id, err := strconv.Atoi(request.PathParameter("id"))
	if err != nil {
		err := fmt.Errorf("Please input the right id, it shoudl be digital", id)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	log.Printf("Request Level%s, id: %d\n", strings.TrimLeft(level, "l"), id)

	response.WriteEntity(c)
}
