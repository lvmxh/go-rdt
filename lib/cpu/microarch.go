package cpu

// ignore stepping
var m = map[uint32]string{
	0x406e0: "Skylake",
	0x506e0: "Skylake",
	0x50650: "Skylake",
	0x806e0: "Kaby Lake",
	0x906e0: "Kaby Lake",
}

func GetMicroArch(sig uint32) string {
	s := sig & 0xFFFFFFF0
	if v, ok := m[s]; ok {
		return v
	}
	return ""
}
