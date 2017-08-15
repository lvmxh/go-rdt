package workload

// workload api objects to represent resources in RDTAgent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
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
	"openstackcore-rdtagent/util"
	"openstackcore-rdtagent/util/rdtpool"
	. "openstackcore-rdtagent/util/rdtpool/base"
)

// FIXME this is not a global lock
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
	ID string `json:"id,omitempty"`
	// core ids, the work load run on top of cores/cpus
	CoreIDs []string `json:"core_ids"`
	// task ids, the work load's task ids
	TaskIDs []string `json:"task_ids"`
	// policy the workload want to apply
	Policy string `json:"policy"`
	// Status
	Status string
	// CosName
	CosName string
	// Max Cache ways, use pointer to distinguish 0 value and empty value
	MaxCache *uint32 `json:"max_cache,omitempty"`
	// Min Cache ways, use pointer to distinguish 0 value and empty value
	MinCache *uint32 `json:"min_cache,omitempty"`
}

// build this struct when create Resasscciation
type EnforceRequest struct {
	// all resassociations on the host
	Resall map[string]*resctrl.ResAssociation
	// max cache ways
	MaxWays uint32
	// min cache ways, not used yet
	MinWays uint32
	// enforce on which cache ids
	Cache_IDs []uint32
	// consume from base group or not
	Consume bool
	// request type
	Type string
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

	if w.Policy == "" {
		if w.MaxCache == nil || w.MinCache == nil {
			return fmt.Errorf("Need to provide max_cache and min_cache if no policy specified.")
		}
	}

	return nil
}

