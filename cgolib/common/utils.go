package utils

import (
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

type CElement interface {
	Len() uint32
	Step(v reflect.Value) uint32
}

func readU8(r io.ByteReader) (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return b, err
	}

	return uint8(b & 0xff), nil
}

func readU16(r io.ByteReader) (uint16, error) {
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

func readU32(r io.ByteReader) (uint32, error) {
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

func readU64(r io.ByteReader) (uint64, error) {
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

func addptr(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func NewStruct(dest interface{}, r io.ByteReader, celem map[string]CElement) error {

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
					iface := celem["capability"]
					vstep := value.FieldByName(slices[0])
					step := iface.Step(vstep)
					len := iface.Len()
					fmt.Println("parser", len*step, "bytes")
				} else {
					err = fmt.Errorf(
						"Could not know how to handle slice type: %s", typ)
				}

			case reflect.Struct:
				err = NewStruct(sfl.Addr().Interface(), r, celem)

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
