package cat

/*
#cgo CFLAGS: -I../../src/cmt-cat/lib -L../../src/cmt-cat/lib
#cgo CFLAGS: -I${SRCDIR}/cmt-cat/lib -L${SRCDIR}/cmt-cat/lib
#cgo LDFLAGS: -lpqos -fPIE
#cgo CFLAGS: -pthread -Wall -Winline -D_FILE_OFFSET_BITS=64 -g -O0
#cgo CFLAGS: -fstack-protector -D_FORTIFY_SOURCE=2 -fPIE
#cgo CFLAGS: -D_GNU_SOURCE -DPQOS_NO_PID_API
#include <allocation.h>
#include <pqos.h>
#include <stdlib.h>
#include "cat.h"
*/
import "C"
import (
	"bytes"
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

func NewCOS(s *C.struct_cgo_cos) (*COS, error) {
	raw := unsafe.Pointer(s)
	data := *(*[C.sizeof_struct_cgo_cos]byte)(raw)
	r := bytes.NewReader(data[:])

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
	defer C.free(unsafe.Pointer(addr))
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
