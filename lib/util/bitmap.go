package util

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	. "unsafe"
)

type BitMaps struct {
	Len  int
	Bits []int
}

func NewBitMaps(l int, map_list ...[]string) (*BitMaps, error) {
	b := new(BitMaps)
	b.Len = l
	if len(map_list) > 0 {
		bits, err := genBits(map_list[0], l)
		b.Bits = bits
		return b, err
	} else {
		b.Bits = genBitMaps(l)
	}
	return b, nil
}

// Union
func (b *BitMaps) Or(m *BitMaps) *BitMaps {
	// FIXME (Shaohe) The follow code are same with and, any design pattern for it?
	maxc := len(b.Bits)
	minc := len(m.Bits)
	maxl := b.Len
	minl := m.Len
	maxb := b
	minb := m
	if maxl < minl {
		maxc, minc = minc, maxc
		maxb, minb = minb, maxb
		maxl, minl = minl, maxl
	}

	r, _ := NewBitMaps(maxl)
	copy(r.Bits, maxb.Bits)
	for i := 0; i < minc; i++ {
		r.Bits[i] = maxb.Bits[i] | minb.Bits[i]
	}

	return r
}

// Intersection
func (b *BitMaps) And(m *BitMaps) *BitMaps {
	// FIXME (Shaohe) The follow code are same with or, any design pattern for it?
	maxc := len(b.Bits)
	minc := len(m.Bits)
	maxl := b.Len
	minl := m.Len
	maxb := b
	minb := m
	if maxl < minl {
		maxc, minc = minc, maxc
		maxb, minb = minb, maxb
		maxl, minl = minl, maxl
	}

	r, _ := NewBitMaps(minl)
	for i := 0; i < minc; i++ {
		r.Bits[i] = maxb.Bits[i] & minb.Bits[i]
	}

	return r
}

// Difference
func (b *BitMaps) Xor(m *BitMaps) *BitMaps {
	// FIXME (Shaohe) The follow code are same with or, any design pattern for it?
	maxc := len(b.Bits)
	minc := len(m.Bits)
	maxl := b.Len
	minl := m.Len
	maxb := b
	minb := m
	if maxl < minl {
		maxc, minc = minc, maxc
		maxb, minb = minb, maxb
		maxl, minl = minl, maxl
	}

	r, _ := NewBitMaps(maxl)
	copy(r.Bits, maxb.Bits)
	for i := 0; i < minc; i++ {
		r.Bits[i] = maxb.Bits[i] ^ minb.Bits[i]
	}

	return r
}

// asymmetric difference
func (b *BitMaps) Axor(m *BitMaps) *BitMaps {
	// FIXME (Shaohe) The follow code are same with or, any design pattern for it?
	maxc := len(b.Bits)
	minc := len(m.Bits)
	maxl := b.Len
	minl := m.Len
	maxb := b
	minb := m
	if maxl < minl {
		maxc, minc = minc, maxc
		maxb, minb = minb, maxb
		maxl, minl = minl, maxl
	}

	var r *BitMaps
	if b.Len == maxl {
		r, _ = NewBitMaps(maxl)
		copy(r.Bits, maxb.Bits)
	} else {
		r, _ = NewBitMaps(minl)
	}

	for i := 0; i < minc; i++ {
		r.Bits[i] = (maxb.Bits[i] ^ minb.Bits[i]) & b.Bits[i]
	}

	return r
}

// To hex string
func (b *BitMaps) ToString() string {
	str := ""
	l := len(b.Bits)
	for i, v := range b.Bits {
		if i == 0 {
			// FIXME(Shaohe) Hard code 8.
			str = fmt.Sprintf("%08x", v)
		} else if i == l-1 {
			str = fmt.Sprintf("%x", v) + "," + str
		} else {
			// FIXME(Shaohe) Hard code 8.
			str = fmt.Sprintf("%08x", v) + "," + str
		}
	}
	return str
}