func (w *RDTWorkLoad) Enforce() *AppError {
	w.Status = Failed

	l.Lock()
	defer l.Unlock()
	resaall := resctrl.GetResAssociation()

	er := &EnforceRequest{}
	if err := populateEnforceRequest(er, w); err != nil {
		return err
	}

	targetLev := strconv.FormatUint(uint64(libcache.GetLLC()), 10)
	av, err := rdtpool.GetAvailableCacheSchemata(resaall, []string{"infra", "."}, er.Type, "L"+targetLev)
	if err != nil {
		return NewAppError(http.StatusInternalServerError,
			"Error to get available cache", err)
	}

	reserved := rdtpool.GetReservedInfo()

	candidate := make(map[string]*libutil.Bitmap, 0)
	for k, v := range av {
		cacheId, _ := strconv.Atoi(k)
		if !inCacheList(uint32(cacheId), er.Cache_IDs) {
			candidate[k], _ = libutil.NewBitmap(GetCosInfo().CbmMaskLen, GetCosInfo().CbmMask)
			continue
		}
		switch er.Type {
		case rdtpool.Guarantee:
			// TODO
			// candidate[k] = v.GetBestMatchConnectiveBits(er.MaxWays, 0, true)
			candidate[k] = v.GetConnectiveBits(er.MaxWays, 0, false)
		case rdtpool.Besteffort:
			candidate[k], _ = libutil.NewBitmap(GetCosInfo().CbmMaskLen, "")
			tmp := v.GetConnectiveBits(er.MinWays, 0, true)
			if !tmp.IsEmpty() {
				candidate[k] = v.Or(reserved[rdtpool.Shared].Schemata[k]).GetConnectiveBits(er.MaxWays, 0, true)

			} else {
				tmp = v.GetConnectiveBits(er.MinWays, 0, false)
				if !tmp.IsEmpty() {
					candidate[k] = v.Or(reserved[rdtpool.Shared].Schemata[k]).GetConnectiveBits(er.MaxWays, 0, false)
				}
			}
		case rdtpool.Shared:
			candidate[k] = reserved[rdtpool.Shared].Schemata[k]
		}

		if candidate[k].IsEmpty() {
			return AppErrorf(http.StatusBadRequest,
				"Not enough cache left on cache_id %s", k)
		}
	}

	resAss := newResAss(candidate, targetLev)
	fmt.Println(resAss)

	cpubitstr := ""
	if len(w.CoreIDs) >= 0 {
		bm, _ := CpuBitmaps(w.CoreIDs)
		cpubitstr = bm.ToString()
	}
	resAss.Tasks = append(resAss.Tasks, w.TaskIDs...)
	resAss.CPUs = cpubitstr

	var grpName string

	if len(w.TaskIDs) > 0 {
		grpName = w.TaskIDs[0] + "-" + er.Type
	} else if len(w.CoreIDs) > 0 {
		grpName = w.CoreIDs[0] + "-" + er.Type
	}

	if err = resAss.Commit(grpName); err != nil {
		log.Errorf("Error while try to commit resource group for workload %s, group name %s", w.ID, grpName)
		return NewAppError(http.StatusInternalServerError,
			"Error to commit resource group for workload.", err)
	}

	// reset os group
	if err = rdtpool.SetOSGroup(); err != nil {
		log.Errorf("Error while try to commit resource group for default group")
		resctrl.DestroyResAssociation(grpName)
		return NewAppError(http.StatusInternalServerError,
			"Error while try to commit resource group for default group.", err)
	}

	w.CosName = grpName
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

	r.Tasks = util.SubtractStringSlice(r.Tasks, w.TaskIDs)

	// safely remove resource group if no tasks and cpu bit map is empty
	if len(r.Tasks) < 1 {
		log.Printf("Remove resource group: %s", w.CosName)
		if err := resctrl.DestroyResAssociation(w.CosName); err != nil {
			return err
		}
		if err := rdtpool.SetOSGroup(); err != nil {
			return err
		}
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

// Patch a workload
func (w *RDTWorkLoad) Update(patched *RDTWorkLoad) (*RDTWorkLoad, *AppError) {

	// if we change policy/max_cache/min_cache, release current resource group
	// and re-enforce it.
	reEnforce := false
	if patched.MaxCache != nil {
		if w.MaxCache == nil {
			w.MaxCache = patched.MaxCache
			reEnforce = true
		}
		if w.MaxCache != nil && *w.MaxCache != *patched.MaxCache {
			*w.MaxCache = *patched.MaxCache
			reEnforce = true
		}
	}

	if patched.MinCache != nil {
		if w.MinCache == nil {
			w.MinCache = patched.MinCache
			reEnforce = true
		}
		if w.MinCache != nil && *w.MinCache != *patched.MinCache {
			*w.MinCache = *patched.MinCache
			reEnforce = true
		}
	}

	if patched.Policy != w.Policy {
		w.Policy = patched.Policy
		reEnforce = true
	}

	if reEnforce == true {
		if err := w.Release(); err != nil {
			return w, NewAppError(http.StatusInternalServerError, "Faild to release workload",
				fmt.Errorf(""))
		}

		if len(patched.TaskIDs) > 0 {
			w.TaskIDs = patched.TaskIDs
		}
		if len(patched.CoreIDs) > 0 {
			w.CoreIDs = patched.CoreIDs
		}
		return w, w.Enforce()
	}

	l.Lock()
	defer l.Unlock()
	resaall := resctrl.GetResAssociation()

	if !reflect.DeepEqual(patched.CoreIDs, w.CoreIDs) ||
		!reflect.DeepEqual(patched.TaskIDs, w.TaskIDs) {
		err := patched.Validate()
		if err != nil {
			return w, NewAppError(http.StatusBadRequest, "Failed to validate workload", err)
		}

		targetResAss, ok := resaall[w.CosName]
		if !ok {
			return w, NewAppError(http.StatusInternalServerError, "Can not find resource group name",
				fmt.Errorf(""))
		}

		if len(patched.TaskIDs) > 0 {
			// FIXME (Shaohe) Is this a bug? Seems the targetResAss.Tasks is inconsistent with w.TaskIDs
			targetResAss.Tasks = append(targetResAss.Tasks, patched.TaskIDs...)
			w.TaskIDs = patched.TaskIDs
		}
		if len(patched.CoreIDs) > 0 {
			bm, err := CpuBitmaps(patched.CoreIDs)
			if err != nil {
				return w, NewAppError(http.StatusBadRequest,
					"Failed to Pareser workload coreIDs.", err)
			}
			// TODO: check if this new CoreIDs overwrite other resource group
			targetResAss.CPUs = bm.ToString()
			w.CoreIDs = patched.CoreIDs
		}
		// commit changes
		if err = targetResAss.Commit(w.CosName); err != nil {
			log.Errorf("Error while try to commit resource group for workload %s, group name %s", w.ID, w.CosName)
			return w, NewAppError(http.StatusInternalServerError,
				"Error to commit resource group for workload.", err)
		}
	}
	return w, nil
}

// Calculate offset for the pos'th cache of cattype based on sub_grp
// e.g.
// sub_grp = [base-sub1]
// base-sub1: L3:0=f;1=1
// calculateOffset(r, sub_grp, L3, 0) = 4
// calculateOffset(r, sub_grp, L3, 1) = 1
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

func populateEnforceRequest(req *EnforceRequest, w *RDTWorkLoad) *AppError {

	w.Status = None
	cpubitstr := ""
	if len(w.CoreIDs) >= 0 {
		bm, err := CpuBitmaps(w.CoreIDs)
		if err != nil {
			return NewAppError(http.StatusBadRequest,
				"Failed to Parese workload coreIDs.", err)
		}
		cpubitstr = bm.ToString()
	}

	cacheinfo := &cache.CacheInfos{}
	cacheinfo.GetByLevel(libcache.GetLLC())

	cpunum := cpu.HostCpuNum()
	if cpunum == 0 {
		return AppErrorf(http.StatusInternalServerError,
			"Unable to get Total CPU numbers on Host")
	}

	req.Cache_IDs = getCacheIDs(cpubitstr, cacheinfo, cpunum)

	populatePolicy := true

	if w.MinCache != nil {
		req.MinWays = *w.MinCache
	}
	if w.MaxCache != nil {
		req.MaxWays = *w.MaxCache
		populatePolicy = false
	}
	// else get max/min from policy
	if populatePolicy {
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

		maxWays, err := strconv.Atoi(p["MaxCache"])
		if err != nil {
			return NewAppError(http.StatusInternalServerError,
				"Error define MaxCache in Polciy.", err)
		}
		req.MaxWays = uint32(maxWays)

		minWays, err := strconv.Atoi(p["MinCache"])
		if err != nil {
			return NewAppError(http.StatusInternalServerError,
				"Error define MinCache in Polciy.", err)
		}
		req.MinWays = uint32(minWays)
	}

	if req.MaxWays == 0 {
		req.Type = rdtpool.Shared
	} else if req.MaxWays > req.MinWays && req.MinWays != 0 {
		req.Type = rdtpool.Besteffort
	} else if req.MaxWays == req.MinWays {
		req.Type = rdtpool.Guarantee
	} else {
		return AppErrorf(http.StatusBadRequest,
			"Bad request, max_cache=%d, min_cache=%d", req.MaxWays, req.MinWays)
	}
	return nil
}

func newResAss(r map[string]*libutil.Bitmap, level string) *resctrl.ResAssociation {
	newResAss := resctrl.ResAssociation{}
	newResAss.Schemata = make(map[string][]resctrl.CacheCos)

	targetLev := "L" + level

	for k, v := range r {
		cacheId, _ := strconv.Atoi(k)
		newcos := resctrl.CacheCos{Id: uint8(cacheId), Mask: v.ToString()}
		newResAss.Schemata[targetLev] = append(newResAss.Schemata[targetLev], newcos)

		log.Debugf("Newly created Mask for Cache %s is %s", k, newcos.Mask)
	}
	return &newResAss
}
