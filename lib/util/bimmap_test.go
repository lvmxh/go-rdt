package util

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

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