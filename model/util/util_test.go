package util

import (
	"testing"
)

func TestCbmLen(t *testing.T) {
	cbm := "fffff"
	if CbmLen(cbm) != 20 {
		t.Errorf("Wrong bdw cbm length")
	}

	cbm = "7ff"
	if CbmLen(cbm) != 11 {
		t.Errorf("Wrong skx cbm length")
	}
}

func TestSubtractStringSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}
	s := []string{"a", "c"}

	newslice := SubtractStringSlice(slice, s)

	if len(newslice) != 1 {
		t.Errorf("New slice length should be 1")
	}
	if newslice[0] != "b" {
		t.Errorf("New slice should be [\"2\"]")
	}
}
