package cpuinfo

/*
#cgo CFLAGS: -I${SRCDIR}/../_c/libs/include
#cgo CFLAGS: -pthread -Wall -Winline -g -O0
#cgo CFLAGS: -fstack-protector -fPIE
#cgo CFLAGS: -D_GNU_SOURCE -DPQOS_NO_PID_API
#cgo CFLAGS: -D_FORTIFY_SOURCE=2 -D_FILE_OFFSET_BITS=64
#cgo LDFLAGS: -L${SRCDIR}/../_c/libs/lib
#cgo LDFLAGS: -lpqos
#include <pqos.h>

typedef struct pqos_cpuinfo *ppqos_cpuinfo;

const struct pqos_cpuinfo * cgo_get_cpuinfo();
const struct pqos_cap *     cgo_get_cap();
*/
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	cgl_utils "openstackcore-rdtagent/cgolib/common"
)

/* coreinfo in pqos lib
struct pqos_coreinfo {
        unsigned lcore;
        unsigned socket;
        unsigned l3_id;
        unsigned l2_id;
};
*/

type PqosCoreInfo struct {
	Lcore  uint32
	Socket uint32
	L3_id  uint32
	L2_id  uint32
}

/* cacheinfo in pqos lib
struct pqos_cacheinfo {
        int detected;
        unsigned num_ways;
        unsigned num_sets;
        unsigned num_partitions;
        unsigned line_size;
        unsigned total_size;
        unsigned way_size;
};
*/

type PqosCacheInfo struct {
	Detected       int32
	Num_ways       uint32
	Num_sets       uint32
	Num_partitions uint32
	Line_size      uint32
	Total_size     uint32
	Way_size       uint32
}

/* The pqos_cpuinfo is used to descripe the cpu info and defined in pqos lib
struct pqos_cpuinfo {
        unsigned mem_size;
        struct pqos_cacheinfo l2;
        struct pqos_cacheinfo l3;
        unsigned num_cores;
        struct pqos_coreinfo cores[0];
};
*/

/* an example of cupinfo in memory
   {mem_size = 0,
    l2 = {detected = 1, num_ways = 8, num_sets = 512,
          num_partitions = 1, line_size = 64, total_size = 262144,
          way_size = 32768},
    l3 = {detected = 1, num_ways = 20, num_sets = 45056,
          num_partitions = 1, line_size = 64, total_size = 57671680,
          way_size = 2883584},
    num_cores = 88}
*/

type PqosCpuInfo struct {
	Mem_size  uint32
	L2        PqosCacheInfo
	L3        PqosCacheInfo
	Num_cores uint32
	Cores     []*PqosCoreInfo `slice:"Num_cores,coreinfo"`
}

func NewPqosCoreInfo(s *C.struct_pqos_coreinfo) (*PqosCoreInfo, error) {
	raw := unsafe.Pointer(s)
	r := cgl_utils.NewReader(raw, C.sizeof_struct_pqos_coreinfo)

	var rr *PqosCoreInfo = &PqosCoreInfo{}
	err := cgl_utils.NewStruct(rr, r, cmeta)
	return rr, err
}

func NewPqosCacheInfo(s *C.struct_pqos_cacheinfo) (*PqosCacheInfo, error) {
	raw := unsafe.Pointer(s)
	r := cgl_utils.NewReader(raw, C.sizeof_struct_pqos_cacheinfo)

	var rr *PqosCacheInfo = &PqosCacheInfo{}
	err := cgl_utils.NewStruct(rr, r, cmeta)
	return rr, err
}

func NewPqosCpuInfo(s *C.struct_pqos_cpuinfo) (*PqosCpuInfo, error) {
	raw := unsafe.Pointer(s)
	r := cgl_utils.NewReader(raw, C.sizeof_struct_pqos_cpuinfo)
	fmt.Println("the Top struct slice  addr is:", unsafe.Pointer(r.Addr()))

	var rr *PqosCpuInfo = &PqosCpuInfo{}
	err := cgl_utils.NewStruct(rr, r, cmeta)
	if err != nil {
		fmt.Println(err)
	}
	return rr, err
}

type Cacheinfo struct {
	detected       int
	num_ways       uint32
	num_sets       uint32
	num_partitions uint32
	line_size      uint32
	total_size     uint32
	way_size       uint32
}

