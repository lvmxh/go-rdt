package cpu

import (
	"fmt"
	"testing"
)

func TestIsZeroHexString(t *testing.T) {
	sig := getSignature()
	fmt.Println("CPU signature is", sig)
	if sig == 0 {
		t.Errorf("CPU signature should be >0.\n")
	}
}
