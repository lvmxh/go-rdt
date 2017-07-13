package cpu

import (
	"fmt"
	"testing"
)

func TestGetSignature(t *testing.T) {
	sig := getSignature()
	fmt.Printf("CPU signature is 0x%x.\n", sig)
	if sig == 0 {
		t.Errorf("CPU signature should be >0.\n")
	}
}
func TestGetMicroArch(t *testing.T) {
	m := GetMicroArch(0x50650)
	fmt.Println("CPU MicroArch is", m)
	if m != "Skylake" {
		t.Errorf("CPU MicroArch should be %s.\n", "Skylake")
	}

	m = GetMicroArch(0x50659)
	fmt.Println("CPU MicroArch is", m)
	if m != "Skylake" {
		t.Errorf("CPU MicroArch should be %s.\n", "Skylake")
	}
}