func GetCpuInfo() (*PqosCpuInfo, error) {
	defer C.pqos_fini()
	s := C.cgo_get_cpuinfo()
	if s == nil {
		// FIXME, we had better to get the libqpos error message,
		// and report it to User.
		err := fmt.Errorf("Error initializing cpuinfo. Could not get cpuinfo.")
		return nil, err
	}
	cpuinfo, err := NewPqosCpuInfo(s)
	return cpuinfo, err
}

/* pqos_mon_event generaged by gofmts:
go tool cgo -godefs pqos.go > generaged_pqos.go
cat pqos.go
#include <pqos.h>

#cgo LDFLAGS: -L. -lgb
import "C"

const Pqos_Version = C.PQOS_VERSION
const Pqos_Max_L3ca_Cos = C.PQOS_MAX_L3CA_COS

type PqosCapL3Ca C.struct_pqos_cap_l3ca
type PqosCapL2Ca C.struct_pqos_cap_l2ca
type PqosCapMba C.struct_pqos_cap_mba
type PqosCapMon C.struct_pqos_cap_mon

type PqosMontor C.struct_pqos_monitor
type PqosCapability C.struct_pqos_capability
*/

const Pqos_Version = 0x6a
const Pqos_Max_L3ca_Cos = 0x10

type PqosCapL3Ca struct {
	Mem_size       uint32
	Num_classes    uint32
	Num_ways       uint32
	Way_size       uint32
	Way_contention uint64
	Cdp            int32
	Cdp_on         int32
}
type PqosCapL2Ca struct {
	Mem_size       uint32
	Num_classes    uint32
	Num_ways       uint32
	Way_size       uint32
	Way_contention uint64
}
type PqosCapMba struct {
	Mem_size      uint32
	Num_classes   uint32
	Throttle_max  uint32
	Throttle_step uint32
	Is_linear     int32
}
type PqosCapMon struct {
	Mem_size   uint32
	Max_rmid   uint32
	L3_size    uint32
	Num_events uint32
	Events     []*PqosMontor `slice:"Num_events,monitor"`
}

func (c PqosMontor) Len(typ ...uint32) uint32 {
	return C.sizeof_struct_pqos_monitor
}

func (c PqosMontor) New(typ ...uint32) interface{} {
	return &PqosMontor{}
}

type PqosMontor struct {
	Type         uint32
	Max_rmid     uint32
	Scale_factor uint32
	Pid_support  uint32
}

type PqosCapability struct {
	Type    uint32
	Support int32
	U       [8]byte `union:"Type,ucapability"`
}

func IsvalidCapName(cap string) bool {
	switch cap {
	case "Num_classes",
		"Num_ways",
		"Way_size",
		"Way_contention",
		"Cdp",
		"Cdp_on",
		"Throttle_max",
		"Throttle_step",
		"Is_linear",
		"Max_rmid",
		"Num_events":
		return true
	}
	return false
}

func GetCapsMeta(r interface{}) string {
	value := reflect.ValueOf(r).Elem()
	typ := value.Type()
	meta := ""
	for i := 0; i < value.NumField(); i++ {
		sfl := value.Field(i)
		fl := typ.Field(i)
		if sfl.IsValid() {
			if fl.Type.Kind() != reflect.Slice &&
				(fl.Type.Kind() == reflect.Uint32 || fl.Type.Kind() == reflect.Uint64) {
				if IsvalidCapName(fl.Name) {
					meta = fmt.Sprintf("%s%s=%d;", meta, fl.Name, sfl.Uint())
				}
			}
		}
	}
	return meta
}

var CapEventsMap = map[uint32]string{
	C.PQOS_MON_EVENT_L3_OCCUP:  "LLC",
	C.PQOS_MON_EVENT_LMEM_BW:   "MBML",
	C.PQOS_MON_EVENT_TMEM_BW:   "MBMT",
	C.PQOS_MON_EVENT_RMEM_BW:   "MBMR",
	C.PQOS_PERF_EVENT_LLC_MISS: "LLC_MISS",
	C.PQOS_PERF_EVENT_IPC:      "IPC",
}

