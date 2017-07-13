package cpu

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	SysCpu      = "/sys/devices/system/cpu"
	CpuInfoPath = "/proc/cpuinfo"
)

// REF: https://www.kernel.org/doc/Documentation/cputopology.txt
// another way is call sysconf via cgo, like libpqos
func HostCpuNum() (int, error) {

	path := filepath.Join(SysCpu, "possible")
	data, _ := ioutil.ReadFile(path)
	strs := strings.TrimSpace(string(data))
	num, err := strconv.Atoi(strings.SplitN(strs, "-", 2)[1])
	if err != nil {
		return 0, err
	}
	num++
	return num, err
}

// ignore stepping and processor type.
func getSignature() uint32 {
	// family, model string
	var family, model string

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
			family = strings.TrimSpace(sl[1])
			find++
		} else if strings.HasPrefix(s, "model") {
			sl := strings.Split(s, ":")
			if strings.TrimSpace(sl[0]) == "model" {
				model = strings.TrimSpace(sl[1])
				find++
			}
		}
		if find >= 2 {
			if len(model) == 1 {
				model = "0" + model
			}
			if len(family) == 1 {
				family = "0" + family
			}
			exf := family[0:1]
			f := family[1:2]
			exm := model[0:1]
			m := model[1:2]
			sig, err := strconv.ParseUint(exf+exm+"0"+f+m+"0", 16, 32)
			if err != nil {
				return 0
			}
			return uint32(sig)
		}
	}
	return 0
}
