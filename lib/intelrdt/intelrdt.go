// +build linux

package intelrdt

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

	"openstackcore-rdtagent/lib/configs"
	"openstackcore-rdtagent/lib/resourcemanager"
)

/*
 * About Intel RDT/CAT feature:
 * Intel platforms with new Xeon CPU support Resource Director Technology (RDT).
 * Intel Cache Allocation Technology (CAT) is a sub-feature of RDT. Currently L3
 * Cache is the only resource that is supported in RDT.
 *
 * This feature provides a way for the software to restrict cache allocation to a
 * defined 'subset' of L3 cache which may be overlapping with other 'subsets'.
 * The different subsets are identified by class of service (CLOS) and each CLOS
 * has a capacity bitmask (CBM).
 *
 * For more information about Intel RDT/CAT can be found in the section 17.17
 * of Intel Software Developer Manual.
 *
 * About Intel RDT/CAT kernel interface:
 * In Linux kernel, the interface is defined and exposed via "resource control"
 * filesystem, which is a "cgroup-like" interface.
 *
 * Comparing with cgroups, it has similar process management lifecycle and
 * interfaces in a container. But unlike cgroups' hierarchy, it has single level
 * filesystem layout.
 *
 * Intel RDT "resource control" filesystem hierarchy:
 * mount -t resctrl resctrl /sys/fs/resctrl
 * tree /sys/fs/resctrl
 * /sys/fs/resctrl/
 * |-- info
 * |   |-- L3
 * |       |-- cbm_mask
 * |       |-- min_cbm_bits
 * |       |-- num_closids
 * |-- cpus
 * |-- schemata
 * |-- tasks
 * |-- <container_id>
 *     |-- cpus
 *     |-- schemata
 *     |-- tasks
 *
 * For runc, we can make use of `tasks` and `schemata` configuration for L3 cache
 * resource constraints.
 *
 *  The file `tasks` has a list of tasks that belongs to this group (e.g.,
 * <container_id>" group). Tasks can be added to a group by writing the task ID
 * to the "tasks" file  (which will automatically remove them from the previous
 * group to which they belonged). New tasks created by fork(2) and clone(2) are
 * added to the same group as their parent. If a pid is not in any sub group, it is
 * in root group.
 *
 * The file `schemata` has allocation bitmasks/values for L3 cache on each socket,
 * which contains L3 cache id and capacity bitmask (CBM).
 * 	Format: "L3:<cache_id0>=<cbm0>;<cache_id1>=<cbm1>;..."
 * For example, on a two-socket machine, L3's schema line could be `L3:0=ff;1=c0`
 * which means L3 cache id 0's CBM is 0xff, and L3 cache id 1's CBM is 0xc0.
 *
 * The valid L3 cache CBM is a *contiguous bits set* and number of bits that can
 * be set is less than the max bit. The max bits in the CBM is varied among
 * supported Intel Xeon platforms. In Intel RDT "resource control" filesystem
 * layout, the CBM in a group should be a subset of the CBM in root. Kernel will
 * check if it is valid when writing. e.g., 0xfffff in root indicates the max bits
 * of CBM is 20 bits, which mapping to entire L3 cache capacity. Some valid CBM
 * values to set in a group: 0xf, 0xf0, 0x3ff, 0x1f00 and etc.
 *
 * For more information about Intel RDT/CAT kernel interface:
 * https://www.kernel.org/doc/Documentation/x86/intel_rdt_ui.txt
 *
 * An example for runc:
 * There are two L3 caches in the two-socket machine, the default CBM is 0xfffff
 * and the max CBM length is 20 bits. This configuration assigns 4/5 of L3 cache
 * id 0 and the whole L3 cache id 1 for the container:
 *
 * "linux": {
 * 	"intelRdt": {
 * 		"l3CacheSchema": "L3:0=ffff0;1=fffff"
 * 	}
 * }
 */

type Manager interface {
	resourcemanager.ResourceManager

	// Returns Intel RDT "resource control" filesystem path to save in
	// a state file and to be able to restore the object later
	GetPath() string
}