// To binary string
func (b *BitMaps) ToBinString() string {
	// FIXME(Shaohe) Hard code 32.
	bs32 := fmt.Sprintf("%032d", 0)
	ts := ""
	for i, v := range b.Bits {
		rb := strconv.FormatUint(uint64(v), 2)
		l := len(rb)
		end := 0
		if 32 >= l {
			end = 32 - l
		}
		if len(b.Bits) == i+1 {
			if b.Len%32 >= l {
				end = b.Len%32 - l
			}
		}
		if 0 == i {
			ts = (bs32[0:end] + rb)
		} else {
			ts = (bs32[0:end] + rb) + "," + ts
		}
	}
	return ts
}

// To binary strings
func (b *BitMaps) ToBinStrings() []string {
	ss := []string{}
	ts := strings.Replace(b.ToBinString(), ",", "", -1)
	l := len(ts)
	org_v := ts[l-1]
	org_index := l
	for i := b.Len - 1; i >= 0; i-- {
		if ts[i] != org_v {
			ss = append(ss, ts[i+1:org_index])
			org_index = i + 1
			org_v = ts[i]
		}
		if i == 0 {
			ss = append(ss, ts[0:org_index])
		}
	}
	return ss
}

// Get MaxConnectiveBits
func (b *BitMaps) MaxConnectiveBits() *BitMaps {
	ss := b.ToBinStrings()
	total_l := 0
	max_i := 0
	max_len := 0
	cur := 0
	for i, v := range ss {
		l := len(v)
		if strings.Contains(v, "1") {
			if max_len < l {
				max_len = l
				max_i = i
				cur = total_l
			}
		}
		total_l += l
	}

	// Generate the new BitMaps
	var r *BitMaps
	scope := ""
	if max_len == 0 {
		r, _ = NewBitMaps(b.Len)
		return r
	} else if len(ss[max_i]) == 1 {
		scope = fmt.Sprintf("%d", cur)
	} else {
		scope = fmt.Sprintf("%d-%d", cur, cur+len(ss[max_i])-1)
	}
	r, _ = NewBitMaps(b.Len, []string{scope})
	return r
}

// get a connective bits for BitMaps by given ways, offset, and order
func (b *BitMaps) GetConnectiveBits(ways, offset uint32, fromLow bool) *BitMaps {
	ts := strings.Replace(b.ToBinString(), ",", "", -1)
	var total uint32 = 0
	var cur uint32 = 0
	if fromLow {
		for i := uint32(b.Len) - 1 - offset; i >= 0; i-- {
			if ts[i] == "1"[0] {
				total++
				if total >= ways {
					cur = i
					break
				}
			} else {
				total = 0
			}
		}
	} else {
		// FIXME(Shaohe)  duplicated code
		for i := offset; i < uint32(b.Len); i++ {
			if ts[i] == "1"[0] {
				total++
				if total >= ways {
					cur = i
					break
				}
			} else {
				total = 0
			}
		}
	}

	scope := strconv.Itoa(b.Len - 1 - int(cur))
	if ways > 1 {
		// Low
		if fromLow {
			scope = fmt.Sprintf("%d-%d", uint32(b.Len)-1-cur-(ways-1),
				uint32(b.Len)-1-cur)
		} else {
			// High
			scope = fmt.Sprintf("%d-%d", uint32(b.Len)-1-cur,
				uint32(b.Len)-1-cur+(ways-1))
		}
	}
	if total < ways {
		r, _ := NewBitMaps(b.Len)
		return r
	}
	r, _ := NewBitMaps(b.Len, []string{scope})
	return r
}

var EmptyMapHex = []uint{0x0, 0x0, 0x0}

var BITMAP_BAD_EXPRESSION = regexp.MustCompile(`([^\^\d-,]+)|([^\d]+-.*(,|$))|` +
	`([^,]*-[^\d]+)|(\^[^\d]+)|((\,\s)?\^$)`)
