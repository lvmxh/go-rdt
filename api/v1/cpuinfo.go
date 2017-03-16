package v1

import (
	"github.com/emicklei/go-restful"

	"github.com/emicklei/go-restful/log"

	"openstackcore-rdtagent/cgolib/cpuinfo"
)

// CPU Info

type Cacheinfo struct {
	CacheLevel	string `json:"cache_level"`
	CacheSize	uint32 `json:"cache_size"`
	CacheWay	uint32 `json:"cache_way"`
}

type CpuInfo struct {
	CpuNum	uint32 `json:"cpu_num"`
	L2Cache 	Cacheinfo `json:"l2_cache"`
	L3Cache 	Cacheinfo `json:"l3_cache"`
}

// CPU topology
type Cpu struct {
	Id	uint32 `json:"cpu_id"`
}

type Core struct {
	Id 	uint32 `json:"core_id"`
	Cpus	[]Cpu `json:"cpus"`

}

type Socket struct {
	Id 	uint16 `json:"socket_id"`
	Cores 	[]Core `json:"cores"`
}

type CpuTopo []Socket


type CpuinfoResource struct {
	// normally one would use DAO (data access object)
	info map[string]CpuTopo
}

// Fake cpu info
func MakeCpuInfo() CpuInfo {

	ci := cpuinfo.GetCpuInfo()
	var c CpuInfo
	c.CpuNum = ci.Num_cores
	c.L2Cache = Cacheinfo{"l2", c.L2Cache.Total_size, c.L2Cache.CacheWay}
	c.L3Cache = Cacheinfo{"l3", c.L3Cache.Total_size, c.L3Cache.CacheWay}
	return c
}

// Fake cpu topology
func MakeCpuTopo() CpuTopo {
	// 1 sockets, 2 cores, 2 cpu per core
	var t CpuTopo

	var s Socket
	s.Cores = make([]Core, 2)

	cpu0 := Cpu {Id: 0}
	cpu1 := Cpu {Id: 1}
	cpu2 := Cpu {Id: 2}
	cpu3 := Cpu {Id: 3}

	s.Cores[0].Cpus = make([]Cpu, 2)
	s.Cores[1].Cpus = make([]Cpu, 2)

	s.Cores[0].Cpus[0] = cpu0
	s.Cores[0].Cpus[1] = cpu1
	s.Cores[1].Cpus[0] = cpu2
	s.Cores[1].Cpus[1] = cpu3


	t = append(t, s)
	return t
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

	ws.Route(ws.GET("/capacity").To(cpuinfo.CpuinfoCapacityGet).
		Doc("Get the capacities of the cache, memory").
		Operation("CpuinfoCapacityGet").
		Writes(Cpu{}))

	container.Add(ws)
}

// GET http://localhost:8081/cpuinfo
func (cpuinfo CpuinfoResource) CpuinfoGet(request *restful.Request, response *restful.Response) {
	c := MakeCpuInfo()
	response.WriteEntity(c)
}

// GET http://localhost:8081/cpuinfo/topology
func (cpuinfo CpuinfoResource) CpuinfoTopologyGet(request *restful.Request, response *restful.Response) {
	t := MakeCpuTopo()
	response.WriteEntity(t)
}

// GET http://localhost:8081/cpuinfo/capacity

func (cpuinfo CpuinfoResource) CpuinfoCapacityGet(request *restful.Request, response *restful.Response) {

	log.Printf("In get socket id, received Request: %s", request.PathParameter("socket-id"))

	// TODO

	info := new(Cpu)
	info.Id = 0

	response.WriteEntity(info)
}

func (c CpuInfo) SwaggerDoc() map[string]string {
    return map[string]string{
	    "":	"Cpu Info",
	    "CpuNum":	"cpu number",
	    "L2Cache":	"level 2 cache information",
	    "L3Cache":	"level 3 cache information",

    }
}

func (c CpuTopo) SwaggerDoc() map[string]string {
    return map[string]string{
	    "":	"Cpu topology",

    }
}

func (s Socket) SwaggerDoc() map[string]string {
    return map[string]string{
	    "":	"Cpu Scoket",
	    "id":	"Socket id",

    }
}