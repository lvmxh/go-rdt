package workload

// workload api objects to represent resources in RDTAgent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"sync"

	libcache "openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	libutil "openstackcore-rdtagent/lib/util"

	. "openstackcore-rdtagent/api/error"
	"openstackcore-rdtagent/model/cache"
	"openstackcore-rdtagent/model/policy"
	modelutil "openstackcore-rdtagent/model/util"
)

// global lock for when doing enforce/update/release for a workload.
// This is a simple way to control RDAgent to access resctrl one
// goroutine one time
var l sync.Mutex

const (
	Successful = "Successful"
	Failed     = "Failed"
	Invalid    = "Invalid"
	None       = "None"
)

type RDTWorkLoad struct {
	// ID
	ID string
	// core ids, the work load run on top of cores/cpus
	CoreIDs []string `json:"core_ids"`
	// task ids, the work load's task ids
	TaskIDs []string `json:"task_ids"`
	// policy the workload want to apply
	Policy string `json:"policy"`
	// Status
	Status string
	// Group
	Group []string `json:"group"`
	// CosName
	CosName string
}

// build this struct when create Resasscciation
type EnforceRequest struct {
	// all resassociations on the host
	Resall map[string]*resctrl.ResAssociation
	// on which group we allocate cache
	BaseGrp string
	// new group name of the workload
	NewGrp string
	// sub group list in the base groups
	// this will be used to calculate offset
	SubGrps []string
	// max cache ways
	MaxWays uint32
	// min cache ways, not used yet
	MinWays uint32
	// enforce on which cache ids
	Cache_IDs []uint32
	// consume from base group or not
	Consume bool
}

func (w *RDTWorkLoad) Validate() error {
	if len(w.TaskIDs) <= 0 && len(w.CoreIDs) <= 0 {
		return fmt.Errorf("No task or core id specified")
	}

	// Firstly verify the task.
	ps := proc.ListProcesses()
	for _, task := range w.TaskIDs {
		if _, ok := ps[task]; !ok {
			return fmt.Errorf("The task: %s does not exist", task)
		}
	}
	// user don't need to provide group name anymore, if we configured
	// infra_group, then let RDAgent append the group name
	// e.g. w.Group = append(w.Group, "infra")
	// e.g. w.Group = append(w.Group, "w.name")
	if len(w.Group) > 2 {
		return fmt.Errorf("Can not specified more then 2 group list name")
	}

	return nil
}