func (u PqosCapability) GetInfo() (string, string) {
	addr := &u.U[0]
	var t string
	var m string
	m = ""
	element_addr := *((*uintptr)(unsafe.Pointer(addr)))
	switch u.Type {
	case C.PQOS_CAP_TYPE_MON:
		cap := (*(*PqosCapMon)(unsafe.Pointer(element_addr)))
		t = "Monitor"
		m = GetCapsMeta(&cap)
		m = fmt.Sprintf("%sEvents[", m)
		for i := 0; i < int(cap.Num_events); i++ {
			m = fmt.Sprintf("%s{Type=%s,Max_rmid=%d,Scale_factor=%d,Pid_support=%d}",
				m,
				CapEventsMap[cap.Events[i].Type],
				cap.Events[i].Max_rmid,
				cap.Events[i].Scale_factor,
				cap.Events[i].Pid_support)
		}
	case C.PQOS_CAP_TYPE_L3CA:
		cap := (*(*PqosCapL3Ca)(unsafe.Pointer(element_addr)))
		t = "L3CAT"
		m = GetCapsMeta(&cap)
	case C.PQOS_CAP_TYPE_L2CA:
		cap := (*(*PqosCapL2Ca)(unsafe.Pointer(element_addr)))
		t = "L2CAT"
		m = GetCapsMeta(&cap)
	case C.PQOS_CAP_TYPE_MBA:
		cap := (*(*PqosCapMba)(unsafe.Pointer(element_addr)))
		t = "MBA"
		m = GetCapsMeta(&cap)
	}
	return t, m
}

// An empty struct for pqos_capability union
type UCapability struct {
}

// Global variable to keep capabilities size
var LibpqosUnionCapabilitySize = [...]uint32{
	C.sizeof_struct_pqos_cap_mon,
	C.sizeof_struct_pqos_cap_l3ca,
	C.sizeof_struct_pqos_cap_l2ca,
	C.sizeof_struct_pqos_cap_mba,
}

func (u UCapability) Len(typ ...uint32) uint32 {
	return LibpqosUnionCapabilitySize[typ[0]]
}

func (u UCapability) New(typ ...uint32) interface{} {
	switch typ[0] {
	case C.PQOS_CAP_TYPE_MON:
		return &PqosCapMon{}
	case C.PQOS_CAP_TYPE_L3CA:
		return &PqosCapL3Ca{}
	case C.PQOS_CAP_TYPE_L2CA:
		return &PqosCapL2Ca{}
	case C.PQOS_CAP_TYPE_MBA:
		return &PqosCapMba{}
	default:
		return nil
	}
}

type PqosCap struct {
	Mem_size     uint32
	Version      uint32
	Num_cap      uint32
	Os_enabled   uint32
	Capabilities []*PqosCapability `slice:"Num_cap,capability"`
}

func (c PqosCapability) Len(typ ...uint32) uint32 {
	return C.sizeof_struct_pqos_capability
}

func (c PqosCapability) New(typ ...uint32) interface{} {
	return &PqosCapability{}
}

func NewPqosCapability(s *C.struct_pqos_capability) (*PqosCapability, error) {
	raw := unsafe.Pointer(s)

	r := cgl_utils.NewReader(raw, C.sizeof_struct_pqos_capability)

	var rr *PqosCapability = &PqosCapability{}
	err := cgl_utils.NewStruct(rr, r, cmeta)

	return rr, err
}

func NewPqosCaps(c *C.struct_pqos_cap) (*PqosCap, error) {
	raw := unsafe.Pointer(c)

	r := cgl_utils.NewReader(raw, C.sizeof_struct_pqos_capability)
	var rr *PqosCap = &PqosCap{}
	err := cgl_utils.NewStruct(rr, r, cmeta)
	return rr, err
}

func GetCpuCaps() (*PqosCap, error) {
	defer C.pqos_fini()
	s := C.cgo_get_cap()
	if s == nil {
		// FIXME, we had better to get the libqpos error message,
		// and report it to User.
		err := fmt.Errorf(
			"Error initializing cpu capablity. Could not get cpu_cap.")
		return nil, err
	}
	caps, err := NewPqosCaps(s)
	return caps, err
}

// CMeta interface for describe C data type: pqos_coreinfo
func (c PqosCoreInfo) Len(typ ...uint32) uint32 {
	return C.sizeof_struct_pqos_coreinfo
}

func (c PqosCoreInfo) New(typ ...uint32) interface{} {
	return &PqosCoreInfo{}
}

var cmeta = map[string]cgl_utils.CMeta{
	"monitor":     &PqosMontor{},
	"capability":  &PqosCapability{},
	"ucapability": &UCapability{},
	"coreinfo":    &PqosCoreInfo{},
}
