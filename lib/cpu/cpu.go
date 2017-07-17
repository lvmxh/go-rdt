package cpu

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	SysCpu      = "/sys/devices/system/cpu"
	CpuInfoPath = "/proc/cpuinfo"
)

var cpunOnce sync.Once
var isolCPUsOnce sync.Once

var cpuNumber int = 0
var isolatedCPUs string = ""

// REF: https://www.kernel.org/doc/Documentation/cputopology.txt
// another way is call sysconf via cgo, like libpqos
func HostCpuNum() int {
	cpunOnce.Do(func() {
		path := filepath.Join(SysCpu, "possible")
		data, _ := ioutil.ReadFile(path)
		strs := strings.TrimSpace(string(data))
		num, err := strconv.Atoi(strings.SplitN(strs, "-", 2)[1])
		if err != nil {
			log.Fatalf("Failed to get cup numbers on host: %v", err)
			return
		}
		num++
		cpuNumber = num
	})
	return cpuNumber
}

// ignore stepping and processor type.
// NOTE, Guess all cpus in one hose are same microarch
func GetSignature() uint32 {
	// family, model string
	var family, model int

	f, err := os.Open(CpuInfoPath)
	if err != nil {
		return 0
	}
	defer f.Close()

	br := bufio.NewReader(f)
	find := 0

	for err == nil {
		s, err := br.ReadString('\n')
		if err != nil {
			return 0
		}
		if strings.HasPrefix(s, "cpu family") {
			sl := strings.Split(s, ":")
			family, _ = strconv.Atoi(strings.TrimSpace(sl[1]))
			find++
		} else if strings.HasPrefix(s, "model") {
			sl := strings.Split(s, ":")
			if strings.TrimSpace(sl[0]) == "model" {
				model, _ = strconv.Atoi(strings.TrimSpace(sl[1]))
				find++
			}
		}
		if find >= 2 {
			sig := (family>>4)<<20 + (family&0xf)<<8
			sig |= (model>>4)<<16 + (model&0xf)<<4
			return uint32(sig)
		}
	}
	return 0
}

// Get isolated CPUs.
// The result will be as follow:
// 2-21,24-43,46-65,68-87
// This result can generate a Bitmap
func IsolatedCPUs() string {
	isolCPUsOnce.Do(func() {
		path := filepath.Join(SysCpu, "isolated")
		data, _ := ioutil.ReadFile(path)
		strs := strings.TrimSpace(string(data))
		isolatedCPUs = strs
	})
	return isolatedCPUs
}
