package cpuinfo

/*
#cgo CFLAGS: -I../../src/cmt-cat/lib -L../../src/cmt-cat/lib
#cgo CFLAGS: -I${SRCDIR}/cmt-cat/lib -L${SRCDIR}/cmt-cat/lib
#cgo LDFLAGS: -lpqos -fPIE
#cgo CFLAGS: -pthread -Wall -Winline -D_FILE_OFFSET_BITS=64 -g -O0
#cgo CFLAGS: -fstack-protector -D_FORTIFY_SOURCE=2 -fPIE
#cgo CFLAGS: -D_GNU_SOURCE -DPQOS_NO_PID_API
#include <cpuinfo.h>
#include <pqos.h>
#include <stdlib.h>

typedef struct pqos_cpuinfo *ppqos_cpuinfo;
const struct pqos_cpuinfo * cgo_cpuinfo_init();

*/
import "C"

import (
    "unsafe"
    "bytes"

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
    Lcore uint32
    Socket uint32
    L3_id uint32
    L2_id uint32
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
        Detected        int32
        Num_ways        uint32
        Num_sets        uint32
        Num_partitions  uint32
        Line_size       uint32
        Total_size      uint32
        Way_size        uint32
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
        Mem_size        uint32
        L2              PqosCacheInfo
        L3              PqosCacheInfo
        Num_cores       uint32
        Cores           []*PqosCoreInfo
}


func NewPqosCoreInfo(s *C.struct_pqos_coreinfo) (*PqosCoreInfo, error) {
    raw := unsafe.Pointer(s)
    data := *(*[C.sizeof_struct_pqos_coreinfo]byte)(raw)
    r := bytes.NewReader(data[:])

    var rr *PqosCoreInfo = &PqosCoreInfo{}
    err := cgl_utils.NewStruct(rr, r)
    return rr, err
}

func NewPqosCacheInfo(s *C.struct_pqos_cacheinfo) (*PqosCacheInfo, error) {
    raw := unsafe.Pointer(s)
    data := *(*[C.sizeof_struct_pqos_cacheinfo]byte)(raw)
    r := bytes.NewReader(data[:])

    var rr *PqosCacheInfo = &PqosCacheInfo{}
    err := cgl_utils.NewStruct(rr, r)
    return rr, err
}

func NewPqosCpuInfo(s *C.struct_pqos_cpuinfo) (*PqosCpuInfo, error) {
    raw := unsafe.Pointer(s)
    data := *(*[C.sizeof_struct_pqos_cpuinfo]byte)(raw)
    r := bytes.NewReader(data[:])

    var rr *PqosCpuInfo = &PqosCpuInfo{}
    err := cgl_utils.NewStruct(rr, r)
    if err != nil {
        return rr, err
    }
    // FIXME(Shaohe Feng) consider merge these code to NewStruct
    core0 := uintptr(raw) + C.sizeof_struct_pqos_cpuinfo
    core_size := uint32(C.sizeof_struct_pqos_coreinfo)
    var i uint32 = 0
    for ; i < rr.Num_cores; i++ {
        addr := (*C.struct_pqos_coreinfo)(unsafe.Pointer(core0))
        core_info, _ := NewPqosCoreInfo(addr)
        rr.Cores = append(rr.Cores, core_info)
        core0 = core0 + uintptr(core_size)
    }
    return rr, err
}

type Cacheinfo  struct{
        detected       int
        num_ways       uint32
        num_sets       uint32
        num_partitions uint32
        line_size      uint32
        total_size     uint32
        way_size       uint32
};

func GetCpuInfo() (*PqosCpuInfo, error){
    defer C.cpuinfo_fini()
    cpuinfo, err := NewPqosCpuInfo(C.cgo_cpuinfo_init())
    return cpuinfo, err
}
