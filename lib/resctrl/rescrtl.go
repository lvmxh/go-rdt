// +build linux

package resctrl

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	libutil "openstackcore-rdtagent/lib/util"
)

const (
	SysResctrl = "/sys/fs/resctrl"
)

var (
	// The absolute path to the root of the Intel RDT "resource control" filesystem
	intelRdtRootLock sync.Mutex
	intelRdtRoot     string
)

type NotFoundError struct {
	ResourceControl string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("mountpoint for %s not found", e.ResourceControl)
}

func NewNotFoundError(res string) error {
	return &NotFoundError{
		ResourceControl: res,
	}
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}

type CacheCos struct {
	Id  string
	cos string
}

//
type ResAssociation struct {
	Tasks    []string
	Cpus     string
	Schemata map[string][]CacheCos
}

//Usage:
//    policys := make(map[string]*ResAssociation)
//	  filepath.Walk(SysResctrl, ParserResAssociation(SysResctrl, ignore, policys))
func ParserResAssociation(basepath string, ignore []string, ps map[string]*ResAssociation) filepath.WalkFunc {
	parser := func(res *ResAssociation, name string, val []byte) error {
		switch name {
		case "Cpus":
			str := strings.TrimSpace(string(val))
			libutil.SetField(res, name, str)
			return nil
		case "Schemata":
			strs := strings.Split(string(val), "\n")
			if len(strs) > 1 {
				res.Schemata = make(map[string][]CacheCos)
			}
			for _, data := range strs {
				datas := strings.SplitAfterN(data, ":", 2)
				key := datas[0]
				if key == "" {
					return nil
				}

				if _, ok := res.Schemata[key]; !ok {
					res.Schemata[key] = make([]CacheCos, 0, 10)
				}

				coses := strings.Split(datas[1], ";")
				for _, cos := range coses {
					infos := strings.SplitN(cos, "=", 2)
					cacheCos := &CacheCos{infos[0], infos[1]}
					res.Schemata[key] = append(res.Schemata[key], *cacheCos)
				}

			}
			return nil
		default:
			strs := strings.Split(string(val), "\n")
			libutil.SetField(res, name, strs)
			return nil
		}
		return nil

	}

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// add log
			return nil
		}
		f := filepath.Base(path)
		rel, err := filepath.Rel(basepath, path)
		pkey := rel
		if info.IsDir() {
			// ignore dir.
			for _, d := range ignore {
				if d == f {
					return filepath.SkipDir
				}
			}
			ps[pkey] = &ResAssociation{}
			return nil
		}
		for _, d := range ignore {
			if d == f {
				return nil
			}
		}

		dir := filepath.Dir(path)
		rel, err = filepath.Rel(basepath, dir)
		pkey = rel

		name := strings.Replace(strings.Title(strings.Replace(f, "_", " ", -1)), " ", "", -1)
		data, err := ioutil.ReadFile(path)
		pl := ps[pkey]
		parser(pl, name, data)
		return nil
	}
}

// access the resctrl need flock to avoid race with other agent.
// Go does not support flock lib.
// That need cgo, please ref:
// https://gist.github.com/ericchiang/ce0fdcac5659d0a80b38
func GetResAssociation() map[string]*ResAssociation {
	ignore := []string{"info"}
	policys := make(map[string]*ResAssociation)
	filepath.Walk(SysResctrl, ParserResAssociation(SysResctrl, ignore, policys))
	return policys
}

type RdtCosInfo struct {
	CbmMask    string
	MinCbmBits int
	NumClosids int
}

