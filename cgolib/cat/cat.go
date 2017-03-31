package cat

/*
#cgo CFLAGS: -I${SRCDIR}/../_c/libs/include
#cgo CFLAGS: -pthread -Wall -Winline -g -O0
#cgo CFLAGS: -fstack-protector -fPIE
#cgo CFLAGS: -D_GNU_SOURCE -DPQOS_NO_PID_API
#cgo CFLAGS: -D_FORTIFY_SOURCE=2 -D_FILE_OFFSET_BITS=64
#cgo LDFLAGS: -L${SRCDIR}/../_c/libs/lib
#cgo LDFLAGS: -lpqos
#include <pqos.h>
#include <stdlib.h>
#include "cat.h"
*/
import "C"

import (
	"fmt"
	"unsafe"

	cgl_utils "openstackcore-rdtagent/cgolib/common"
)

/* TODO move these to a better place */
type COS struct {
	Socket_id uint32
	Cos_id    uint32
	Mask      uint64
}

type COSs struct {
	CosNum uint32
	Coss   []*COS
}

type COSAssociation struct {
	Cpu_id uint32
	Cos_id uint32
}

type COSAssociations struct {
	Socket_id uint32
	CAs       []*COSAssociation
}

func NewCOS(s *C.struct_cgo_cos) (*COS, error) {
	raw := unsafe.Pointer(s)
	r := cgl_utils.NewReader(raw, C.sizeof_struct_cgo_cos)

	var rr *COS = &COS{}
	// pass nil to ignore element parsing
	err := cgl_utils.NewStruct(rr, r, nil)
	return rr, err
}

func NewCOSs(s *C.struct_cgo_cos, n C.unsigned) (*COSs, error) {
	var c *COSs = &COSs{}
	c.CosNum = uint32(n)
	raw := unsafe.Pointer(s)
	cos0 := uintptr(raw)
	cos_size := uint32(C.sizeof_struct_cgo_cos)
	for i := 0; i < int(n); i++ {
		addr := (*C.struct_cgo_cos)(unsafe.Pointer(cos0))
		newc, _ := NewCOS(addr)
		c.Coss = append(c.Coss, newc)
		cos0 = cos0 + uintptr(cos_size)
	}
	return c, nil
}

// Get all COS on the host
func GetCOS() []*COSs {
	defer C.pqos_fini()
	var num C.unsigned
	var sock_count C.unsigned
	cpuinfo := C.cgo_cat_init()

	C.pqos_cpu_get_sockets(cpuinfo, &sock_count)

	cs := make([]*COSs, 0)

	for i := 0; i < int(sock_count); i++ {
		addr := C.cgo_cat_get_cos(C.uint(i), &num)
		defer C.free(unsafe.Pointer(addr))
		cos := (*C.struct_cgo_cos)(unsafe.Pointer(addr))
		c, _ := NewCOSs(cos, num)
		cs = append(cs, c)
	}
	return cs
}

// Get all COS on the host by socket id
func GetCOSBySocketId(Sid uint16) *COSs {
	defer C.pqos_fini()
	C.cgo_cat_init()
	var num C.unsigned
	addr := C.cgo_cat_get_cos(C.uint(Sid), &num)
	defer func() {
		if addr != nil {
			C.free(unsafe.Pointer(addr))
		}
	}()
	if addr == nil {
		return nil
	}
	cos := (*C.struct_cgo_cos)(unsafe.Pointer(addr))
	c, _ := NewCOSs(cos, num)
	return c
}

// Get COS on specified socket and cos id
func GetCOSBySocketIdCosId(Sid, Cosid uint16) *COS {
	return GetCOSBySocketId(Sid).Coss[Cosid]
}

// Get COS on specified socket and cos id
func SetCOSBySocketIdCosId(Sid, Cosid uint16, mask uint64) *COS {
	defer C.pqos_fini()
	var num C.unsigned
	C.cgo_cat_init()
	addr := C.cgo_cat_set_cos(C.uint(Sid), C.uint(Cosid), &num, C.ulonglong(mask))
	defer C.free(unsafe.Pointer(addr))
	cos := (*C.struct_cgo_cos)(unsafe.Pointer(addr))
	c, _ := NewCOSs(cos, num)
	return c.Coss[Cosid]
}

func GetCOSAssociations() []*COSAssociations {
	defer C.pqos_fini()
	//var num C.unsigned
	var sock_count C.unsigned

	cas := make([]*COSAssociations, 0)
	cpuinfo := C.cgo_cat_init()
	sockets := C.pqos_cpu_get_sockets(cpuinfo, &sock_count)
	// Notes, golang doesn't support pointer arithmetic
	// we create a large go array  1 << 8 = 64 is big enough
	// to save cpu sockets
	// https://groups.google.com/forum/#!topic/golang-nuts/sV_f0VkjZTA
	sockets_go := (*[1 << 8]C.unsigned)(unsafe.Pointer(sockets))
	for i := 0; i < int(sock_count); i++ {
		var lcount C.unsigned
		var cosa *COSAssociations = &COSAssociations{}
		cosa.Socket_id = uint32(sockets_go[i])
		lcores := C.pqos_cpu_get_cores(cpuinfo, sockets_go[i], &lcount)
		lcores_go := (*[1 << 16]C.unsigned)(unsafe.Pointer(lcores))
		for j := 0; j < int(lcount); j++ {
			var cosid C.unsigned
			var ca *COSAssociation = &COSAssociation{}
			ret := C.pqos_alloc_assoc_get(C.unsigned(lcores_go[j]), &cosid)
			if ret != C.PQOS_RETVAL_OK {
				// TODO
				fmt.Println("error :")
				break
			} else {
				ca.Cpu_id = uint32(lcores_go[j])
				ca.Cos_id = uint32(cosid)
			}
			cosa.CAs = append(cosa.CAs, ca)
		}
		cas = append(cas, cosa)
	}
	return cas
}

func GetCOSAssociation(Cpuid uint32) *COSAssociation {
	defer C.pqos_fini()
	C.cgo_cat_init()
	var cosid C.unsigned
	var ca *COSAssociation = &COSAssociation{}
	ret := C.pqos_alloc_assoc_get(C.unsigned(Cpuid), &cosid)
	if ret != C.PQOS_RETVAL_OK {
		fmt.Println("Failed to get association for ", Cpuid)
		return nil
	} else {
		ca.Cpu_id = Cpuid
		ca.Cos_id = uint32(cosid)
	}
	return ca
}

func SetCOSAssociation(Cosid, Cpuid uint32) *COSAssociation {
	defer C.pqos_fini()
	C.cgo_cat_init()
	var ca *COSAssociation = &COSAssociation{}
	ret := C.pqos_alloc_assoc_set(C.unsigned(Cpuid), C.unsigned(Cosid))
	if ret != C.PQOS_RETVAL_OK {
		// TODO
		fmt.Println("error :")
		return nil
	} else {
		var cosid C.unsigned
		ret := C.pqos_alloc_assoc_get(C.unsigned(Cpuid), &cosid)
		if ret != C.PQOS_RETVAL_OK {
			// TODO
			return nil
		}
		ca.Cpu_id = Cpuid
		ca.Cos_id = uint32(cosid)
	}
	return ca
}