func (w *RDTWorkLoad) Enforce() *AppError {
	if err := w.Validate(); err != nil {
		log.Errorf("Failed to validate workload %s, error: %s", w.ID, err)
		w.Status = Invalid
		return NewAppError(http.StatusBadRequest,
			"Failed to validate workload.", err)
	}

	w.Status = None

	cpubitstr := ""
	if len(w.CoreIDs) >= 0 {
		bm, err := CpuBitmaps(w.CoreIDs)
		if err != nil {
			return NewAppError(http.StatusBadRequest,
				"Failed to Pareser workload coreIDs.", err)
		}
		cpubitstr = bm.ToString()
	}

	// status will be updated to successful if no errors
	w.Status = Failed

	l.Lock()
	defer l.Unlock()
	resaall := resctrl.GetResAssociation()

	base_grp, new_grp, sub_grp := getGroupNames(w, resaall)

	if base_grp == "" {
		// log group information
		return AppErrorf(http.StatusBadRequest, "Faild to find a suitable group")
	}

	log.Debugf("base group %s, new group %s, sub group %v", base_grp, new_grp, sub_grp)

	pf := cpu.GetMicroArch(cpu.GetSignature())
	if pf == "" {
		return AppErrorf(http.StatusInternalServerError,
			"Unknow platform, please update the cpu_map.toml conf file.")
	}

	p, err := policy.GetPolicy(strings.ToLower(pf), w.Policy)
	if err != nil {
		return NewAppError(http.StatusInternalServerError,
			"Could not find the Polciy.", err)
	}

	ways, err := strconv.Atoi(p["MaxCache"])
	if err != nil {
		return NewAppError(http.StatusInternalServerError,
			"Error define MaxCache in Polciy.", err)
	}

	cacheinfo := &cache.CacheInfos{}
	cacheinfo.GetByLevel(libcache.GetLLC())

	cpunum := cpu.HostCpuNum()
	if cpunum == 0 {
		return AppErrorf(http.StatusInternalServerError,
			"Unable to get Total CPU numbers on Host")
	}

	er := &EnforceRequest{Resall: resaall,
		BaseGrp:   base_grp,
		NewGrp:    new_grp,
		SubGrps:   sub_grp,
		MaxWays:   uint32(ways),
		Cache_IDs: getCacheIDs(cpubitstr, cacheinfo, cpunum)}

	targetResAss, err := createOrGetResAss(er)
	if err != nil {
		log.Errorf("Error while try to create resource group for workload %s", w.ID)
		return NewAppError(http.StatusInternalServerError,
			"Error to create resource group.", err)
	}

	targetResAss.Tasks = append(targetResAss.Tasks, w.TaskIDs...)
	targetResAss.CPUs = cpubitstr

	if base_grp != new_grp && base_grp != "." {
		new_grp = base_grp + "-" + new_grp
	}

	if err = targetResAss.Commit(new_grp); err != nil {
		log.Errorf("Error while try to commit resource group for workload %s, group name %s", w.ID, new_grp)
		return NewAppError(http.StatusInternalServerError,
			"Error to commit resource group for workload.", err)
	}

	if base_grp == "." {
		if err = resaall["."].Commit("."); err != nil {
			log.Errorf("Error while try to commit resource group for default group")
			resctrl.DestroyResAssociation(new_grp)
			return NewAppError(http.StatusInternalServerError,
				"Error while try to commit resource group for default group.", err)
		}
	}

	if len(w.Group) == 0 {
		w.Group = append(w.Group, new_grp)
	}

	w.CosName = new_grp
	w.Status = Successful
	return nil
}

// Release Cos
func (w *RDTWorkLoad) Release() error {
	l.Lock()
	defer l.Unlock()

	resaall := resctrl.GetResAssociation()

	r, ok := resaall[w.CosName]

	if !ok {
		return nil
	}

	r.Tasks = modelutil.SubtractStringSlice(r.Tasks, w.TaskIDs)

	// safely remove resource group if no tasks and cpu bit map is empty
	if len(r.Tasks) < 1 {
		log.Printf("Remove resource group: %s", w.CosName)
		if err := resctrl.DestroyResAssociation(w.CosName); err != nil {
			return err
		}
		delete(resaall, w.CosName)
		newDefaultGrp := calculateDefaultGroup(resaall, []string{"."}, true)
		newDefaultGrp.Commit(".")
	}
	// remove workload task ids from resource group
	if len(w.TaskIDs) > 0 {
		if err := resctrl.RemoveTasks(w.TaskIDs); err != nil {
			log.Printf("Ignore Error while remove tasks %s", err)
			return nil
		}
	}

	return nil
}

// return base group name, new group name, sub group name list.
// e.g.
// CG1 L3:0=ffff;1=ffff
// CG1-SUB1 L3:0=f;1=f
// CG2 L3:0=f0000;1=f0000
//
// if w.Group is ["CG1", "SUB2"]
// getGroupNames will return CG1, SUB2, [CG1-SUB1]
func getGroupNames(w *RDTWorkLoad, m map[string]*resctrl.ResAssociation) (b, n string, s []string) {
	var new_grp string
	var base_grp string
	sub_grp := []string{}
	// no group specify
	if len(w.Group) == 0 {
		if len(w.TaskIDs) > 0 {
			// use the first task id as gname
			new_grp = w.TaskIDs[0]
		} else {
			// FIXME generate a better group name
			new_grp = w.ID
		}
		return ".", new_grp, []string{}
	}

	if len(w.Group) == 1 {
		_, ok := m[w.Group[0]]
		if ok {
			// find existed group
			// new group and base group are same
			return w.Group[0], w.Group[0], []string{}
		} else {
			// doesn't find one, create a new one
			return ".", w.Group[0], []string{}
		}
	}

	if len(w.Group) == 2 {
		_, ok1 := m[w.Group[0]]
		_, ok2 := m[w.Group[1]]
		if ok1 && ok2 {
			// FIXME error
			return "", "", []string{}
		}
		if !ok1 && !ok2 {
			// FIXME error
			return "", "", []string{}
		}

		if ok1 {
			base_grp = w.Group[0]
			new_grp = w.Group[1]
		} else {
			base_grp = w.Group[1]
			new_grp = w.Group[2]
		}
		for g, _ := range m {
			// sub group names like base-sub
			if strings.HasPrefix(g, base_grp+"-") {
				sub_grp = append(sub_grp, g)
			}
		}
		return base_grp, new_grp, sub_grp
	}
	// error
	return "", "", []string{}
}