/*
Usage:
    ignore := []string{"info"}  // ignore the toppath
	info := make(map[string]*RdtCosInfo)
    basepath := SysResctrl+"/info"
	filepath.Walk(basepath, ParserRdtCosInfo(basepath, ignore, info))
	fmt.Println(info["l3data"])  //for RDT, we can get info["l3data"]

*/
func ParserRdtCosInfo(basepath string, ignore []string, mres map[string]*RdtCosInfo) filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// add log
			return nil
		}
		f := filepath.Base(path)
		rel, err := filepath.Rel(basepath, path)
		pkey := rel
		if info.IsDir() {
			for _, d := range ignore {
				if d == f {
					return nil
				}
			}
			mres[strings.ToLower(pkey)] = &RdtCosInfo{}
			return nil
		}
		for _, d := range ignore {
			if d == f {
				return nil
			}
		}

		dir := filepath.Dir(path)
		rel, err = filepath.Rel(basepath, dir)
		pkey = strings.ToLower(rel)

		name := strings.Replace(strings.Title(strings.Replace(f, "_", " ", -1)), " ", "", -1)
		data, err := ioutil.ReadFile(path)
		strs := strings.TrimSpace(string(data))
		res := mres[pkey]
		libutil.SetField(res, name, strs)
		return nil
	}
}

// access the resctrl need flock to avoid race with other agent.
// Go does not support flock lib.
// That need cgo, please ref:
// https://gist.github.com/ericchiang/ce0fdcac5659d0a80b38
func GetRdtCosInfo() map[string]*RdtCosInfo {
	ignore := []string{"info"} // ignore the toppath
	info := make(map[string]*RdtCosInfo)
	basepath := SysResctrl + "/info"
	filepath.Walk(basepath, ParserRdtCosInfo(basepath, ignore, info))
	return info
}

// Return the mount point path of Intel RDT "resource control" filesysem
func findIntelRdtMountpointDir() (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		text := s.Text()
		fields := strings.Split(text, " ")
		// Safe as mountinfo encodes mountpoints with spaces as \040.
		index := strings.Index(text, " - ")
		postSeparatorFields := strings.Fields(text[index+3:])
		numPostFields := len(postSeparatorFields)

		// This is an error as we can't detect if the mount is for "Intel RDT"
		if numPostFields == 0 {
			return "", fmt.Errorf("Found no fields post '-' in %q", text)
		}

		if postSeparatorFields[0] == "resctrl" {
			// Check that the mount is properly formated.
			if numPostFields < 3 {
				return "", fmt.Errorf("Error found less than 3 fields post '-' in %q", text)
			}

			return fields[4], nil
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}

	return "", NewNotFoundError("Intel RDT")
}

// Gets the root path of Intel RDT "resource control" filesystem
func getIntelRdtRoot() (string, error) {
	intelRdtRootLock.Lock()
	defer intelRdtRootLock.Unlock()

	if intelRdtRoot != "" {
		return intelRdtRoot, nil
	}

	root, err := findIntelRdtMountpointDir()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(root); err != nil {
		return "", err
	}

	intelRdtRoot = root
	return intelRdtRoot, nil
}

func isIntelRdtMounted() bool {
	_, err := getIntelRdtRoot()
	if err != nil {
		if !IsNotFound(err) {
			return false
		}
	}
	return true
}

func DisableRdt() bool {
	if isIntelRdtMounted() {
		if err := exec.Command("umount", "/sys/fs/resctrl").Run(); err != nil {
			return false
		}
	}
	return true
}

func EnableCat() bool {
	// mount -t resctrl resctrl /sys/fs/resctrl
	if err := os.MkdirAll("/sys/fs/resctrl", 0755); err != nil {
		return false
	}
	if err := exec.Command("mount", "-t", "resctrl", "resctrl", "/sys/fs/resctrl").Run(); err != nil {
		return false
	}
	return true
}

func EnableCdp() bool {
	// mount -t resctrl -o cdp resctrl /sys/fs/resctrl
	if err := os.MkdirAll("/sys/fs/resctrl", 0755); err != nil {
		return false
	}
	if err := exec.Command("mount", "-t", "resctrl", "-o", "cdp", "resctrl", "/sys/fs/resctrl").Run(); err != nil {
		return false
	}
	return true
}
