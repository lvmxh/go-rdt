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
