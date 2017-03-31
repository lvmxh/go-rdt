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
	U       [8]byte
}

type PqosCap struct {
	Mem_size     uint32
	Version      uint32
	Num_cap      uint32
	Os_enabled   uint32
	Capabilities []*PqosCapability `slice:"Num_cap,capability"`
}

type CstructPqosCapablity struct {
	Name string
}

func (c CstructPqosCapablity) Len() uint32 {
	return C.sizeof_struct_pqos_capability
}

func (c CstructPqosCapablity) New() interface{} {
	var rr *PqosCapability = &PqosCapability{}
	return rr
}

func NewPqosCapability(s *C.struct_pqos_capability) (*PqosCapability, error) {
	raw := unsafe.Pointer(s)

	r := cgl_utils.NewReader(raw, C.sizeof_struct_pqos_capability)

	var rr *PqosCapability = &PqosCapability{}
	err := cgl_utils.NewStruct(rr, r, cmeta)

	/* struct pqos_capability {
		    enum pqos_cap_type type;
		    int os_support;
		     union {
		                  struct pqos_cap_mon *mon;
		                   struct pqos_cap_l3ca *l3ca;
		                    struct pqos_cap_l2ca *l2ca;
		                    struct pqos_cap_mba *mba;
		                    void *generic_ptr;
		            } u;
		    };
	    Get the addr of union
	*/

	// c := uintptr(raw) + C.sizeof_struct_pqos_capability - C.sizeof_intptr_t
	switch t := rr.Type; t {
	case 0:
		// mon
		//		addr := (*C.struct_pqos_cap_mon)(unsafe.Pointer(c))
		//		fmt.Println(*addr)
	case 1:
		// l3 cat
		//		addr := (*C.struct_pqos_cap_l3_ca)(unsafe.Pointer(c))
		//		fmt.Println(*addr)
	case 2:
		// l2 cat
		//  	addr := (*C.struct_pqos_cap_l2ca)(unsafe.Pointer(c))
		//		fmt.Println(*addr)
	default:
		// error
	}
	return rr, err
}

func NewPqosCaps(c *C.struct_pqos_cap) (*PqosCap, error) {
	raw := unsafe.Pointer(c)

	r := cgl_utils.NewReader(raw, C.sizeof_struct_pqos_capability)
	var rr *PqosCap = &PqosCap{}
	err := cgl_utils.NewStruct(rr, r, cmeta)
	if err != nil {
		return rr, err
	}
	//fixme handle array
	cap0 := uintptr(raw) + C.sizeof_struct_pqos_cap
	cap_size := uint32(C.sizeof_struct_pqos_capability)
	var i uint32 = 0
	for ; i < rr.Num_cap; i++ {
		addr := (*C.struct_pqos_capability)(unsafe.Pointer(cap0))

		cap, _ := NewPqosCapability(addr)
		rr.Capabilities = append(rr.Capabilities, cap)
		cap0 = cap0 + uintptr(cap_size)
	}
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
func (c PqosCoreInfo) Len() uint32 {
	return C.sizeof_struct_pqos_coreinfo
}

func (c PqosCoreInfo) New() interface{} {
	return &PqosCoreInfo{}
}

var cmeta = map[string]cgl_utils.CMeta{
	"capability": &CstructPqosCapablity{"capability"},
	"coreinfo":   &PqosCoreInfo{},
}
