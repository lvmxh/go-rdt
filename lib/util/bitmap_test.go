package util

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestNewBitMapsUnion(t *testing.T) {
	b, _ := NewBitMaps(88, []string{"0-7,9-12,85-87"})
	m, _ := NewBitMaps(64, []string{"6-9"})
	r := b.Or(m)
	if r.Bits[0] != 0x1FFF || r.Bits[2] != 0xe00000 {
		t.Errorf("The union should be : 0xE00000,00000000,00001FFF, now it is 0x%x,%08x,%08x",
			r.Bits[2], r.Bits[1], r.Bits[0])
	}
}

func TestNewBitMaps(t *testing.T) {
	b, _ := NewBitMaps(96, "3df00cfff00ffafff")
	wants := []int{0xffafff, 0xdf00cfff, 0x3}
	for i, v := range wants {
		if v != b.Bits[i] {
			t.Errorf("The bitmap of index %d should be: 0x%x, but it is: 0x%x",
				i, v, b.Bits[i])
		}
	}
}

func TestNewBitMapsIntersection(t *testing.T) {
	minlen := 64
	b, _ := NewBitMaps(88, []string{"0-7,9-12,32-50,85-87"})
	m, _ := NewBitMaps(minlen, []string{"6-9,32-48"})
	r := b.And(m)
	// r.Bits[0]
	len := len(r.Bits)
	if len != 2 {
		t.Error("The length of intersection of bit maps should be %d, but get %d.",
			minlen/32, len)
	}
	if r.Bits[0] != 0x2C0 || r.Bits[1] != 0x1FFFF {
		t.Errorf("The intersection should be : 0x00001FFFF,000002C0, now it is 0x%x,%08x",
			r.Bits[1], r.Bits[0])
	}
}

func TestNewBitMapsDifference(t *testing.T) {
	b, _ := NewBitMaps(88, []string{"0-7,9-12,85-87"})
	m, _ := NewBitMaps(64, []string{"6-9"})
	r := b.Xor(m)
	if r.Bits[0] != 0x1d3f || r.Bits[2] != 0xe00000 {
		t.Errorf("The difference should be : 0xE00000,00000000,00001d3f, now it is 0x%x,%08x,%08x",
			r.Bits[2], r.Bits[1], r.Bits[0])
	}
}

func TestNewBitMapsAsymmetricDiff(t *testing.T) {
	minlen := 64
	b, _ := NewBitMaps(88, []string{"0-7,9-12,85-87"})
	m, _ := NewBitMaps(minlen, []string{"6-9"})
	r := b.Axor(m)
	if r.Bits[0] != 0x1c3f || r.Bits[2] != 0xe00000 {
		t.Errorf("The asymmetric difference should be : 0xE00000,00000000,00001c3f, now it is 0x%x,%08x,%08x",
			r.Bits[2], r.Bits[1], r.Bits[0])
	}

	r = m.Axor(b)
	len := len(r.Bits)
	if len != 2 {
		t.Error("The length of intersection of bit maps should be %d, but get %d.",
			minlen/32, len)
	}
	if r.Bits[0] != 0x100 || r.Bits[1] != 0x0 {
		t.Errorf("The asymmetric difference should be : 0x0,00000100, now it is 0x%x,%08x",
			r.Bits[1], r.Bits[0])
	}
}

func TestBitMapsToString(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	b, _ := NewBitMaps(88, map_list)
	fmt.Println(b.ToString())
	str := b.ToString()
	want := "bffffc,1f000000,00000366"
	if want != str {
		t.Errorf("The value should be '%s', but get '%s'", want, str)
	}
}

func TestBitMapsToBinString(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	b, _ := NewBitMaps(88, map_list)
	str := b.ToBinString()
	want := "101111111111111111111100,00011111000000000000000000000000,00000000000000000000001101100110"
	if want != str {
		t.Errorf("The value should be '%s', but get '%s'", want, str)
	}
}

func TestBitMapsToBinStrings(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	b, _ := NewBitMaps(88, map_list)
	ss := b.ToBinStrings()

	if len(ss) != 12 {
		t.Error("The length of bit maps string sliece should be %d, but get %d.",
			12, len(ss))
	}
}

func TestBitMapsMaxConnectiveBits(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	b, _ := NewBitMaps(88, map_list)
	r := b.MaxConnectiveBits()
	want := 0x3FFFFC
	if want != r.Bits[2] {
		t.Errorf("The value should be '%d', but get '%d'", want, r.Bits[2])
	}

	map_list = []string{"1"}
	b, _ = NewBitMaps(24, map_list)
	r = b.MaxConnectiveBits()
	want = 0x2
	if want != r.Bits[0] {
		t.Errorf("The value should be '%d', but get '%d'", want, r.Bits[0])
	}
}