// return a Resassociation with proper mask set
func createOrGetResAss(er *EnforceRequest) (t resctrl.ResAssociation, err error) {
	if er.BaseGrp == er.NewGrp {
		return *er.Resall[er.BaseGrp], nil
	}
	for _, sg := range er.SubGrps {
		if er.BaseGrp+"-"+er.NewGrp == sg {
			// new_grp has existed
			return *er.Resall[sg], nil
		}
	}
	// consider move consume checking to createNewResassociation
	if er.BaseGrp == "." {
		// sub_grp should be empty if the base group is "."
		// or that should be an internal error.
		er.Consume = true
		er.SubGrps = []string{}
	} else {
		er.Consume = false
	}
	return createNewResassociation(er)
}

// return a new Resassociation based on the given resctrl.ResAssociation
func createNewResassociation(er *EnforceRequest) (t resctrl.ResAssociation, err error) {
	baseRes := er.Resall[er.BaseGrp]
	if er.BaseGrp == "." {
		// if infra group are created, should be added it to ignore group.
		baseRes = calculateDefaultGroup(er.Resall, []string{"."}, false)
		er.Resall["."] = baseRes
	}

	rdtinfo := resctrl.GetRdtCosInfo()
	// loop for each level 3 cache to construct new resassociation
	newResAss := resctrl.ResAssociation{}
	newResAss.Schemata = make(map[string][]resctrl.CacheCos)

	for cattype, res := range baseRes.Schemata {
		// construct ResAssociation for each cache id
		catinfo := rdtinfo[strings.ToLower(cattype)]
		for i, _ := range res {
			var newcos resctrl.CacheCos
			// fill the new mask with cbm_mask
			if !inCacheList(uint32(i), er.Cache_IDs) {
				newcos = resctrl.CacheCos{Id: uint8(i), Mask: catinfo.CbmMask}
			} else {
				// compute sub_grp's offset for the i(th) 'cattype'
				offset := calculateOffset(er.Resall, er.SubGrps, cattype, uint32(i))
				bmbase, _ := libutil.NewBitmap(res[i].Mask)
				newbm := bmbase.GetConnectiveBits(er.MaxWays, offset, true)

				if newbm.IsEmpty() {
					return newResAss, fmt.Errorf("Not enough cache can be allocated")
				}

				if er.Consume {
					bmbase = bmbase.Xor(newbm)
				}

				tmpbm := bmbase.MaxConnectiveBits()
				res[i].Mask = tmpbm.ToString()
				newcos = resctrl.CacheCos{Id: uint8(i), Mask: newbm.ToString()}
			}

			newResAss.Schemata[cattype] = append(newResAss.Schemata[cattype], newcos)
			log.Debugf("Newly created Mask for Cache %d is %s", i, newcos.Mask)
			log.Debugf("Default Mask for Cache %d is %s", i, res[i].Mask)
		}
	}
	return newResAss, nil
}

