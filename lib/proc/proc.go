package proc

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// CPUInfoPath is the patch to cpuinfo
	CPUInfoPath = "/proc/cpuinfo"
	// MountInfoPath is the mount info path
	MountInfoPath = "/proc/self/mountinfo"
	// ResctrlPath is the patch to resctrl
	ResctrlPath = "/sys/fs/resctrl"
)

// rdt_a, cat_l3, cdp_l3, cqm, cqm_llc, cqm_occup_llc
// cqm_mbm_total, cqm_mbm_local
func parseCPUInfoFile(flag string) (bool, error) {
	f, err := os.Open(CPUInfoPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if err := s.Err(); err != nil {
			return false, err
		}

		text := s.Text()
		flags := strings.Split(text, " ")

		for _, f := range flags {
			if f == flag {
				return true, nil
			}
		}
	}
	return false, nil
}

// IsRdtAvailiable returns RDT feature available or not
func IsRdtAvailiable() (bool, error) {
	return parseCPUInfoFile("rdt_a")
}

// IsCqmAvailiable returns CMT feature available or not
func IsCqmAvailiable() (bool, error) {
	return parseCPUInfoFile("cqm")
}

// IsCdpAvailiable returns CDP feature available or not
func IsCdpAvailiable() (bool, error) {
	return parseCPUInfoFile("cdp_l3")
}

// we can use shell command: "mount -l -t resctrl"
func findMountDir(mountdir string) (string, error) {
	f, err := os.Open(MountInfoPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		text := s.Text()
		if strings.Contains(text, mountdir) {
			// http://man7.org/linux/man-pages/man5/proc.5.html
			// text = strings.Replace(text, " - ", " ", -1)
			// fields := strings.Split(text, " ")[4:]
			return text, nil
		}
	}
	return "", fmt.Errorf("Can not found the mount entry: %s", mountdir)
}

// IsEnableRdt returns if RDT is enabled or not
func IsEnableRdt() bool {
	mount, err := findMountDir(ResctrlPath)
	if err != nil {
		return false
	}
	return len(mount) > 0
}

// IsEnableCdp returns if CDP is enabled or not
func IsEnableCdp() bool {
	var flag = "cdp"
	mount, err := findMountDir(ResctrlPath)
	if err != nil {
		return false
	}
	return strings.Contains(mount, flag)
}

// IsEnableCat returns if CAT is enabled or not
func IsEnableCat() bool {
	var flag = "cdp"
	mount, err := findMountDir(ResctrlPath)
	if err != nil {
		return false
	}
	return !strings.Contains(mount, flag) && len(mount) > 0
}

// Process struct with pid and command line
type Process struct {
	Pid     int
	CmdLine string
}

// ListProcesses returns all process on the host
var ListProcesses = func() map[string]Process {
	processes := make(map[string]Process)
	files, _ := filepath.Glob("/proc/[0-9]*/cmdline")
	for _, file := range files {

		listfs := strings.Split(file, "/")
		if pid, err := strconv.Atoi(listfs[2]); err == nil {

			cmd, _ := ioutil.ReadFile(file)
			cmdString := strings.Join(strings.Split(string(cmd), "\x00"), " ")
			processes[listfs[2]] = Process{pid, cmdString}
		}
	}

	return processes
}
