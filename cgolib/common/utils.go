package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// ReadUvarint reads an encoded unsigned integer from r and returns it as a uint64.
var overflow = errors.New("binary: varint overflows a X-bit integer")

// type Reader bytes.Reader
type Reader struct {
	bytes.Reader
	S unsafe.Pointer
}

func (r *Reader) Addr() unsafe.Pointer {
	return r.S
}

func NewReader(addr unsafe.Pointer, len int) *Reader {
	ns := reflect.SliceHeader{uintptr(addr), len, len}
	b := *(*[]byte)(unsafe.Pointer(&ns))
	r := bytes.NewReader(b)
	return &Reader{*r, addr}
}

type ByteReader interface {
	io.ByteReader
	Size() int64
	Len() int
	Addr() unsafe.Pointer
}

/* CMeta is used to describe the infomation of the data type
in C Language.

Copy C data in memory to Golang data in memory.
By CMeta, the NewStruct can translate the C data type to
Golang data type.
*/
type CMeta interface {
	Len(typ ...uint32) uint32      // The bytes length of C data type in memory need copy once
	New(typ ...uint32) interface{} // Gererate a new pointer of C data type
}

func readU8(r ByteReader) (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return b, err
	}

	return uint8(b & 0xff), nil
}

func readU16(r ByteReader) (uint16, error) {
	var x uint16
	var s uint
	for i := 0; i < 2; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		x |= uint16(b&0xff) << s
		s += 8
	}

	return x, nil
}

func readU32(r ByteReader) (uint32, error) {
	var x uint32
	var s uint
	for i := 0; i < 4; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		x |= uint32(b&0xff) << s
		s += 8
	}

	return x, nil
}

func readU64(r ByteReader) (uint64, error) {
	var x uint64
	var s uint
	for i := 0; i < 8; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		x |= uint64(b&0xff) << s
		s += 8
	}

	return x, nil
}

// A pointer can be 32 bit  or 64 bit
func readPointer(r ByteReader, n uint32) (unsafe.Pointer, error) {
	var x uint64
	var s uint
	var i uint32
	for i = 0; i < n; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		x |= uint64(b&0xff) << s
		s += 8
	}
	addr := unsafe.Pointer(uintptr(x))

	return addr, nil
}

func addptr(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func NewStruct(dest interface{}, r ByteReader, cmeta map[string]CMeta) error {

	value := reflect.ValueOf(dest).Elem()
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		sfl := value.Field(i)
		fl := typ.Field(i)
		rawSeek := fl.Tag.Get("seek")
		if len(rawSeek) > 0 {
			base := 10
			if strings.HasPrefix(rawSeek, "0x") {
				base = 16
				rawSeek = rawSeek[2:]
			}
			seek, err := strconv.ParseUint(rawSeek, base, 32)
			if err != nil {
				return err
			}

			seeker, ok := r.(io.Seeker)
			if !ok {
				return errors.New("io.Seeker interface is required")
			}
			seeker.Seek(int64(seek), 0)
		}

		if sfl.IsValid() && sfl.CanSet() {
			var err error
			switch fl.Type.Kind() {
			case reflect.Uint8:
				u8, err := readU8(r)
				if err == nil {
					sfl.SetUint(uint64(u8))
				}
			case reflect.Uint16:
				u16, err := readU16(r)
				if err == nil {
					sfl.SetUint(uint64(u16))
				}
			case reflect.Uint32:
				u32, err := readU32(r)
				if err == nil {
					sfl.SetUint(uint64(u32))
				}
			case reflect.Int32:
				u32, err := readU32(r)
				if err == nil {
					sfl.SetInt(int64(u32))
				}
			case reflect.Uint64:
				u64, err := readU64(r)
				if err == nil {
					sfl.SetUint(u64)
				}

			case reflect.Slice:
				// FIXME need improve
				slice := fl.Tag.Get("slice")
				slices := strings.Split(slice, ",")
				if fl.Tag == "" {
					err = fmt.Errorf(
						"Could not handle slice type in the middle of %s",
						typ)
				} else if slices[0] == "" || slices[0] == "-" {
					fmt.Println("Skip slice parser for", typ,
						". Let caller handle it")
				} else if len(slices) > 1 {
					var i_slice uint64 = 0
					iface := cmeta[slices[1]]
					num := value.FieldByName(slices[0]).Uint()
					len := iface.Len()
					readlen := r.Size() - int64(r.Len())
					addr := addptr(r.Addr(), uintptr(readlen))
					for ; i_slice < num; i_slice++ {
						iv := iface.New()
						nr := NewReader(addr, int(len))
						err = NewStruct(iv, nr, cmeta)
						sfl.Set(reflect.Append(sfl, reflect.ValueOf(iv)))
						addr = addptr(addr, uintptr(len))
					}
				} else {
					err = fmt.Errorf(
						"Could not know how to handle slice type: %s", typ)
				}

			case reflect.Array:
				slice := fl.Tag.Get("union")
				slices := strings.Split(slice, ",")
				if fl.Tag == "" {
					err = fmt.Errorf(
						"Could not handle Array type, if you just want ",
						"to copy binary data from C memory to Go, please ",
						"implement it. ")
				} else if slices[0] == "" || slices[0] == "-" {
					fmt.Println("Skip union parser for", typ,
						". Let caller handle it")
				} else if len(slices) > 1 {
					iface := cmeta[slices[1]]
					typ := uint32(value.FieldByName(slices[0]).Uint())
					iv := iface.New(typ)
					len := iface.Len(typ)
					size := sfl.Len()
					addr, err := readPointer(r, uint32(size))
					if err == nil {
						nr := NewReader(addr, int(len))
						err = NewStruct(iv, nr, cmeta)
						fmt.Println(iv)
					}
				} else {
					err = fmt.Errorf(
						"Could not know how to handle slice type: %s", typ)
				}
			case reflect.Struct:
				err = NewStruct(sfl.Addr().Interface(), r, cmeta)

			default:
				err = fmt.Errorf(
					"This is an unhandle type: %s, during parsing %s",
					fl.Type.Kind(), typ)
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}
