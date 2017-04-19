package v1

import (
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"openstackcore-rdtagent/pkg/capabilities"
)

// CPU Info
type CpuInfo struct {
	CpuNum uint32 `json:"cpu_num"`
}

// CPU topology
type Cpu struct {
	Id        uint32 `json:"cpu_id"`
	Core_Id   uint32 `json:"core_id"`
	Socket_Id uint32 `json:"socket_id"`
}

type Core struct {
	Id        uint32 `json:"core_id"`
	Socket_Id uint32 `json:"socket_id"`
	Cpus      []Cpu  `json:"cpus"`
}

type Socket struct {
	Id    uint16 `json:"socket_id"`
	Cores []Core `json:"cores"`
}

type CpuTopo []Socket

type Capability struct {
	Type string `json:"cap_type"`
	Meta string `json:"meta"`
}

type Capabilities struct {
	Num  uint32       `json:"num_cap"`
	Caps []Capability `json:"capabilities"`
}

type CpuinfoResource struct {
}

func GetCpuInfo() (*CpuInfo, error) {
	return nil, nil
}

func GetCpuTopo() (CpuTopo, error) {
	return nil, nil
}

func GetCaps() (*Capabilities, error) {
	return nil, nil
}

func (cpuinfo CpuinfoResource) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/v1/cpuinfo").
		Doc("Show the CPU information of a host").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").To(cpuinfo.CpuinfoGet).
		Doc("Get the CPU information, include level 2, level 3 cache informatione").
		Operation("CpuinfoGet").
		Writes(CpuInfo{}))
	ws.Route(ws.GET("/topology").To(cpuinfo.CpuinfoTopologyGet).
		Doc("Get the CPU topology information of sockets, core and thread in each core on the host machine").
		Operation("CpuinfoTopologyGet").
		Writes(CpuTopo{}))

	ws.Route(ws.GET("/capabilities").To(cpuinfo.CpuinfoCapGet).
		Doc("Get the capacities of the cache, memory").
		Operation("CpuinfoCapGet").
		Writes(Capabilities{}))

	container.Add(ws)
}

// GET http://localhost:8081/cpuinfo
func (cpuinfo CpuinfoResource) CpuinfoGet(request *restful.Request, response *restful.Response) {
	c, err := GetCpuInfo()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(c)
}

// GET http://localhost:8081/cpuinfo/topology
func (cpuinfo CpuinfoResource) CpuinfoTopologyGet(request *restful.Request, response *restful.Response) {
	t, err := GetCpuTopo()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(t)
}

// GET http://localhost:8081/cpuinfo/capacity

func (cpuinfo CpuinfoResource) CpuinfoCapGet(request *restful.Request, response *restful.Response) {
	capp := capabilities.Get()
	log.Println(*capp.L3Cat)

	caps, err := GetCaps()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(caps)
}

func (c CpuInfo) SwaggerDoc() map[string]string {
	return map[string]string{
		"":        "Cpu Info",
		"CpuNum":  "cpu number",
		"L2Cache": "level 2 cache information",
		"L3Cache": "level 3 cache information",
	}
}

func (c CpuTopo) SwaggerDoc() map[string]string {
	return map[string]string{
		"": "Cpu topology",
	}
}

func (s Socket) SwaggerDoc() map[string]string {
	return map[string]string{
		"":   "Cpu Scoket",
		"id": "Socket id",
	}
}
