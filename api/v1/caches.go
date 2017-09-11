package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"

	"openstackcore-rdtagent/lib/cache"
	m_cache "openstackcore-rdtagent/model/cache"
)

// Cache Info
type CachesResource struct{}

// Cache Level Info
// This should merge into CachesResource
type CachesLevelResource struct {
}

func getCacheLevelFromUrl(request *restful.Request) uint32 {
	var ilev uint32
	slev := strings.TrimLeft(request.PathParameter("cache-level"), "l")
	if slev == "c" {
		ilev = syscache.GetLLC()
	} else {
		// go-resfull re is a gate for badrequest. so no err here.
		l, _ := strconv.Atoi(slev)
		ilev = uint32(l)
	}
	return ilev
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

	ws.Route(ws.GET("/{cache-level:^l([2-3]|lc)$}").To(cache.CachesLevelGet).
		Doc("Get the info of a specified level cache.").
		Param(ws.PathParameter("cache-level", "cache level").DataType("string")).
		Operation("CachesLevelGet").
		Writes(CachesLevelResource{}))

	// FIXME (Shaohe): should use pattern, \d\{1,3\}
	ws.Route(ws.GET("/{cache-level:^l([2-3]|lc)$}/{id:^[0-9]{1,9}$").To(cache.CacheGet).
		Doc("Get the info of a specified cache.").
		Param(ws.PathParameter("cache-level", "cache level").DataType("string")).
		Param(ws.PathParameter("id", "cache id").DataType("uint")).
		Operation("CacheGet").
		Writes(CachesLevelResource{}))
	// NOTE (Shaohe): seems DataType("uint") just for check?

	container.Add(ws)
}

// GET /v1/cache
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

// GET /v1/cache/l[2|3|lc]
func (cache CachesResource) CachesLevelGet(request *restful.Request, response *restful.Response) {
	c := &m_cache.CacheInfos{}

	ilev := getCacheLevelFromUrl(request)
	log.Printf("Request Level %d", ilev)
	err := c.GetByLevel(ilev)
	if err != nil {
		response.WriteErrorString(err.Code, err.Error())
		return
	}
	response.WriteEntity(c)
}

// GET /v1/cache/l[2, 3]/{id}
func (cache CachesResource) CacheGet(request *restful.Request, response *restful.Response) {
	c := &m_cache.CacheInfos{}

	ilev := getCacheLevelFromUrl(request)
	// FIXME (Shaohe): should use pattern, \d\{1,3\}
	id, err := strconv.Atoi(request.PathParameter("id"))
	if err != nil {
		err := fmt.Errorf("Please input the correct id, it shoudl be digital", id)
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	log.Printf("Request Level%d, id: %d\n", ilev, id)

	err = c.GetByLevel(ilev)
	// FIXME (Shaohe): We should classify the error.
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	ci, ok := c.Caches[uint32(id)]
	if !ok {
		err := fmt.Errorf("Cache id %d for level %d is not found", id, ilev)
		response.WriteError(http.StatusNotFound, err)
		return
	}
	response.WriteEntity(ci)
}