// This implements interface Manager
type IntelRdtManager struct {
	mu     sync.Mutex
	Config *configs.Config
	Id     string
	Path   string
}

const (
	IntelRdtTasks = "tasks"
	SysResctrl    = "/sys/fs/resctrl"
)

var (
	// The absolute path to the root of the Intel RDT "resource control" filesystem
	intelRdtRootLock sync.Mutex
	intelRdtRoot     string
)

// The read-only Intel RDT related system information in root
type IntelRdtInfo struct {
	CbmMask    uint64 `json:"cbm_mask,omitempty"`
	MinCbmBits uint64 `json:"min_cbm_bits,omitempty"`
	NumClosid  uint64 `json:"num_closid,omitempty"`
}

type intelRdtData struct {
	root   string
	config *configs.Config
	pid    int
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

		// If not mounted, we try to mount again:
		// mount -t resctrl resctrl /sys/fs/resctrl
		if err := os.MkdirAll("/sys/fs/resctrl", 0755); err != nil {
			return false
		}
		if err := exec.Command("mount", "-t", "resctrl", "resctrl", "/sys/fs/resctrl").Run(); err != nil {
			return false
		}
	}

	return true
}

func parseCpuInfoFile(path string) (bool, error) {
	f, err := os.Open(path)
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

		for _, flag := range flags {
			if flag == "rdt_a" {
				return true, nil
			}
		}
	}
	return false, nil
}

func parseUint(s string, base, bitSize int) (uint64, error) {
	value, err := strconv.ParseUint(s, base, bitSize)
	if err != nil {
		intValue, intErr := strconv.ParseInt(s, base, bitSize)
		// 1. Handle negative values greater than MinInt64 (and)
		// 2. Handle negative values lesser than MinInt64
		if intErr == nil && intValue < 0 {
			return 0, nil
		} else if intErr != nil && intErr.(*strconv.NumError).Err == strconv.ErrRange && intValue < 0 {
			return 0, nil
		}

		return value, err
	}

	return value, nil
}

// Gets a single uint64 value from the specified file.
func getIntelRdtParamUint(path, file string) (uint64, error) {
	fileName := filepath.Join(path, file)
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, err
	}

	res, err := parseUint(strings.TrimSpace(string(contents)), 10, 64)
	if err != nil {
		return res, fmt.Errorf("unable to parse %q as a uint from file %q", string(contents), fileName)
	}
	return res, nil
}

// Gets a string value from the specified file
func getIntelRdtParamString(path, file string) (string, error) {
	contents, err := ioutil.ReadFile(filepath.Join(path, file))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(contents)), nil
}

func readTasksFile(dir string) ([]int, error) {
	f, err := os.Open(filepath.Join(dir, IntelRdtTasks))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		s   = bufio.NewScanner(f)
		out = []int{}
	)

	for s.Scan() {
		if t := s.Text(); t != "" {
			pid, err := strconv.Atoi(t)
			if err != nil {
				return nil, err
			}
			out = append(out, pid)
		}
	}
	return out, nil
}

func writeFile(dir, file, data string) error {
	if dir == "" {
		return fmt.Errorf("no such directory for %s", file)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, file), []byte(data+"\n"), 0700); err != nil {
		return fmt.Errorf("failed to write %v to %v: %v", data, file, err)
	}
	return nil
}

func getIntelRdtData(c *configs.Config, pid int) (*intelRdtData, error) {
	rootPath, err := getIntelRdtRoot()
	if err != nil {
		return nil, err
	}
	return &intelRdtData{
		root:   rootPath,
		config: c,
		pid:    pid,
	}, nil
}

// WriteIntelRdtTasks writes the specified pid into the "tasks" file
func WriteIntelRdtTasks(dir string, pid int) error {
	if dir == "" {
		return fmt.Errorf("no such directory for %s", IntelRdtTasks)
	}

	// Dont attach any pid if -1 is specified as a pid
	if pid != -1 {
		if err := ioutil.WriteFile(filepath.Join(dir, IntelRdtTasks), []byte(strconv.Itoa(pid)), 0700); err != nil {
			return fmt.Errorf("failed to write %v to %v: %v", pid, IntelRdtTasks, err)
		}
	}
	return nil
}