// calculate a new default group which can be consumed by
// subtract bit mask in existed resource group
// e.g.
// r = ['.': [res1], 'group1': 'res2', 'infra': 'bla']
// ignore_grp = ['.', 'infra']
// calculateDefaultGroup will create a new resctrl.ResAssociation, the mask
// value is calculated by default mask subtract group1's schemata.
func calculateDefaultGroup(r map[string]*resctrl.ResAssociation, ignore_grp []string, consecutive bool) *resctrl.ResAssociation {

	defaultGrp := r["."]

	for _, v := range ignore_grp {
		delete(r, v)
	}

	// FIXME rdtinfo could be a global variable
	rdtinfo := resctrl.GetRdtCosInfo()
	newRes := new(resctrl.ResAssociation)
	newRes.Schemata = make(map[string][]resctrl.CacheCos)

	for t, schemata := range defaultGrp.Schemata {
		catinfo := rdtinfo[strings.ToLower(t)]
		newRes.Schemata[t] = make([]resctrl.CacheCos, 0, 10)
		for id, v := range schemata {
			bm, _ := libutil.NewBitmap(catinfo.CbmMask)
			// loop for all groups
			for _, g := range r {
				// ignore whole cbm
				if g.Schemata[t][v.Id].Mask == catinfo.CbmMask {
					continue
				}
				gbm, _ := libutil.NewBitmap(g.Schemata[t][v.Id].Mask)
				bm = bm.Xor(gbm)
			}

			var newcbm string

			if consecutive {
				tmpbm := bm.MaxConnectiveBits()
				newcbm = tmpbm.ToString()
			} else {
				newcbm = bm.ToString()
			}
			cacheCos := &resctrl.CacheCos{uint8(id), newcbm}
			newRes.Schemata[t] = append(newRes.Schemata[t], *cacheCos)

			log.Debugf("New default Mask for Cache %d is %s", cacheCos.Id, newcbm)
		}
	}

	return newRes
}

// Calculate offset for the pos'th cache of cattype based on sub_grp
// e.g.
// sub_grp = [base-sub1]
// base-sub1: L3:0=f;1=1
// calculateOffset(r, sub_grp, L3, 0) = 4
// calculateOffset(r, sub_grp, L3, 1) = 1
func calculateOffset(r map[string]*resctrl.ResAssociation, sub_grp []string, cattype string, pos uint32) uint32 {
	// len is used to avoid error prone. Though it is not so much time costing,
	// we really do not need to query it every time. It can be a global variable.
	// But it had better not to be sync.Once. It had better to be a singleton
	// that depens on who enable resctrl.
	// And we need a wrap for the NewBitmap with len. such as:
	// func NewCosBitmap(v string) {
	//     len := GetLenofCosOnce()
	//     return NewBitmap(len, v)
	// }
	bm0, _ := libutil.NewBitmap("")
	for _, g := range sub_grp {
		b, _ := libutil.NewBitmap(r[g].Schemata[cattype][pos].Mask)
		bm0 = bm0.Or(b)
	}
	if bm0.IsEmpty() {
		return 0
	} else {
		return bm0.Maximum()
	}
}

func getCacheIDs(cpubitmap string, cacheinfos *cache.CacheInfos, cpunum int) []uint32 {
	var CacheIDs []uint32
	cpubm, _ := libutil.NewBitmap(cpunum, cpubitmap)

	for _, c := range cacheinfos.Caches {
		// Okay, NewBitmap only support string list if we using human style
		bm, _ := libutil.NewBitmap(cpunum, strings.Split(c.ShareCpuList, "\n"))
		if !cpubm.And(bm).IsEmpty() {
			CacheIDs = append(CacheIDs, c.ID)
		}
	}
	return CacheIDs
}

func inCacheList(cache uint32, cache_list []uint32) bool {
	// TODO: if this case, workload has taskids.
	// Later we need to have abilitity to discover if has taskset
	// to pin this taskids on a cpuset or not, for now we allocate
	// cache on all cache.
	// FIXME: this shouldn't happen here actually
	if len(cache_list) == 0 {
		return true
	}

	for _, c := range cache_list {
		if cache == c {
			return true
		}
	}
	return false
}