func TestBitMapsGetConnectiveBits(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	// 101111111111111111111100,00011111000000000000000000000000,00000000000000000000001101100110
	b, _ := NewBitMaps(88, map_list)
	r := b.GetConnectiveBits(10, 10, false)
	want := 0x3FF0
	if want != r.Bits[2] {
		t.Errorf("The value should be '0x%x', but get '0x%x'", want, r.Bits[2])
	}

	r = b.GetConnectiveBits(3, 4, false)
	want = 0xe0000
	if want != r.Bits[2] {
		t.Errorf("The value should be '0x%x', but get '0x%x'", want, r.Bits[2])
	}

	r = b.GetConnectiveBits(1, 3, false)
	want = 0x100000
	if want != r.Bits[2] {
		t.Errorf("The value should be '0x%x', but get '0x%x'", want, r.Bits[2])
	}

	r = b.GetConnectiveBits(1, 0, false)
	want = 0x800000
	if want != r.Bits[2] {
		t.Errorf("The value should be '0x%x', but get '0x%x'", want, r.Bits[2])
	}

	/********************* True **************************************/
	r = b.GetConnectiveBits(2, 3, true)
	want = 0x60
	if want != r.Bits[0] {
		t.Errorf("The value should be '%d', but get '%x'", want, r.Bits[0])
	}

	r = b.GetConnectiveBits(1, 3, true)
	want = 0x20
	if want != r.Bits[0] {
		t.Errorf("The value should be '%d', but get '%x'", want, r.Bits[0])
	}
}

func TestGenCpuResStringSimple(t *testing.T) {
	map_list := []string{"0-7"}
	s, e := GenCpuResString(map_list, 88)
	if e != nil {
		t.Errorf("Get CpuResString error: %v", e)
	}

	fmt.Println(s)
	// Output:
	// 0,0,ff
}

func TestGenCpuResString(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	map_bin := []string{"1101100110",
		"00011111000000000000000000000000", "101111111111111111111100"}

	s, e := GenCpuResString(map_list, 88)
	if e != nil {
		t.Errorf("Get CpuResString error: %v", e)
	}

	cpus := strings.Split(s, ",")
	len := len(cpus)
	if len != 3 {
		t.Error("Get Wrong cpus map string.")
	}

	for i, v := range map_bin {
		v1, _ := strconv.ParseInt(v, 2, 64)
		v2, _ := strconv.ParseInt(cpus[len-1-i], 16, 64)
		if v1 != v2 {
			t.Errorf("The bitmap of index %d should be: %s, but it is: %s",
				i, v, fmt.Sprintf("%b", v2))
		}
	}
}

func TestGenCpuResStringOutofRange(t *testing.T) {
	map_list := []string{"1-8,^3-4,^7,9", "56-88,^86,^61-65,1024"}
	_, e := GenCpuResString(map_list, 88)
	if e != nil {
		reason := fmt.Sprintf(
			"The biggest index %d is not less than the bit map length %d", 1024, 88)
		es := fmt.Sprintf("%v", e)
		if reason == es {
			t.Log(es)
		} else {
			t.Errorf("Get CpuResString error: %v", e)
		}
	} else {
		t.Errorf("Get CpuResString should failed.")
	}

}

func TestGenCpuResStringWithWrongExpression(t *testing.T) {
	map_list := []string{"abc1-8,^3-4,^7,9", "56-87,^86,^61-65"}
	_, e := GenCpuResString(map_list, 88)
	if e != nil {
		reason := "wrong expression"
		es := fmt.Sprintf("%v", e)
		if strings.Contains(es, reason) {
			t.Log(es)
		} else {
			t.Errorf("Get CpuResString error: %v", e)
		}
	}
}

func TestString2data(t *testing.T) {
	hex_datas := []uint{0xffffff0f, 0xf1, 0xff2fff}
	datas, _ := string2data("ff2fff,f1,ffffff0f")
	for i, v := range datas {
		if v != hex_datas[i] {
			t.Errorf("Parser error, the %d element should be: 0x%x, but get: 0x%x. \n",
				i, hex_datas[i], v)
		} else {
			fmt.Printf("Parser %d element, get: 0x%x. \n", i, v)
		}
	}
	fmt.Println("*****************************************")
	hex_datas = []uint{0x00ffafff, 0xdf00cfff, 0x3}
	datas, _ = string2data("3df00cfff00ffafff")
	for i, v := range datas {
		if v != hex_datas[i] {
			t.Errorf("Parser error, the %d element should be: 0x%x, but get: 0x%x.\n",
				i, hex_datas[i], v)
		} else {
			fmt.Printf("Parser %d element, get: 0x%x. \n", i, v)
		}
	}
}

func TestIsEmptyBitMap(t *testing.T) {

	cpus := "000000,00000000,00000000"
	empty := IsEmptyBitMap(cpus)
	if !empty {
		t.Errorf("Parser error, the %s element is empty bit map\n", cpus)
	}

	cpus = "000000,00000000,00000001"
	empty = IsEmptyBitMap(cpus)
	if empty {
		t.Errorf("Parser error, the %s element is not empty bit map\n", cpus)
	}
}