// Check if Intel RDT is enabled
func IsIntelRdtEnabled() bool {
	// 1. check if hardware and kernel support Intel RDT feature
	// "rdt" flag is set if supported
	isFlagSet, err := parseCpuInfoFile("/proc/cpuinfo")
	if err != nil {
		return false
	}

	// 2. check if Intel RDT "resource control" filesystem is mounted
	isMounted := isIntelRdtMounted()

	return isFlagSet && isMounted
}

// Get Intel RDT "resource control" filesystem path
func GetIntelRdtPath(id string) (string, error) {
	rootPath, err := getIntelRdtRoot()
	if err != nil {
		return "", err
	}

	path := filepath.Join(rootPath, id)
	return path, nil
}

// Get read-only Intel RDT related system information
func GetIntelRdtInfo() (*IntelRdtInfo, error) {
	intelRdtInfo := &IntelRdtInfo{}

	rootPath, err := getIntelRdtRoot()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(rootPath, "info", "l3")
	cbmMask, err := getIntelRdtParamUint(path, "cbm_mask")
	if err != nil {
		return nil, err
	}
	minCbmBits, err := getIntelRdtParamUint(path, "min_cbm_bits")
	if err != nil {
		return nil, err
	}
	numClosid, err := getIntelRdtParamUint(path, "num_closid")
	if err != nil {
		return nil, err
	}

	intelRdtInfo.CbmMask = cbmMask
	intelRdtInfo.MinCbmBits = minCbmBits
	intelRdtInfo.NumClosid = numClosid

	return intelRdtInfo, nil
}

// Applies configuration to the process with the specified pid
func (m *IntelRdtManager) Apply(pid int) (err error) {
	d, err := getIntelRdtData(m.Config, pid)
	if err != nil && !IsNotFound(err) {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	path, err := d.join(m.Id)
	if err != nil {
		return err
	}

	m.Path = path
	return nil
}

// Returns the PIDs inside Intel RDT "resource control" filesystem at path
func (m *IntelRdtManager) GetPids() ([]int, error) {
	return readTasksFile(m.GetPath())
}

// Returns all the PIDs inside Intel RDT "resource control" filesystem at path
func (m *IntelRdtManager) GetAllPids() ([]int, error) {
	return m.GetPids()
}

// Toggles the freezer cgroup according with specified state
// func (m *IntelRdtManager) Freeze(state configs.FreezerState) error {
// 	return nil
// }

// Destroys the Intel RDT "resource control" filesystem
func (m *IntelRdtManager) Destroy() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := os.RemoveAll(m.Path); err != nil {
		return err
	}
	m.Path = ""
	return nil
}

// Returns Intel RDT "resource control" filesystem paths to save in
// a state file and to be able to restore the object later
func (m *IntelRdtManager) GetPaths() map[string]string {
	m.mu.Lock()
	paths := make(map[string]string)
	paths["intelrdt"] = m.Path
	m.mu.Unlock()
	return paths
}

// Returns Intel RDT "resource control" filesystem path to save in
// a state file and to be able to restore the object later
func (m *IntelRdtManager) GetPath() string {
	if m.Path == "" {
		m.Path, _ = GetIntelRdtPath(m.Id)
	}
	return m.Path
}

// Returns statistics for Intel RDT
func (m *IntelRdtManager) GetStats() (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stats := NewStats()

	// The read-only default "schemata" in root, for reference
	rootPath, err := getIntelRdtRoot()
	if err != nil {
		return nil, err
	}
	schemaRoot, err := getIntelRdtParamString(rootPath, "schemata")
	if err != nil {
		return nil, err
	}
	stats.IntelRdtRootStats.L3CacheSchema = schemaRoot

	// The stats in "container_id" group
	schema, err := getIntelRdtParamString(m.GetPath(), "schemata")
	if err != nil {
		return nil, err
	}
	stats.IntelRdtStats.L3CacheSchema = schema

	return stats, nil
}

