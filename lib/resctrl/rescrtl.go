// +build linux

package resctrl

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

func writeFile(dir, file, data string) error {
	if dir == "" {
		return fmt.Errorf("no such directory for %s", file)
	}
	if !strings.HasSuffix(data, "\n") {
		data = data + "\n"
	}
	if err := ioutil.WriteFile(filepath.Join(dir, file), []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write %v to %v: %v", data, file, err)
	}
	return nil
}

func DestroyResAssociation(group string) error {
	path := filepath.Join(SysResctrl, group)
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

type CacheCos struct {
	Id   uint8
	Mask uint32
}

// FIXME, should Tasks be int?
type ResAssociation struct {
	Tasks    []string
	Cpus     string
	Schemata map[string][]CacheCos
}

//Usage:
//    ress := make(map[string]*ResAssociation)
//	  filepath.Walk(SysResctrl, ParserResAssociation(SysResctrl, ignore, ress))
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
				datas := strings.SplitN(data, ":", 2)
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
					id, _ := strconv.ParseUint(infos[0], 10, 8)
					mask, _ := strconv.ParseUint(infos[1], 16, 32)
					cacheCos := &CacheCos{uint8(id), uint32(mask)}
					res.Schemata[key] = append(res.Schemata[key], *cacheCos)
				}

			}
			return nil
		default:
			strs := strings.Split(string(val), "\n")
			// FIXME, remove the last element, it's a empty string
			// It will cause error while write back tasks to resctrl
			libutil.SetField(res, name, strs[:len(strs)-1])
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
// now we can use lib/flock/flock.go
func GetResAssociation() map[string]*ResAssociation {
	ignore := []string{"info"}
	ress := make(map[string]*ResAssociation)
	filepath.Walk(SysResctrl, ParserResAssociation(SysResctrl, ignore, ress))
	return ress
}

func (r ResAssociation) Commit(group string) error {
	if !isIntelRdtMounted() {
		return fmt.Errorf("Can't apply this association, for resctrl is not mounted!")
	}

	path := SysResctrl

	if strings.ToLower(group) != "default" && group != "." {
		path = filepath.Join(SysResctrl, group)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		}
	}
	if r.Cpus != "" {
		writeFile(path, "cpus", r.Cpus)
	} else {
		if _, err := os.Stat(path); os.IsExist(err) {
			writeFile(path, "cpus", r.Cpus)
		}
	}
	// only commit a user deinfed group's task to sys fs
	if group != "." && len(r.Tasks) > 0 {
		// write one task one time, or write will fail
		for _, t := range r.Tasks {
			err := writeFile(path, "tasks", t)
			if err != nil {
				return err
			}
		}
	}

	if len(r.Schemata) > 0 {
		schemata := make([]string, 0, 10)
		for k, v := range r.Schemata {
			str := make([]string, 0, 10)
			// resctrl require we have strict cache id order
			for cacheid := 0; cacheid < len(v); cacheid++ {
				for _, cos := range v {
					if uint8(cacheid) == cos.Id {
						str = append(str, fmt.Sprintf("%d=%x", cos.Id, cos.Mask))
						break
					}
				}
			}
			schemata = append(schemata, strings.Join([]string{k, strings.Join(str, ";")}, ":"))
		}
		data := strings.Join(schemata, "\n")
		err := writeFile(path, "schemata", data)
		if err != nil {
			path := filepath.Join(SysResctrl, group)
			os.Remove(path)
		}
		return err
	}

	return nil
}

// FIXME need to catch error
func CommitAll(m_res map[string]*ResAssociation) error {
	ress := GetResAssociation()
	for name, res := range m_res {
		res.Commit(name)
	}

	// Golang does not support set difference
	for name := range ress {
		if _, ok := m_res[name]; !ok && name != "." {
			DestroyResAssociation(name)
		}
	}
	return nil
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
// now we can use lib/flock/flock.go
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