var ALL_DATAS = regexp.MustCompile(`(\d+)`)

func SliceString2Int(s []string) ([]int, error) {
	// 2^32 -1 = 4294967295
	// len("4294967295") = 10
	si := make([]int, len(s), len(s))
	for i, v := range s {
		ni, err := strconv.ParseInt(v, 10, 32)
		si[i] = int(ni)
		if err != nil {
			return si, err
		}
	}
	return si, nil
}

func genBitMaps(num int, platform ...int) []int {
	p := 32
	if len(platform) > 0 {
		p = platform[0]
	}
	n := (num + p - 1) / p
	return make([]int, n, n)
}

// "2-6,^3-4,^5"
func fillBitMap(bits int, scope string, platform ...int) (int, error) {
	// "2-6"
	hyphen_span := func(scope string, platform ...int) (int, error) {
		p := 32
		if len(platform) > 0 {
			p = platform[0]
		}
		scopes := strings.SplitN(scope, "-", 2)
		low, err := strconv.Atoi(scopes[0])
		if err != nil {
			return 0, err
		}
		high, err := strconv.Atoi(scopes[1])
		if err != nil {
			return 0, err
		}
		low = low % p
		high = high % p
		// overflow
		sv := ((1 << uint(high-low+1)) - 1) << uint(low)
		return sv, nil
	}

	// "5"
	single_bit := func(bit int) int {
		// bit should less than than the platform bits
		return 1 << uint(bit)
	}

	p := 32
	if len(platform) > 0 {
		p = platform[0]
	}
	scopes := strings.Split(scope, ",")
	for i, v := range scopes {
		// negative false, positive ture
		sign := true
		var err error = nil
		sv := 0
		if strings.Contains(v, "^") {
			sign = false
			v = strings.TrimSpace(v)
			v = strings.TrimLeft(v, "^")
		}

		if strings.Contains(v, "-") {
			sv, err = hyphen_span(v)
			if err != nil {
				// change it to log
				fmt.Printf("the %d element is %s, can not be parser", i, scopes[i])
				return bits, err
			}
		} else {
			vi, err := strconv.Atoi(v)
			if err != nil {
				// change it to log
				fmt.Printf("the %d element is %s, can not be parser", i, scopes[i])
				return bits, err
			}
			vi = vi % p
			sv = single_bit(vi)
		}
		switch sign {
		case true:
			bits = bits | sv
		case false:
			bits = bits &^ sv
		}
	}
	return bits, nil
}