// Set Intel RDT "resource control" filesystem as configured.
func (m *IntelRdtManager) Set(container *configs.Config) error {
	path := m.GetPath()

	// About L3 cache schema file:
	// The schema has allocation masks/values for L3 cache on each socket,
	// which contains L3 cache id and capacity bitmask (CBM).
	//     Format: "L3:<cache_id0>=<cbm0>;<cache_id1>=<cbm1>;..."
	// For example, on a two-socket machine, L3's schema line could be:
	//     L3:0=ff;1=c0
	// Which means L3 cache id 0's CBM is 0xff, and L3 cache id 1's CBM is 0xc0.
	//
	// About L3 cache CBM validity:
	// The valid L3 cache CBM is a *contiguous bits set* and number of
	// bits that can be set is less than the max bit. The max bits in the
	// CBM is varied among supported Intel Xeon platforms. In Intel RDT
	// "resource control" filesystem layout, the CBM in a group should
	// be a subset of the CBM in root. Kernel will check if it is valid
	// when writing.
	// e.g., 0xfffff in root indicates the max bits of CBM is 20 bits,
	// which mapping to entire L3 cache capacity. Some valid CBM values
	// to set in a group: 0xf, 0xf0, 0x3ff, 0x1f00 and etc.
	l3CacheSchema := container.IntelRdt.L3CacheSchema
	if l3CacheSchema != "" {
		if err := writeFile(path, "schemata", l3CacheSchema); err != nil {
			return err
		}
	}

	return nil
}

func (raw *intelRdtData) join(id string) (string, error) {
	path := filepath.Join(raw.root, id)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	if err := WriteIntelRdtTasks(path, raw.pid); err != nil {
		return "", err
	}
	return path, nil
}

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

func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	val := reflect.ValueOf(value)

	structFieldValue.Set(val)
	return nil
}

//
type ResAssociation struct {
	Tasks    []string
	Cpus     []string
	Schemata []string
}

//Usage:
//    policys := make(map[string]*ResAssociation)
//	  filepath.Walk(SysResctrl, ParserResAssociation(SysResctrl, ignore, policys))
func ParserResAssociation(basepath string, ignore []string, ps map[string]*ResAssociation) filepath.WalkFunc {

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
		strs := strings.Split(string(data), "\n")
		pl := ps[pkey]
		SetField(pl, name, strs)
		return nil
	}
}

// access the resctrl need flock to avoid race with other agent.
// Go does not support flock lib.
// That need cgo, please ref:
// https://gist.github.com/ericchiang/ce0fdcac5659d0a80b38
func GetResAssociation() map[string]*ResAssociation {
	ignore = []string{"info"}
	policys := make(map[string]*ResAssociation)
	filepath.Walk(SysResctrl, ParserResAssociation(SysResctrl, ignore, policys))
	return policys
}

type RdtCosInfo struct {
	CbmMask    string
	MinCbmBits int
	NumClosids int
}

func GetL3CosInfo(typ string) (*RdtCosInfo, error) {
	// typ can be L3CODE, L3DATA
	// FIXME, we need to check the L2 and L3 RDT case.
	cos := &RdtCosInfo{}
	data, err := ioutil.ReadFile(SysResctrl + "/info/" + typ + "/cbm_mask")
	if err != nil {
		// add log
		return cos, err
	}
	cos.CbmMask = string(data)
	data, err = ioutil.ReadFile(SysResctrl + "/info/" + typ + "/min_cbm_bits")
	if err != nil {
		// add log
		return cos, err
	}
	cos.MinCbmBits, err = strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		// add log
		return cos, err
	}
	data, err = ioutil.ReadFile(SysResctrl + "/info/" + typ + "/num_closids")
	if err != nil {
		// add log
		return cos, err
	}
	cos.NumClosids, err = strconv.Atoi(strings.TrimSpace(string(data)))
	return cos, nil
}
