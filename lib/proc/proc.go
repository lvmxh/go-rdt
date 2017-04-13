package proc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	CpuInfoPath   = "/proc/cpuinfo"
	MountInfoPath = "/proc/self/mountinfo"
	ResctrlPath   = "/sys/fs/resctrl"
)

// rdt_a, cat_l3, cdp_l3, cqm, cqm_llc, cqm_occup_llc
// cqm_mbm_total, cqm_mbm_local
func parseCpuInfoFile(flag string) (bool, error) {
	f, err := os.Open(CpuInfoPath)
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

func IsRdtAvailiable() (bool, error) {
	return parseCpuInfoFile("rdt_a")
}

func IsCqmAvailiable() (bool, error) {
	return parseCpuInfoFile("cqm")
}

func IsCdpAvailiable() (bool, error) {
	return parseCpuInfoFile("cdp_l3")
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
	return "", fmt.Errorf("Can not found the mount entry: %s!", mountdir)
}

func IsEnableRdt() bool {
	mount, err := findMountDir(ResctrlPath)
	if err != nil {
		return false
	}
	return len(mount) > 0
}

func IsEnableCdp() bool {
	var flag = "cdp"
	mount, err := findMountDir(ResctrlPath)
	if err != nil {
		return false
	}
	return strings.Contains(mount, flag)
}

func IsEnableCat() bool {
	var flag = "cdp"
	mount, err := findMountDir(ResctrlPath)
	if err != nil {
		return false
	}
	return !strings.Contains(mount, flag) && len(mount) > 0
}