//{"2-8,^3-4,^7,9", "56-87,^86"}
func genBits(map_list []string, bit_len int) ([]int, error) {
	bitmaps := genBitMaps(bit_len)
	is_span := func(span string) bool {
		return strings.Contains(span, "-")
	}

	locate := func(pos int, platform ...int) int {
		p := 32
		if len(platform) > 0 {
			p = platform[0]
		}
		return pos / p
	}

	span_phypen2int := func(span string) (int, int, error) {
		scopes := strings.SplitN(span, "-", 2)
		low, err := strconv.Atoi(scopes[0])
		if err != nil {
			return 0, 0, err
		}
		high, err := strconv.Atoi(scopes[1])
		if err != nil {
			return low, 0, err
		}
		return low, high, nil
	}

	// a span maybe a cross span, need to split them into small span.
	// but we must set the max length of span(step).
	silit_span := func(span string, steps ...int) ([]string, error) {
		step := 32
		if len(steps) > 0 {
			step = steps[0]
		}
		sign := true
		var err error = nil
		v := span
		spans := []string{}
		if !is_span(span) {
			return spans, nil
		}
		if strings.Contains(span, "^") {
			sign = false
			span = strings.TrimSpace(span)
			v = strings.TrimLeft(span, "^")
		}
		low, high, err := span_phypen2int(v)
		if err != nil {
			return spans, err
		}
		if high/step == low/step {
			return append(spans, span), nil
		}
		for i := low / step; i <= high/step; i++ {
			begin := low
			end := high
			if ((i+1)*step - 1) < high {
				end = (i+1)*step - 1
			}
			if i > low/step {
				begin = i * step
			}
			s := fmt.Sprintf("%d-%d", begin, end)
			if !sign {
				s = "^" + s
			}
			spans = append(spans, s)
		}
		return spans, err
	}

	m := ALL_DATAS.FindAllString(strings.Join(map_list, ","), -1)
	si, err := SliceString2Int(m)
	if err != nil {
		return bitmaps, err
	}
	sort.Ints(si)
	if si[len(si)-1] >= bit_len {
		return bitmaps, fmt.Errorf("The biggest index %d is not less than the bit map length %d",
			si[len(si)-1], bit_len)
	}

	for _, v := range map_list {
		// FIXME, remove to before ALL_DATAS?
		m := BITMAP_BAD_EXPRESSION.FindAllString(v, -1)
		if len(m) > 0 {
			return bitmaps, fmt.Errorf("wrong expression : %s", v)
		}
		scopes := strings.Split(v, ",")
		for _, v := range scopes {
			// negative false, positive ture
			if is_span(v) {
				spans, err := silit_span(v)
				if err != nil {
					return bitmaps, err
				}
				for _, span := range spans {
					span = strings.TrimSpace(span)
					low, _, _ := span_phypen2int(strings.TrimLeft(span, "^"))
					pos := locate(low)
					bitmaps[pos], _ = fillBitMap(bitmaps[pos], span)
				}
			} else {
				i, err := strconv.Atoi(strings.TrimLeft(v, "^"))
				if err != nil {
					return bitmaps, err
				}
				pos := locate(i)
				bitmaps[pos], _ = fillBitMap(bitmaps[pos], v)
			}
		}
	}
	return bitmaps, nil
}

//{"2-8,^3-4,^7,9", "56-87,^86"}
func GenCpuResString(map_list []string, bit_len int) (string, error) {
	bitmaps, err := genBits(map_list, bit_len)
	str := ""
	if err != nil {
		return str, err
	}
	for i, v := range bitmaps {
		if i == 0 {
			str = fmt.Sprintf("%x", v)
		} else {
			str = fmt.Sprintf("%x", v) + "," + str
		}
	}
	return str, nil
}

func string2data(s string) ([]uint, error) {
	var dummy uint
	int_len := int(Sizeof(dummy))
	s = strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	// a string with comma, such as "ff2fff,f1,ffffff0f"
	if strings.Contains(s, ",") {
		ss := strings.Split(s, ",")
		var l int = len(ss)
		datas := make([]uint, l)
		for i, v := range ss {
			if len(v) > int_len {
				return datas, fmt.Errorf(
					"string lenth > %d. I'm not so smart to guest the data type.", int_len)
			}
			if ui, err := strconv.ParseUint(v, 16, 32); err == nil {
				datas[l-1-i] = uint(ui)
			} else {
				return datas, fmt.Errorf("Can not parser %s in  %s.", v, s)
			}
		}
		return datas, nil
	} else { // a string without comma, such as "3df00cfff00ffafff"
		var l int = len(s)
		n := (l - 1 + int_len) / int_len
		datas := make([]uint, n)
		for i := 0; i < n; i++ {
			start := l - (i+1)*int_len
			end := l - i*int_len
			var ns string = s[:end]
			if start > 0 {
				ns = s[start:end]
			}
			if ui, err := strconv.ParseUint(ns, 16, 32); err == nil {
				datas[i] = uint(ui)
			} else {
				return datas, fmt.Errorf("Can not parser %s in  %s.", ns, s)
			}
		}
		return datas, nil
	}
}

func IsEmptyBitMap(s string) bool {
	hex, err := string2data(s)

	if err != nil {
		return false
	}

	for i, v := range hex {
		if v != EmptyMapHex[i] {
			return false
		}
	}
	return true
}
