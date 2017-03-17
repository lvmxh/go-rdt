package v1

import (
	"github.com/emicklei/go-restful"

	"github.com/emicklei/go-restful/log"

	cgl_cpuinfo "openstackcore-rdtagent/cgolib/cpuinfo"
)

// CPU Info

type Cacheinfo struct {
	CacheLevel string `json:"cache_level"`
	CacheSize  uint32 `json:"cache_size"`
	CacheWay   uint32 `json:"cache_way"`
}

type CpuInfo struct {
	CpuNum  uint32    `json:"cpu_num"`
	L2Cache Cacheinfo `json:"l2_cache"`
	L3Cache Cacheinfo `json:"l3_cache"`
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

type CpuinfoResource struct {
}

func MakeCpuInfo() CpuInfo {

	ci, _ := cgl_cpuinfo.GetCpuInfo()
	var c CpuInfo
	c.CpuNum = ci.Num_cores
	c.L2Cache = Cacheinfo{"l2", ci.L2.Total_size, ci.L2.Num_ways}
	c.L3Cache = Cacheinfo{"l3", ci.L3.Total_size, ci.L3.Num_ways}
	return c
}

func GetTopo(cpuinfo *cgl_cpuinfo.PqosCpuInfo) (s, c, p int) {
	s_map := make(map[uint32]int)
	c_map := make(map[uint32]int)
	p_map := make(map[uint32]int)
	for _, i := range cpuinfo.Cores {
		s_map[i.Socket] = 1
		c_map[i.L2_id] = 1
		p_map[i.Lcore] = 1
	}
	return len(s_map), len(c_map), len(p_map)
}

func MakeCpuTopo() CpuTopo {
	ci, _ := cgl_cpuinfo.GetCpuInfo()
	s, _, _ := GetTopo(ci)
	cputopo := make([]Socket, s)

	for _, cpu := range ci.Cores {
		find_core := false
		cputopo[cpu.Socket].Id = uint16(cpu.Socket)
		for i, core := range cputopo[cpu.Socket].Cores {
			if cpu.L2_id == core.Id {
				new_cpu := Cpu{Id: cpu.Lcore, Core_Id: cpu.L2_id, Socket_Id: cpu.Socket}
				cputopo[cpu.Socket].Cores[i].Cpus = append(cputopo[cpu.Socket].Cores[i].Cpus, new_cpu)
				find_core = true
				break
			}
		}
		if !find_core {
			new_core := Core{Id: cpu.L2_id, Socket_Id: cpu.Socket}
			new_cpu := Cpu{Id: cpu.Lcore, Core_Id: cpu.L2_id, Socket_Id: cpu.Socket}
			new_core.Cpus = append(new_core.Cpus, new_cpu)
			cputopo[cpu.Socket].Cores = append(cputopo[cpu.Socket].Cores, new_core)
		}
	}
	return cputopo
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
