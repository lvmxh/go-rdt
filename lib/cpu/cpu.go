package cpu

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	SysCpu = "/sys/devices/system/cpu"
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
